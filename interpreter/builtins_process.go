//go:build !js

package interpreter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultProcessMaxOutputBytes int64 = 1 << 20

var errProcessOutputLimit = errors.New("process output exceeds maxOutputBytes")

type processRunOptions struct {
	command        string
	args           []string
	cwd            string
	env            map[string]string
	inheritEnv     bool
	stdin          *string
	timeoutMs      int64
	maxOutputBytes int64
}

func registerProcessBuiltins() {
	builtins["processRun"] = &Builtin{Name: "processRun", Fn: builtinProcessRun}
}

func builtinProcessRun(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "processRun expects options object"}
	}

	opts, err := parseProcessRunOptions(args[0])
	if err != nil {
		return nil, err
	}

	ctx, cancelProcess, cleanup := runtimeProcessRunContext(e, opts.timeoutMs)
	defer cleanup()

	cmd := exec.CommandContext(ctx, opts.command, opts.args...)

	if opts.cwd != "" {
		dir, err := resolveProcessRunDir(e, opts.cwd)
		if err != nil {
			return nil, recoverableError("process_spawn", "processRun spawn error: "+err.Error())
		}
		cmd.Dir = dir
	}

	env := processRunEnvSnapshot(e, opts)
	if env != nil {
		cmd.Env = env
	}

	if opts.stdin != nil {
		cmd.Stdin = strings.NewReader(*opts.stdin)
	}

	output := newProcessOutputCollector(opts.maxOutputBytes, cancelProcess)
	cmd.Stdout = output.stdoutWriter()
	cmd.Stderr = output.stderrWriter()

	startedAt := time.Now()
	runErr := cmd.Run()
	durationMs := time.Since(startedAt).Milliseconds()
	if durationMs < 0 {
		durationMs = 0
	}

	stdout, stderr := output.snapshot()

	if output.exceeded() {
		return nil, recoverableError(
			"process_output_limit",
			fmt.Sprintf("processRun output exceeded maxOutputBytes (%d)", opts.maxOutputBytes),
		)
	}

	if runErr != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, recoverableError("process_timeout", fmt.Sprintf("processRun timeout after %dms", opts.timeoutMs))
		}

		if canceled := processRunCanceledError(e, ctx); canceled != nil {
			return nil, canceled
		}

		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode, killed := processStateStatus(exitErr.ProcessState)
			return processRunResult(false, exitCode, stdout, stderr, durationMs, false, killed), nil
		}

		if cmd.ProcessState == nil {
			return nil, recoverableError("process_spawn", "processRun spawn error: "+runErr.Error())
		}

		return nil, recoverableError("process_io", "processRun error: "+runErr.Error())
	}

	exitCode, killed := processStateStatus(cmd.ProcessState)
	ok := exitCode == 0
	return processRunResult(ok, exitCode, stdout, stderr, durationMs, false, killed), nil
}

