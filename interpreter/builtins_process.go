//go:build !js

package interpreter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	processModePipe    = "pipe"
	processModeInherit = "inherit"
	processModeNull    = "null"

	processOverflowTruncate = "truncate"
	processOverflowError    = "error"

	defaultProcessCaptureBytes int64 = 1048576
)

type processStageSpec struct {
	command    string
	args       []string
	cwd        string
	env        map[string]string
	inheritEnv bool
}

type Process struct {
	mu sync.Mutex

	cmds []*exec.Cmd

	cancel context.CancelFunc

	waitStart    sync.Once
	waitLoopFunc func()

	waitCh       chan processWaitResult
	completeOnce sync.Once

	done    bool
	running bool
	status  Value
	waitErr error

	startedAt time.Time

	aborted bool

	stdinMode  string
	stdoutMode string
	stderrMode string

	stdinType  string
	stdoutType string
	stderrType string

	stdinStream  *StreamWriter
	stdoutStream *StreamReader
	stderrStream *StreamReader

	ioErr error
}

func (p *Process) Type() ValueType { return PROCESS }
func (p *Process) Inspect() string { return "<process>" }

func (p *Process) markDone(status Value, err error) {
	if p == nil {
		return
	}
	p.completeOnce.Do(func() {
		p.mu.Lock()
		p.done = true
		p.running = false
		p.status = status
		p.waitErr = err
		p.mu.Unlock()
		p.waitCh <- processWaitResult{status: status, err: err}
	})
}

func (p *Process) ensureWaitLoop() {
	if p == nil {
		return
	}
	p.waitStart.Do(func() {
		if p.waitLoopFunc != nil {
			p.waitLoopFunc()
		}
	})
}

func (p *Process) snapshotResult() (processWaitResult, bool) {
	if p == nil {
		return processWaitResult{}, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.done {
		return processWaitResult{}, false
	}
	return processWaitResult{status: p.status, err: p.waitErr}, true
}

func (p *Process) setIOError(err error) {
	if p == nil || err == nil {
		return
	}
	p.mu.Lock()
	if p.ioErr == nil {
		p.ioErr = err
	}
	p.mu.Unlock()
}

func (p *Process) getIOError() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	err := p.ioErr
	p.mu.Unlock()
	return err
}

func (p *Process) PID() int64 {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, cmd := range p.cmds {
		if cmd != nil && cmd.Process != nil {
			return int64(cmd.Process.Pid)
		}
	}
	return 0
}

func (p *Process) Running() bool {
	if p == nil {
		return false
	}
	p.mu.Lock()
	running := p.running
	p.mu.Unlock()
	return running
}

func (p *Process) markAborted() {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.aborted = true
	p.mu.Unlock()
}

func (p *Process) abortedState() bool {
	if p == nil {
		return false
	}
	p.mu.Lock()
	aborted := p.aborted
	p.mu.Unlock()
	return aborted
}

