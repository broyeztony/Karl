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

type ProcessCommand struct {
	Stage processStageSpec
}

func (c *ProcessCommand) Type() ValueType { return CMD }
func (c *ProcessCommand) Inspect() string { return "<cmd>" }

type ProcessPipeline struct {
	Stages []processStageSpec
}

func (p *ProcessPipeline) Type() ValueType { return PIPELINE }
func (p *ProcessPipeline) Inspect() string { return "<pipeline>" }

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
	builtins["cmd"] = &Builtin{Name: "cmd", Fn: builtinCmd}
	builtins["proc"] = &Builtin{Name: "proc", Fn: builtinProc}
	builtins["run"] = &Builtin{Name: "run", Fn: builtinRun}
}

func builtinCmd(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "cmd expects 1 object argument"}
	}
	pairs, ok := objectPairs(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "cmd expects object spec"}
	}
	stage, err := parseProcessStageFromPairs(pairs)
	if err != nil {
		return nil, err
	}
	return &ProcessCommand{Stage: stage}, nil
}

func builtinProc(e *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "proc expects (cmdOrPipeline, opts?)"}
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
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "run expects (cmdOrPipeline, opts?)"}
	}
	spec, err := parseRunSpec(args)
	if err != nil {
		return nil, err
	}

	process, err := startProcess(e, spec)
	if err != nil {
		return nil, err
	}

	stdout, ok := process.outputStream()
	if !ok {
		return nil, recoverableError("process_state", "run capture unavailable: stdOut is not piped")
	}
	stderr, ok := process.errorStream()
	if !ok {
		return nil, recoverableError("process_state", "run capture unavailable: stdErr is not piped")
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
	stages, err := parseProcessPlanStages(args[0])
	if err != nil {
		return processSpec{}, &RuntimeError{Message: "proc expects cmd or pipeline as first argument"}
	}
	spec.stages = stages

	if len(args) == 1 || Equivalent(args[1], NullValue) {
		return spec, nil
	}
	pairs, ok := objectPairs(args[1])
	if !ok {
		return processSpec{}, &RuntimeError{Message: "proc options must be object"}
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
	if modeVal, ok := firstDefinedValue(pairs, "stdIn", "stdin"); ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stdin")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdinMode = mode
	}
	if modeVal, ok := firstDefinedValue(pairs, "stdOut", "stdout"); ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stdout")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdoutMode = mode
	}
	if modeVal, ok := firstDefinedValue(pairs, "stdErr", "stderr"); ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stderr")
		if err != nil {
			return processSpec{}, err
		}
		spec.stderrMode = mode
	}
	if typeVal, ok := firstDefinedValue(pairs, "stdinType", "stdInType"); ok && !Equivalent(typeVal, NullValue) {
		mode, err := parseStreamType(typeVal, "stdinType")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdinType = mode
	}
	if typeVal, ok := firstDefinedValue(pairs, "stdoutType", "stdOutType"); ok && !Equivalent(typeVal, NullValue) {
		mode, err := parseStreamType(typeVal, "stdoutType")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdoutType = mode
	}
	if typeVal, ok := firstDefinedValue(pairs, "stderrType", "stdErrType"); ok && !Equivalent(typeVal, NullValue) {
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
	stages, err := parseProcessPlanStages(args[0])
	if err != nil {
		return processSpec{}, &RuntimeError{Message: "run expects cmd or pipeline as first argument"}
	}
	spec.stages = stages

	if len(args) == 1 || Equivalent(args[1], NullValue) {
		return spec, nil
	}
	pairs, ok := objectPairs(args[1])
	if !ok {
		return processSpec{}, &RuntimeError{Message: "run options must be object"}
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

func parseProcessPlanStages(val Value) ([]processStageSpec, error) {
	switch v := val.(type) {
	case *ProcessCommand:
		return []processStageSpec{v.Stage}, nil
	case *ProcessPipeline:
		return append([]processStageSpec(nil), v.Stages...), nil
	default:
		return nil, &RuntimeError{Message: "expects cmd or pipeline"}
	}
}

func parseProcessStageFromPairs(pairs map[string]Value) (processStageSpec, error) {
	commandVal, ok := pairs["command"]
	if !ok {
		return processStageSpec{}, &RuntimeError{Message: "cmd spec expects command"}
	}
	command, ok := stringArg(commandVal)
	if !ok {
		return processStageSpec{}, &RuntimeError{Message: "cmd command must be string"}
	}
	if strings.TrimSpace(command) == "" {
		return processStageSpec{}, &RuntimeError{Message: "cmd command must not be empty"}
	}
	stage := processStageSpec{command: command, inheritEnv: true}

	if argsVal, ok := pairs["args"]; ok && !Equivalent(argsVal, NullValue) {
		args, err := parseProcessArgs(argsVal, "cmd args")
		if err != nil {
			return processStageSpec{}, err
		}
		stage.args = args
	}
	if cwdVal, ok := pairs["cwd"]; ok && !Equivalent(cwdVal, NullValue) {
		cwd, ok := stringArg(cwdVal)
		if !ok {
			return processStageSpec{}, &RuntimeError{Message: "cmd cwd must be string"}
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
			return processStageSpec{}, &RuntimeError{Message: "cmd inheritEnv must be bool"}
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

func parseStreamType(val Value, field string) (string, error) {
	mode, ok := stringArg(val)
	if !ok {
		return "", &RuntimeError{Message: "process " + field + " must be string"}
	}
	mode = strings.TrimSpace(strings.ToLower(mode))
	switch mode {
	case streamTypeText, streamTypeBytes:
		return mode, nil
	default:
		return "", &RuntimeError{Message: "process " + field + " must be \"text\" or \"bytes\""}
	}
}

func firstDefinedValue(pairs map[string]Value, keys ...string) (Value, bool) {
	for _, key := range keys {
		if val, ok := pairs[key]; ok {
			return val, true
		}
	}
	return nil, false
}

func processPipeInfix(left Value, right Value) (Value, error) {
	leftStages, err := processStagesFromValue(left)
	if err != nil {
		return nil, err
	}
	rightStages, err := processStagesFromValue(right)
	if err != nil {
		return nil, err
	}
	combined := make([]processStageSpec, 0, len(leftStages)+len(rightStages))
	combined = append(combined, leftStages...)
	combined = append(combined, rightStages...)
	return &ProcessPipeline{Stages: combined}, nil
}

func processStagesFromValue(v Value) ([]processStageSpec, error) {
	switch p := v.(type) {
	case *ProcessCommand:
		return []processStageSpec{p.Stage}, nil
	case *ProcessPipeline:
		return append([]processStageSpec(nil), p.Stages...), nil
	default:
		return nil, &RuntimeError{Message: "operator '|' expects cmd or pipeline operands"}
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