func parseProcessRunOptions(val Value) (processRunOptions, error) {
	pairs, ok := objectPairs(val)
	if !ok {
		return processRunOptions{}, &RuntimeError{Message: "processRun expects options object"}
	}

	opts := processRunOptions{
		inheritEnv:     true,
		maxOutputBytes: defaultProcessMaxOutputBytes,
	}

	commandVal, ok := pairs["command"]
	if !ok {
		return processRunOptions{}, &RuntimeError{Message: "processRun expects command"}
	}
	command, ok := stringArg(commandVal)
	if !ok {
		return processRunOptions{}, &RuntimeError{Message: "processRun command must be string"}
	}
	if strings.TrimSpace(command) == "" {
		return processRunOptions{}, &RuntimeError{Message: "processRun command must not be empty"}
	}
	opts.command = command

	if argsVal, ok := pairs["args"]; ok && !Equivalent(argsVal, NullValue) {
		arr, ok := argsVal.(*Array)
		if !ok {
			return processRunOptions{}, &RuntimeError{Message: "processRun args must be array"}
		}
		out := make([]string, 0, len(arr.Elements))
		for _, arg := range arr.Elements {
			s, ok := stringArg(arg)
			if !ok {
				return processRunOptions{}, &RuntimeError{Message: "processRun args must contain strings"}
			}
			out = append(out, s)
		}
		opts.args = out
	}

	if cwdVal, ok := pairs["cwd"]; ok && !Equivalent(cwdVal, NullValue) {
		cwd, ok := stringArg(cwdVal)
		if !ok {
			return processRunOptions{}, &RuntimeError{Message: "processRun cwd must be string"}
		}
		opts.cwd = cwd
	}

	if envVal, ok := pairs["env"]; ok && !Equivalent(envVal, NullValue) {
		env, err := parseProcessRunEnv(envVal)
		if err != nil {
			return processRunOptions{}, err
		}
		opts.env = env
	}

	if inheritVal, ok := pairs["inheritEnv"]; ok && !Equivalent(inheritVal, NullValue) {
		inherit, ok := inheritVal.(*Boolean)
		if !ok {
			return processRunOptions{}, &RuntimeError{Message: "processRun inheritEnv must be bool"}
		}
		opts.inheritEnv = inherit.Value
	}

	if stdinVal, ok := pairs["stdin"]; ok && !Equivalent(stdinVal, NullValue) {
		stdin, ok := stringArg(stdinVal)
		if !ok {
			return processRunOptions{}, &RuntimeError{Message: "processRun stdin must be string"}
		}
		opts.stdin = &stdin
	}

	if timeoutVal, ok := pairs["timeoutMs"]; ok && !Equivalent(timeoutVal, NullValue) {
		timeout, ok := timeoutVal.(*Integer)
		if !ok {
			return processRunOptions{}, &RuntimeError{Message: "processRun timeoutMs must be integer milliseconds"}
		}
		if timeout.Value < 0 {
			return processRunOptions{}, &RuntimeError{Message: "processRun timeoutMs must be >= 0"}
		}
		opts.timeoutMs = timeout.Value
	}

	if limitVal, ok := pairs["maxOutputBytes"]; ok && !Equivalent(limitVal, NullValue) {
		limit, ok := limitVal.(*Integer)
		if !ok {
			return processRunOptions{}, &RuntimeError{Message: "processRun maxOutputBytes must be integer"}
		}
		if limit.Value <= 0 {
			return processRunOptions{}, &RuntimeError{Message: "processRun maxOutputBytes must be > 0"}
		}
		opts.maxOutputBytes = limit.Value
	}

	return opts, nil
}

func parseProcessRunEnv(val Value) (map[string]string, error) {
	switch env := val.(type) {
	case *Object:
		return processRunEnvFromPairs(env.Pairs)
	case *ModuleObject:
		if env.Env == nil {
			return map[string]string{}, nil
		}
		return processRunEnvFromPairs(env.Env.Snapshot())
	case *Map:
		out := make(map[string]string, len(env.Pairs))
		for key, value := range env.Pairs {
			if key.Type != STRING && key.Type != CHAR {
				return nil, &RuntimeError{Message: "processRun env keys must be strings"}
			}
			s, ok := stringArg(value)
			if !ok {
				return nil, &RuntimeError{Message: "processRun env values must be strings"}
			}
			out[key.Value] = s
		}
		return out, nil
	default:
		return nil, &RuntimeError{Message: "processRun env must be object or map"}
	}
}

func processRunEnvFromPairs(pairs map[string]Value) (map[string]string, error) {
	out := make(map[string]string, len(pairs))
	for key, value := range pairs {
		s, ok := stringArg(value)
		if !ok {
			return nil, &RuntimeError{Message: "processRun env values must be strings"}
		}
		out[key] = s
	}
	return out, nil
}

func processRunEnvSnapshot(e *Evaluator, opts processRunOptions) []string {
	if !opts.inheritEnv {
		if len(opts.env) == 0 {
			return []string{}
		}
		return processRunSortedEnv(opts.env)
	}

	base := runtimeEnviron(e)
	if len(base) == 0 {
		base = os.Environ()
	}
	if len(opts.env) == 0 {
		return cloneStrings(base)
	}

	merged := makeEnvMap(base)
	for key, value := range opts.env {
		merged[key] = value
	}
	return processRunSortedEnv(merged)
}

func processRunSortedEnv(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key+"="+values[key])
	}
	return out
}

func resolveProcessRunDir(e *Evaluator, cwd string) (string, error) {
	if filepath.IsAbs(cwd) {
		return filepath.Clean(cwd), nil
	}

	root := ""
	if e != nil {
		root = e.projectRoot
	}
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	return filepath.Clean(filepath.Join(root, cwd)), nil
}