func (p *Process) signalAll(sig os.Signal) error {
	if p == nil {
		return errors.New("process unavailable")
	}
	p.mu.Lock()
	cmds := append([]*exec.Cmd(nil), p.cmds...)
	done := p.done
	p.mu.Unlock()
	if done {
		return errors.New("process is not running")
	}

	var firstErr error
	for _, cmd := range cmds {
		if cmd == nil || cmd.Process == nil {
			continue
		}
		if err := cmd.Process.Signal(sig); err != nil && !errors.Is(err, os.ErrProcessDone) {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (p *Process) Abort() error {
	p.markAborted()
	if p.cancel != nil {
		defer p.cancel()
	}
	if sig, _, ok := signalFromName("SIGTERM"); ok {
		return p.signalAll(sig)
	}
	return p.Kill()
}

func (p *Process) Kill() error {
	p.markAborted()
	if p == nil {
		return errors.New("process unavailable")
	}
	p.mu.Lock()
	cmds := append([]*exec.Cmd(nil), p.cmds...)
	done := p.done
	p.mu.Unlock()
	if done {
		return errors.New("process is not running")
	}

	var firstErr error
	for _, cmd := range cmds {
		if cmd == nil || cmd.Process == nil {
			continue
		}
		if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if p.cancel != nil {
		p.cancel()
	}
	return firstErr
}

func (p *Process) Signal(name string) error {
	sig, _, ok := signalFromName(name)
	if !ok {
		return fmt.Errorf("unknown signal: %s", name)
	}
	return p.signalAll(sig)
}

func (p *Process) inputStream() (*StreamWriter, bool) {
	if p == nil {
		return nil, false
	}
	if p.stdinMode != processModePipe || p.stdinStream == nil {
		return nil, false
	}
	return p.stdinStream, true
}

func (p *Process) outputStream() (*StreamReader, bool) {
	if p == nil {
		return nil, false
	}
	if p.stdoutMode != processModePipe || p.stdoutStream == nil {
		return nil, false
	}
	return p.stdoutStream, true
}

func (p *Process) errorStream() (*StreamReader, bool) {
	if p == nil {
		return nil, false
	}
	if p.stderrMode != processModePipe || p.stderrStream == nil {
		return nil, false
	}
	return p.stderrStream, true
}

type processWaitResult struct {
	status Value
	err    error
}

type processSpec struct {
	stages []processStageSpec

	timeoutMs int64

	stdinMode  string
	stdoutMode string
	stderrMode string

	stdinType  string
	stdoutType string
	stderrType string

	stdinText *string

	maxOutputBytes int64
	overflow       string
}

func registerProcessBuiltins() {
	builtins["proc"] = &Builtin{Name: "proc", Fn: builtinProc}
	builtins["run"] = &Builtin{Name: "run", Fn: builtinRun}
}

func builtinProc(e *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, &RuntimeError{Message: "proc expects (spec, opts?) or (command, ...args, opts?)"}
	}
	spec, err := parseProcSpec(args)
	if err != nil {
		return nil, err
	}
	process, err := startProcess(e, spec)
	if err != nil {
		return nil, err
	}
	return process, nil
}

func builtinRun(e *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, &RuntimeError{Message: "run expects (spec, opts?) or (command, ...args, opts?)"}
	}
	spec, err := parseRunSpec(args)
	if err != nil {
		return nil, err
	}
	return executeRunSpec(e, spec)
}

func executeRunSpec(e *Evaluator, spec processSpec) (Value, error) {

	process, err := startProcess(e, spec)
	if err != nil {
		return nil, err
	}

	stdout, ok := process.outputStream()
	if !ok {
		return nil, recoverableError("process_state", "run capture unavailable: stdout is not piped")
	}
	stderr, ok := process.errorStream()
	if !ok {
		return nil, recoverableError("process_state", "run capture unavailable: stderr is not piped")
	}

	type captureResult struct {
		value     string
		truncated bool
	}
	outCh := make(chan captureResult, 1)
	errCh := make(chan captureResult, 1)

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				reportRuntimePanic(e.runtime, "run stdout capture", recovered)
				outCh <- captureResult{}
			}
		}()
		value, truncated := collectProcessStream(stdout, spec.maxOutputBytes, spec.overflow)
		outCh <- captureResult{value: value, truncated: truncated}
	}()
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				reportRuntimePanic(e.runtime, "run stderr capture", recovered)
				errCh <- captureResult{}
			}
		}()
		value, truncated := collectProcessStream(stderr, spec.maxOutputBytes, spec.overflow)
		errCh <- captureResult{value: value, truncated: truncated}
	}()

	var cancelCh <-chan struct{}
	if e != nil && e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}
	status, _, err := processAwaitWithCancel(process, cancelCh, e.runtime)
	if err != nil {
		return nil, err
	}

	out := <-outCh
	errRes := <-errCh

	if spec.overflow == processOverflowError && (out.truncated || errRes.truncated) {
		return nil, recoverableError("process_output_limit", fmt.Sprintf("run output exceeded maxOutputBytes (%d)", spec.maxOutputBytes))
	}

	statusObj, ok := status.(*Object)
	if !ok {
		return nil, &RuntimeError{Message: "invalid run status"}
	}
	statusObj.Pairs["output"] = &String{Value: out.value}
	statusObj.Pairs["error"] = &String{Value: errRes.value}
	statusObj.Pairs["outputTruncated"] = &Boolean{Value: out.truncated}
	statusObj.Pairs["errorTruncated"] = &Boolean{Value: errRes.truncated}
	return statusObj, nil
}