func processRunResult(ok bool, exitCode int64, stdout string, stderr string, durationMs int64, timedOut bool, killed bool) Value {
	return &Object{Pairs: map[string]Value{
		"ok":         &Boolean{Value: ok},
		"exitCode":   &Integer{Value: exitCode},
		"stdout":     &String{Value: stdout},
		"stderr":     &String{Value: stderr},
		"durationMs": &Integer{Value: durationMs},
		"timedOut":   &Boolean{Value: timedOut},
		"killed":     &Boolean{Value: killed},
	}}
}

func processStateStatus(state *os.ProcessState) (int64, bool) {
	if state == nil {
		return 0, false
	}
	code := int64(state.ExitCode())
	return code, code < 0
}

func runtimeProcessRunContext(e *Evaluator, timeoutMs int64) (context.Context, func(), func()) {
	base := context.Background()
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if timeoutMs > 0 {
		ctx, cancel = context.WithTimeout(base, time.Duration(timeoutMs)*time.Millisecond)
	} else {
		ctx, cancel = context.WithCancel(base)
	}

	var cancelOnce sync.Once
	stop := func() {
		cancelOnce.Do(cancel)
	}

	done := make(chan struct{})
	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			close(done)
			stop()
		})
	}

	cancelCh := runtimeCancelSignal(e)
	fatalCh := runtimeFatalSignal(e)
	if cancelCh != nil || fatalCh != nil {
		go func() {
			select {
			case <-cancelCh:
				stop()
			case <-fatalCh:
				stop()
			case <-done:
			}
		}()
	}

	return ctx, stop, cleanup
}

func processRunCanceledError(e *Evaluator, ctx context.Context) error {
	if ctx == nil || !errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

	cancelCh := runtimeCancelSignal(e)
	if cancelCh != nil {
		select {
		case <-cancelCh:
			return canceledError()
		default:
		}
	}

	fatalCh := runtimeFatalSignal(e)
	if fatalCh != nil {
		select {
		case <-fatalCh:
			return runtimeFatalError(e)
		default:
		}
	}

	return canceledError()
}

type processOutputCollector struct {
	mu sync.Mutex

	stdout bytes.Buffer
	stderr bytes.Buffer

	limit int64
	used  int64

	overflow bool
	cancel   func()
}

type processOutputStream int

const (
	processOutputStdout processOutputStream = iota
	processOutputStderr
)

type processOutputWriter struct {
	collector *processOutputCollector
	stream    processOutputStream
}

func newProcessOutputCollector(limit int64, cancel func()) *processOutputCollector {
	return &processOutputCollector{limit: limit, cancel: cancel}
}

func (c *processOutputCollector) stdoutWriter() *processOutputWriter {
	return &processOutputWriter{collector: c, stream: processOutputStdout}
}

func (c *processOutputCollector) stderrWriter() *processOutputWriter {
	return &processOutputWriter{collector: c, stream: processOutputStderr}
}

func (w *processOutputWriter) Write(p []byte) (int, error) {
	if w == nil || w.collector == nil {
		return len(p), nil
	}
	return w.collector.write(w.stream, p)
}

func (c *processOutputCollector) write(stream processOutputStream, p []byte) (int, error) {
	c.mu.Lock()
	if c.overflow {
		c.mu.Unlock()
		return 0, errProcessOutputLimit
	}

	remaining := c.limit - c.used
	if remaining <= 0 {
		c.overflow = true
		c.mu.Unlock()
		c.cancelIfNeeded()
		return 0, errProcessOutputLimit
	}

	written := len(p)
	overflowed := false
	if int64(written) > remaining {
		written = int(remaining)
		overflowed = true
	}

	if written > 0 {
		if stream == processOutputStdout {
			_, _ = c.stdout.Write(p[:written])
		} else {
			_, _ = c.stderr.Write(p[:written])
		}
		c.used += int64(written)
	}

	if overflowed {
		c.overflow = true
	}
	c.mu.Unlock()

	if overflowed {
		c.cancelIfNeeded()
		return written, errProcessOutputLimit
	}

	return len(p), nil
}

func (c *processOutputCollector) snapshot() (string, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stdout.String(), c.stderr.String()
}

func (c *processOutputCollector) exceeded() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.overflow
}

func (c *processOutputCollector) cancelIfNeeded() {
	if c == nil || c.cancel == nil {
		return
	}
	c.cancel()
}