func parseProcSpec(args []Value) (processSpec, error) {
	spec := processSpec{
		stdinMode:  processModeInherit,
		stdoutMode: processModeInherit,
		stderrMode: processModeInherit,
		stdinType:  streamTypeBytes,
		stdoutType: streamTypeBytes,
		stderrType: streamTypeBytes,
	}
	if len(args) == 0 {
		return processSpec{}, &RuntimeError{Message: "proc expects at least one argument"}
	}

	if _, isObject := objectPairs(args[0]); isObject {
		if len(args) > 2 {
			return processSpec{}, &RuntimeError{Message: "proc expects (spec, opts?)"}
		}
		stage, err := parseProcessStageValue(args[0])
		if err != nil {
			return processSpec{}, err
		}
		spec.stages = []processStageSpec{stage}
		var opts Value = NullValue
		if len(args) == 2 {
			opts = args[1]
		}
		return parseProcOptions(spec, opts)
	}

	command, ok := stringArg(args[0])
	if !ok {
		return processSpec{}, &RuntimeError{Message: "proc expects spec object or command string"}
	}
	end := len(args)
	var opts Value = NullValue
	if end > 1 {
		if pairs, ok := objectPairs(args[end-1]); ok && isProcessOptionObject(pairs, []string{
			"timeoutMs", "stdin", "stdout", "stderr", "stdinType", "stdoutType", "stderrType",
		}) {
			opts = args[end-1]
			end--
		}
	}
	stageArgs := make([]string, 0, max(0, end-1))
	for i := 1; i < end; i++ {
		arg, ok := stringArg(args[i])
		if !ok {
			return processSpec{}, &RuntimeError{Message: "proc variadic command args must be strings"}
		}
		stageArgs = append(stageArgs, arg)
	}
	spec.stages = []processStageSpec{{
		command:    command,
		args:       stageArgs,
		inheritEnv: true,
	}}
	return parseProcOptions(spec, opts)
}

func parseProcOptions(spec processSpec, opts Value) (processSpec, error) {
	if Equivalent(opts, NullValue) {
		return spec, nil
	}
	pairs, ok := objectPairs(opts)
	if !ok {
		return processSpec{}, &RuntimeError{Message: "proc options must be object"}
	}
	if err := rejectUnknownObjectKeys(pairs, "proc options", []string{
		"timeoutMs",
		"stdin",
		"stdout",
		"stderr",
		"stdinType",
		"stdoutType",
		"stderrType",
	}); err != nil {
		return processSpec{}, err
	}

	if timeoutVal, ok := pairs["timeoutMs"]; ok && !Equivalent(timeoutVal, NullValue) {
		timeout, ok := timeoutVal.(*Integer)
		if !ok {
			return processSpec{}, &RuntimeError{Message: "process timeoutMs must be integer milliseconds"}
		}
		if timeout.Value < 0 {
			return processSpec{}, &RuntimeError{Message: "process timeoutMs must be >= 0"}
		}
		spec.timeoutMs = timeout.Value
	}
	if modeVal, ok := pairs["stdin"]; ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stdin")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdinMode = mode
	}
	if modeVal, ok := pairs["stdout"]; ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stdout")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdoutMode = mode
	}
	if modeVal, ok := pairs["stderr"]; ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stderr")
		if err != nil {
			return processSpec{}, err
		}
		spec.stderrMode = mode
	}
	if typeVal, ok := pairs["stdinType"]; ok && !Equivalent(typeVal, NullValue) {
		mode, err := parseStreamType(typeVal, "stdinType")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdinType = mode
	}
	if typeVal, ok := pairs["stdoutType"]; ok && !Equivalent(typeVal, NullValue) {
		mode, err := parseStreamType(typeVal, "stdoutType")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdoutType = mode
	}
	if typeVal, ok := pairs["stderrType"]; ok && !Equivalent(typeVal, NullValue) {
		mode, err := parseStreamType(typeVal, "stderrType")
		if err != nil {
			return processSpec{}, err
		}
		spec.stderrType = mode
	}
	return spec, nil
}

func parseRunSpec(args []Value) (processSpec, error) {
	spec := processSpec{
		stdinMode:      processModeNull,
		stdoutMode:     processModePipe,
		stderrMode:     processModePipe,
		stdinType:      streamTypeText,
		stdoutType:     streamTypeText,
		stderrType:     streamTypeText,
		maxOutputBytes: defaultProcessCaptureBytes,
		overflow:       processOverflowTruncate,
	}
	if len(args) == 0 {
		return processSpec{}, &RuntimeError{Message: "run expects at least one argument"}
	}

	if _, isObject := objectPairs(args[0]); isObject {
		if len(args) > 2 {
			return processSpec{}, &RuntimeError{Message: "run expects (spec, opts?)"}
		}
		stage, err := parseProcessStageValue(args[0])
		if err != nil {
			return processSpec{}, err
		}
		spec.stages = []processStageSpec{stage}
		var opts Value = NullValue
		if len(args) == 2 {
			opts = args[1]
		}
		return parseRunOptions(spec, opts)
	}

	command, ok := stringArg(args[0])
	if !ok {
		return processSpec{}, &RuntimeError{Message: "run expects spec object or command string"}
	}
	end := len(args)
	var opts Value = NullValue
	if end > 1 {
		if pairs, ok := objectPairs(args[end-1]); ok && isProcessOptionObject(pairs, []string{
			"stdin", "timeoutMs", "maxOutputBytes", "overflow",
		}) {
			opts = args[end-1]
			end--
		}
	}
	stageArgs := make([]string, 0, max(0, end-1))
	for i := 1; i < end; i++ {
		arg, ok := stringArg(args[i])
		if !ok {
			return processSpec{}, &RuntimeError{Message: "run variadic command args must be strings"}
		}
		stageArgs = append(stageArgs, arg)
	}
	spec.stages = []processStageSpec{{
		command:    command,
		args:       stageArgs,
		inheritEnv: true,
	}}
	return parseRunOptions(spec, opts)
}

func parseRunOptions(spec processSpec, opts Value) (processSpec, error) {
	if Equivalent(opts, NullValue) {
		return spec, nil
	}
	pairs, ok := objectPairs(opts)
	if !ok {
		return processSpec{}, &RuntimeError{Message: "run options must be object"}
	}
	if err := rejectUnknownObjectKeys(pairs, "run options", []string{
		"stdin",
		"timeoutMs",
		"maxOutputBytes",
		"overflow",
	}); err != nil {
		return processSpec{}, err
	}

	if timeoutVal, ok := pairs["timeoutMs"]; ok && !Equivalent(timeoutVal, NullValue) {
		timeout, ok := timeoutVal.(*Integer)
		if !ok {
			return processSpec{}, &RuntimeError{Message: "process timeoutMs must be integer milliseconds"}
		}
		if timeout.Value < 0 {
			return processSpec{}, &RuntimeError{Message: "process timeoutMs must be >= 0"}
		}
		spec.timeoutMs = timeout.Value
	}
	if stdinVal, ok := pairs["stdin"]; ok && !Equivalent(stdinVal, NullValue) {
		stdin, ok := stringArg(stdinVal)
		if !ok {
			return processSpec{}, &RuntimeError{Message: "run stdin must be string"}
		}
		spec.stdinText = &stdin
		spec.stdinMode = processModePipe
	}
	if limitVal, ok := pairs["maxOutputBytes"]; ok && !Equivalent(limitVal, NullValue) {
		limit, ok := limitVal.(*Integer)
		if !ok {
			return processSpec{}, &RuntimeError{Message: "run maxOutputBytes must be integer"}
		}
		if limit.Value <= 0 {
			return processSpec{}, &RuntimeError{Message: "run maxOutputBytes must be > 0"}
		}
		spec.maxOutputBytes = limit.Value
	}
	if overflowVal, ok := pairs["overflow"]; ok && !Equivalent(overflowVal, NullValue) {
		overflow, ok := stringArg(overflowVal)
		if !ok {
			return processSpec{}, &RuntimeError{Message: "run overflow must be string"}
		}
		overflow = strings.TrimSpace(strings.ToLower(overflow))
		if overflow != processOverflowTruncate && overflow != processOverflowError {
			return processSpec{}, &RuntimeError{Message: "run overflow must be \"truncate\" or \"error\""}
		}
		spec.overflow = overflow
	}
	return spec, nil
}

func isProcessOptionObject(pairs map[string]Value, allowed []string) bool {
	if len(pairs) == 0 {
		return true
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range pairs {
		if _, ok := allowedSet[key]; !ok {
			return false
		}
	}
	return true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseProcessStageValue(val Value) (processStageSpec, error) {
	pairs, ok := objectPairs(val)
	if !ok {
		return processStageSpec{}, &RuntimeError{Message: "process spec must be object"}
	}
	return parseProcessStageFromPairs(pairs)
}

func parseProcessStageFromPairs(pairs map[string]Value) (processStageSpec, error) {
	if err := rejectUnknownObjectKeys(pairs, "process spec", []string{
		"command",
		"args",
		"cwd",
		"env",
		"inheritEnv",
	}); err != nil {
		return processStageSpec{}, err
	}

	commandVal, ok := pairs["command"]
	if !ok {
		return processStageSpec{}, &RuntimeError{Message: "process spec expects command"}
	}
	command, ok := stringArg(commandVal)
	if !ok {
		return processStageSpec{}, &RuntimeError{Message: "process command must be string"}
	}
	if strings.TrimSpace(command) == "" {
		return processStageSpec{}, &RuntimeError{Message: "process command must not be empty"}
	}
	stage := processStageSpec{command: command, inheritEnv: true}

	if argsVal, ok := pairs["args"]; ok && !Equivalent(argsVal, NullValue) {
		args, err := parseProcessArgs(argsVal, "process args")
		if err != nil {
			return processStageSpec{}, err
		}
		stage.args = args
	}
	if cwdVal, ok := pairs["cwd"]; ok && !Equivalent(cwdVal, NullValue) {
		cwd, ok := stringArg(cwdVal)
		if !ok {
			return processStageSpec{}, &RuntimeError{Message: "process cwd must be string"}
		}
		stage.cwd = cwd
	}
	if envVal, ok := pairs["env"]; ok && !Equivalent(envVal, NullValue) {
		env, err := parseProcessEnv(envVal)
		if err != nil {
			return processStageSpec{}, err
		}
		stage.env = env
	}
	if inheritVal, ok := pairs["inheritEnv"]; ok && !Equivalent(inheritVal, NullValue) {
		inherit, ok := inheritVal.(*Boolean)
		if !ok {
			return processStageSpec{}, &RuntimeError{Message: "process inheritEnv must be bool"}
		}
		stage.inheritEnv = inherit.Value
	}
	return stage, nil
}

func parseProcessArgs(val Value, label string) ([]string, error) {
	arr, ok := val.(*Array)
	if !ok {
		return nil, &RuntimeError{Message: label + " must be array"}
	}
	out := make([]string, 0, len(arr.Elements))
	for _, arg := range arr.Elements {
		s, ok := stringArg(arg)
		if !ok {
			return nil, &RuntimeError{Message: label + " must contain strings"}
		}
		out = append(out, s)
	}
	return out, nil
}

func parseProcessEnv(val Value) (map[string]string, error) {
	switch env := val.(type) {
	case *Object:
		return processEnvFromPairs(env.Pairs)
	case *ModuleObject:
		if env.Env == nil {
			return map[string]string{}, nil
		}
		return processEnvFromPairs(env.Env.Snapshot())
	case *Map:
		out := make(map[string]string, len(env.Pairs))
		for key, value := range env.Pairs {
			if key.Type != STRING && key.Type != CHAR {
				return nil, &RuntimeError{Message: "process env keys must be strings"}
			}
			s, ok := stringArg(value)
			if !ok {
				return nil, &RuntimeError{Message: "process env values must be strings"}
			}
			out[key.Value] = s
		}
		return out, nil
	default:
		return nil, &RuntimeError{Message: "process env must be object or map"}
	}
}

func processEnvFromPairs(pairs map[string]Value) (map[string]string, error) {
	out := make(map[string]string, len(pairs))
	for key, value := range pairs {
		s, ok := stringArg(value)
		if !ok {
			return nil, &RuntimeError{Message: "process env values must be strings"}
		}
		out[key] = s
	}
	return out, nil
}

func parseProcessMode(val Value, field string) (string, error) {
	mode, ok := stringArg(val)
	if !ok {
		return "", &RuntimeError{Message: "process " + field + " must be string"}
	}
	mode = strings.TrimSpace(strings.ToLower(mode))
	switch mode {
	case processModePipe, processModeInherit, processModeNull:
		return mode, nil
	default:
		return "", &RuntimeError{Message: "process " + field + " must be \"pipe\", \"inherit\", or \"null\""}
	}
}

func processAwaitWithCancel(p *Process, cancelCh <-chan struct{}, runtime *runtimeState) (Value, *Signal, error) {
	if p == nil {
		return nil, nil, &RuntimeError{Message: "wait expects task or process"}
	}
	if snapshot, ok := p.snapshotResult(); ok {
		if snapshot.err != nil {
			return nil, nil, snapshot.err
		}
		return snapshot.status, nil, nil
	}
	p.ensureWaitLoop()

	fatalCh := runtime.fatalSignal()
	var out processWaitResult
	probe := time.NewTicker(runtimeBlockProbeInterval)
	defer probe.Stop()
	for {
		select {
		case out = <-p.waitCh:
			goto done
		case <-cancelCh:
			return nil, nil, canceledError()
		case <-fatalCh:
			if err := runtime.getFatalTaskFailure(); err != nil {
				return nil, nil, err
			}
			return nil, nil, &RuntimeError{Message: "runtime terminated"}
		case <-probe.C:
		}
	}

done:
	if out.err != nil {
		return nil, nil, out.err
	}
	return out.status, nil, nil
}

func startProcess(e *Evaluator, spec processSpec) (*Process, error) {
	if len(spec.stages) == 0 {
		return nil, &RuntimeError{Message: "process requires at least one stage"}
	}

	ctx, stop, cleanup := runtimeProcessContext(e, spec.timeoutMs)

	cmds := make([]*exec.Cmd, 0, len(spec.stages))
	for _, stage := range spec.stages {
		cmd := exec.CommandContext(ctx, stage.command, stage.args...)
		if stage.cwd != "" {
			dir, err := resolveProcessDir(e, stage.cwd)
			if err != nil {
				cleanup()
				return nil, recoverableError("process_spawn", "process spawn error: "+err.Error())
			}
			cmd.Dir = dir
		}
		env := processStageEnvSnapshot(e, stage)
		if env != nil {
			cmd.Env = env
		}
		cmds = append(cmds, cmd)
	}

	process := &Process{
		cmds:       cmds,
		cancel:     stop,
		waitCh:     make(chan processWaitResult, 1),
		running:    true,
		startedAt:  time.Now(),
		stdinMode:  spec.stdinMode,
		stdoutMode: spec.stdoutMode,
		stderrMode: spec.stderrMode,
		stdinType:  spec.stdinType,
		stdoutType: spec.stdoutType,
		stderrType: spec.stderrType,
	}

	last := len(cmds) - 1
	for i := 0; i < last; i++ {
		reader, err := cmds[i].StdoutPipe()
		if err != nil {
			cleanup()
			return nil, recoverableError("process_spawn", "process spawn error: "+err.Error())
		}
		cmds[i+1].Stdin = reader
	}

	var stdinWriter io.WriteCloser
	switch spec.stdinMode {
	case processModePipe:
		if spec.stdinText != nil {
			cmds[0].Stdin = strings.NewReader(*spec.stdinText)
		} else {
			writer, err := cmds[0].StdinPipe()
			if err != nil {
				cleanup()
				return nil, recoverableError("process_spawn", "process spawn error: "+err.Error())
			}
			stdinWriter = writer
		}
	case processModeInherit:
		cmds[0].Stdin = os.Stdin
	case processModeNull:
		cmds[0].Stdin = nil
	}

	var stdoutReader io.ReadCloser
	switch spec.stdoutMode {
	case processModePipe:
		reader, err := cmds[last].StdoutPipe()
		if err != nil {
			cleanup()
			return nil, recoverableError("process_spawn", "process spawn error: "+err.Error())
		}
		stdoutReader = reader
	case processModeInherit:
		cmds[last].Stdout = os.Stdout
	case processModeNull:
		cmds[last].Stdout = io.Discard
	}

	stderrReaders := make([]io.ReadCloser, 0, len(cmds))
	switch spec.stderrMode {
	case processModePipe:
		for _, cmd := range cmds {
			reader, err := cmd.StderrPipe()
			if err != nil {
				cleanup()
				return nil, recoverableError("process_spawn", "process spawn error: "+err.Error())
			}
			stderrReaders = append(stderrReaders, reader)
		}
	case processModeInherit:
		for _, cmd := range cmds {
			cmd.Stderr = os.Stderr
		}
	case processModeNull:
		for _, cmd := range cmds {
			cmd.Stderr = io.Discard
		}
	}

	started := make([]*exec.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			for _, running := range started {
				if running != nil && running.Process != nil {
					_ = running.Process.Kill()
				}
			}
			cleanup()
			return nil, recoverableError("process_spawn", "process spawn error: "+err.Error())
		}
		started = append(started, cmd)
	}

	if stdinWriter != nil {
		process.stdinStream = &StreamWriter{
			writer: stdinWriter,
			closer: stdinWriter,
			mode:   spec.stdinType,
		}
	}
	if stdoutReader != nil {
		process.stdoutStream = &StreamReader{
			reader: stdoutReader,
			closer: stdoutReader,
			mode:   spec.stdoutType,
		}
	}
	if len(stderrReaders) > 0 {
		var runtime *runtimeState
		if e != nil {
			runtime = e.runtime
		}
		merged := processMergeReaders(stderrReaders, runtime)
		process.stderrStream = &StreamReader{
			reader: merged,
			closer: merged,
			mode:   spec.stderrType,
		}
	}

	process.waitLoopFunc = func() {
		var runtime *runtimeState
		if e != nil {
			runtime = e.runtime
		}
		runGuarded(runtime, "process wait", func() {
			processWaitLoop(process, cmds, ctx, last, cleanup)
		})
	}

	return process, nil
}

func processWaitLoop(process *Process, cmds []*exec.Cmd, ctx context.Context, last int, cleanup func()) {
	defer cleanup()
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) && !errors.Is(err, context.Canceled) {
				process.setIOError(err)
			}
		}
	}

	durationMs := time.Since(process.startedAt).Milliseconds()
	if durationMs < 0 {
		durationMs = 0
	}

	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
	aborted := process.abortedState()
	ioErr := process.getIOError()

	code, signalName := processExitState(cmds[last])
	ok := code == 0 && signalName == "" && !timedOut && !aborted
	status := processStatusValue(ok, code, signalName, timedOut, aborted, durationMs)

	if ioErr != nil {
		process.markDone(nil, recoverableError("process_io", "process I/O error: "+ioErr.Error()))
		return
	}
	process.markDone(status, nil)
}

func processExitState(cmd *exec.Cmd) (int64, string) {
	if cmd == nil || cmd.ProcessState == nil {
		return -1, ""
	}
	code := int64(cmd.ProcessState.ExitCode())
	signalName := ""
	if sys := cmd.ProcessState.Sys(); sys != nil {
		if ws, ok := sys.(syscall.WaitStatus); ok && ws.Signaled() {
			signalName = ws.Signal().String()
		}
	}
	return code, signalName
}

func processStatusValue(ok bool, code int64, signal string, timedOut bool, aborted bool, durationMs int64) Value {
	var signalValue Value = NullValue
	if signal != "" {
		signalValue = &String{Value: strings.ToUpper(signal)}
	}
	return &Object{Pairs: map[string]Value{
		"ok":         &Boolean{Value: ok},
		"code":       &Integer{Value: code},
		"signal":     signalValue,
		"timedOut":   &Boolean{Value: timedOut},
		"aborted":    &Boolean{Value: aborted},
		"durationMs": &Integer{Value: durationMs},
	}}
}

func processMergeReaders(readers []io.ReadCloser, runtime *runtimeState) io.ReadCloser {
	pipeReader, pipeWriter := io.Pipe()

	var wg sync.WaitGroup
	for _, reader := range readers {
		r := reader
		wg.Add(1)
		runGuarded(runtime, "process stderr merge copy", func() {
			defer wg.Done()
			defer func() { _ = r.Close() }()
			_, err := io.Copy(pipeWriter, r)
			if err != nil && !streamReadEnded(err) {
				_ = pipeWriter.CloseWithError(err)
			}
		})
	}

	runGuarded(runtime, "process stderr merge finalize", func() {
		wg.Wait()
		_ = pipeWriter.Close()
	})
	return pipeReader
}

func collectProcessStream(stream *StreamReader, limit int64, overflow string) (string, bool) {
	if stream == nil {
		return "", false
	}
	var builder strings.Builder
	truncated := false
	for {
		chunk, eof, err := stream.ReadChunk(defaultStreamReadSize)
		if err != nil {
			break
		}
		if eof {
			break
		}
		s := string(chunk)
		if limit <= 0 {
			truncated = true
			continue
		}
		remaining := limit - int64(builder.Len())
		if remaining <= 0 {
			truncated = true
			continue
		}
		if int64(len(s)) <= remaining {
			builder.WriteString(s)
			continue
		}
		builder.WriteString(s[:remaining])
		truncated = true
		if overflow == processOverflowError {
			// Keep draining the channel to avoid blocking producer,
			// but do not retain additional bytes.
			continue
		}
	}
	return builder.String(), truncated
}

func processStageEnvSnapshot(e *Evaluator, stage processStageSpec) []string {
	if !stage.inheritEnv {
		if len(stage.env) == 0 {
			return []string{}
		}
		return processSortedEnv(stage.env)
	}

	base := runtimeEnviron(e)
	if len(base) == 0 {
		base = os.Environ()
	}
	if len(stage.env) == 0 {
		return cloneStrings(base)
	}

	merged := makeEnvMap(base)
	for key, value := range stage.env {
		merged[key] = value
	}
	return processSortedEnv(merged)
}

func processSortedEnv(values map[string]string) []string {
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

func resolveProcessDir(e *Evaluator, cwd string) (string, error) {
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

func runtimeProcessContext(e *Evaluator, timeoutMs int64) (context.Context, context.CancelFunc, func()) {
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

	done := make(chan struct{})
	cancelCh := runtimeCancelSignal(e)
	fatalCh := runtimeFatalSignal(e)
	if cancelCh != nil || fatalCh != nil {
		var runtime *runtimeState
		if e != nil {
			runtime = e.runtime
		}
		runGuarded(runtime, "process cancel watcher", func() {
			select {
			case <-cancelCh:
				cancel()
			case <-fatalCh:
				cancel()
			case <-done:
			}
		})
	}

	cleanup := func() {
		select {
		case <-done:
		default:
			close(done)
		}
		cancel()
	}
	return ctx, cancel, cleanup
}
