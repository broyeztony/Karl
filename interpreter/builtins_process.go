//go:build !js

package interpreter

import (
	"bufio"
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
	command string
	args    []string
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

	stdinCh  *Channel
	stdoutCh *Channel
	stderrCh *Channel

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

func (p *Process) inputChannel() (*Channel, bool) {
	if p == nil {
		return nil, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stdinMode != processModePipe || p.stdinCh == nil {
		return nil, false
	}
	return p.stdinCh, true
}

func (p *Process) outputChannel() (*Channel, bool) {
	if p == nil {
		return nil, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stdoutMode != processModePipe || p.stdoutCh == nil {
		return nil, false
	}
	return p.stdoutCh, true
}

func (p *Process) errorChannel() (*Channel, bool) {
	if p == nil {
		return nil, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stderrMode != processModePipe || p.stderrCh == nil {
		return nil, false
	}
	return p.stderrCh, true
}

type processWaitResult struct {
	status Value
	err    error
}

type processSpec struct {
	stages []processStageSpec

	cwd        string
	env        map[string]string
	inheritEnv bool
	timeoutMs  int64

	stdinMode  string
	stdoutMode string
	stderrMode string

	stdinText *string

	maxOutputBytes int64
	overflow       string
}

func registerProcessBuiltins() {
	builtins["cmd"] = &Builtin{Name: "cmd", Fn: builtinCmd}
	builtins["proc"] = &Builtin{Name: "proc", Fn: builtinProc}
	builtins["run"] = &Builtin{Name: "run", Fn: builtinRun}
	builtins["stdIn"] = &Builtin{Name: "stdIn", Fn: builtinStdIn}
	builtins["stdOut"] = &Builtin{Name: "stdOut", Fn: builtinStdOut}
	builtins["stdErr"] = &Builtin{Name: "stdErr", Fn: builtinStdErr}
}

func builtinCmd(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, &RuntimeError{Message: "cmd expects (command) or (command, args)"}
	}

	if len(args) == 1 {
		if pairs, ok := objectPairs(args[0]); ok {
			stage, err := parseProcessStageFromPairs(pairs)
			if err != nil {
				return nil, err
			}
			return &ProcessCommand{Stage: stage}, nil
		}
	}

	command, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "cmd command must be string"}
	}
	if strings.TrimSpace(command) == "" {
		return nil, &RuntimeError{Message: "cmd command must not be empty"}
	}
	stage := processStageSpec{command: command}
	if len(args) == 2 {
		parsed, err := parseProcessArgs(args[1], "cmd args")
		if err != nil {
			return nil, err
		}
		stage.args = parsed
	}
	return &ProcessCommand{Stage: stage}, nil
}

func builtinProc(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "proc expects 1 argument"}
	}
	spec, err := parseProcSpec(args[0])
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
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "run expects 1 argument"}
	}
	spec, err := parseRunSpec(args[0])
	if err != nil {
		return nil, err
	}

	process, err := startProcess(e, spec)
	if err != nil {
		return nil, err
	}

	stdout, ok := process.outputChannel()
	if !ok {
		return nil, recoverableError("process_state", "run capture unavailable: stdOut is not piped")
	}
	stderr, ok := process.errorChannel()
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
		value, truncated := collectProcessChannel(stdout, spec.maxOutputBytes, spec.overflow)
		outCh <- captureResult{value: value, truncated: truncated}
	}()
	go func() {
		value, truncated := collectProcessChannel(stderr, spec.maxOutputBytes, spec.overflow)
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

func builtinStdIn(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "stdIn expects process"}
	}
	p, ok := args[0].(*Process)
	if !ok {
		return nil, &RuntimeError{Message: "stdIn expects process"}
	}
	ch, ok := p.inputChannel()
	if !ok {
		return nil, recoverableError("process_state", "stdIn is only available when stdIn mode is \"pipe\"")
	}
	return ch, nil
}

func builtinStdOut(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "stdOut expects process"}
	}
	p, ok := args[0].(*Process)
	if !ok {
		return nil, &RuntimeError{Message: "stdOut expects process"}
	}
	ch, ok := p.outputChannel()
	if !ok {
		return nil, recoverableError("process_state", "stdOut is only available when stdOut mode is \"pipe\"")
	}
	return ch, nil
}

func builtinStdErr(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "stdErr expects process"}
	}
	p, ok := args[0].(*Process)
	if !ok {
		return nil, &RuntimeError{Message: "stdErr expects process"}
	}
	ch, ok := p.errorChannel()
	if !ok {
		return nil, recoverableError("process_state", "stdErr is only available when stdErr mode is \"pipe\"")
	}
	return ch, nil
}

func parseProcSpec(val Value) (processSpec, error) {
	spec := processSpec{
		inheritEnv: true,
		stdinMode:  processModeInherit,
		stdoutMode: processModeInherit,
		stderrMode: processModeInherit,
	}
	return parseProcessSpecValue(val, spec, false)
}

func parseRunSpec(val Value) (processSpec, error) {
	spec := processSpec{
		inheritEnv:     true,
		stdinMode:      processModeNull,
		stdoutMode:     processModePipe,
		stderrMode:     processModePipe,
		maxOutputBytes: defaultProcessCaptureBytes,
		overflow:       processOverflowTruncate,
	}
	return parseProcessSpecValue(val, spec, true)
}

func parseProcessSpecValue(val Value, base processSpec, forRun bool) (processSpec, error) {
	spec := base
	switch v := val.(type) {
	case *ProcessCommand:
		spec.stages = []processStageSpec{v.Stage}
		return spec, nil
	case *ProcessPipeline:
		spec.stages = append([]processStageSpec(nil), v.Stages...)
		return spec, nil
	}

	pairs, ok := objectPairs(val)
	if !ok {
		if forRun {
			return processSpec{}, &RuntimeError{Message: "run expects command spec or cmd/pipeline"}
		}
		return processSpec{}, &RuntimeError{Message: "proc expects command spec or cmd/pipeline"}
	}

	if planVal, ok := pairs["plan"]; ok && !Equivalent(planVal, NullValue) {
		stages, err := parseProcessPlanStages(planVal)
		if err != nil {
			return processSpec{}, err
		}
		spec.stages = stages
	} else {
		stage, err := parseProcessStageFromPairs(pairs)
		if err != nil {
			return processSpec{}, err
		}
		spec.stages = []processStageSpec{stage}
	}

	if cwdVal, ok := pairs["cwd"]; ok && !Equivalent(cwdVal, NullValue) {
		cwd, ok := stringArg(cwdVal)
		if !ok {
			return processSpec{}, &RuntimeError{Message: "process cwd must be string"}
		}
		spec.cwd = cwd
	}

	if envVal, ok := pairs["env"]; ok && !Equivalent(envVal, NullValue) {
		env, err := parseProcessEnv(envVal)
		if err != nil {
			return processSpec{}, err
		}
		spec.env = env
	}

	if inheritVal, ok := pairs["inheritEnv"]; ok && !Equivalent(inheritVal, NullValue) {
		inherit, ok := inheritVal.(*Boolean)
		if !ok {
			return processSpec{}, &RuntimeError{Message: "process inheritEnv must be bool"}
		}
		spec.inheritEnv = inherit.Value
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

	if forRun {
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

	if modeVal, ok := pairs["stdIn"]; ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stdIn")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdinMode = mode
	}
	if modeVal, ok := pairs["stdOut"]; ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stdOut")
		if err != nil {
			return processSpec{}, err
		}
		spec.stdoutMode = mode
	}
	if modeVal, ok := pairs["stdErr"]; ok && !Equivalent(modeVal, NullValue) {
		mode, err := parseProcessMode(modeVal, "stdErr")
		if err != nil {
			return processSpec{}, err
		}
		spec.stderrMode = mode
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
		return nil, &RuntimeError{Message: "process plan must be cmd or pipeline"}
	}
}

func parseProcessStageFromPairs(pairs map[string]Value) (processStageSpec, error) {
	commandVal, ok := pairs["command"]
	if !ok {
		return processStageSpec{}, &RuntimeError{Message: "process command spec expects command"}
	}
	command, ok := stringArg(commandVal)
	if !ok {
		return processStageSpec{}, &RuntimeError{Message: "process command must be string"}
	}
	if strings.TrimSpace(command) == "" {
		return processStageSpec{}, &RuntimeError{Message: "process command must not be empty"}
	}
	stage := processStageSpec{command: command}

	if argsVal, ok := pairs["args"]; ok && !Equivalent(argsVal, NullValue) {
		args, err := parseProcessArgs(argsVal, "process args")
		if err != nil {
			return processStageSpec{}, err
		}
		stage.args = args
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

	fatalCh := runtime.fatalSignal()
	var out processWaitResult
	if cancelCh == nil && fatalCh == nil {
		out = <-p.waitCh
	} else {
		select {
		case out = <-p.waitCh:
		case <-cancelCh:
			return nil, nil, canceledError()
		case <-fatalCh:
			if err := runtime.getFatalTaskFailure(); err != nil {
				return nil, nil, err
			}
			return nil, nil, &RuntimeError{Message: "runtime terminated"}
		}
	}

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
		if spec.cwd != "" {
			dir, err := resolveProcessDir(e, spec.cwd)
			if err != nil {
				cleanup()
				return nil, recoverableError("process_spawn", "process spawn error: "+err.Error())
			}
			cmd.Dir = dir
		}
		env := processEnvSnapshot(e, spec)
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
		input := &Channel{Ch: make(chan Value, 32)}
		process.stdinCh = input
		go processForwardInput(process, input, stdinWriter)
	}
	if stdoutReader != nil {
		out := &Channel{Ch: make(chan Value, 128)}
		process.stdoutCh = out
		go processForwardOutput(process, stdoutReader, out, true)
	}
	if len(stderrReaders) > 0 {
		errCh := &Channel{Ch: make(chan Value, 128)}
		process.stderrCh = errCh
		go processForwardMergedErrors(process, stderrReaders, errCh)
	}

	go func() {
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
	}()

	return process, nil
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

func processForwardInput(p *Process, ch *Channel, writer io.WriteCloser) {
	defer func() {
		_ = writer.Close()
	}()
	for {
		value, ok := <-ch.Ch
		if !ok {
			return
		}
		s, ok := stringArg(value)
		if !ok {
			p.setIOError(fmt.Errorf("stdIn expects string payloads"))
			continue
		}
		if _, err := io.WriteString(writer, s); err != nil {
			p.setIOError(err)
			return
		}
	}
}

func processForwardOutput(p *Process, reader io.ReadCloser, out *Channel, closeOut bool) {
	defer func() {
		_ = reader.Close()
		if closeOut {
			out.Close()
		}
	}()
	r := bufio.NewReader(reader)
	for {
		chunk, err := r.ReadString('\n')
		if len(chunk) > 0 {
			if !channelTrySend(out, &String{Value: chunk}) {
				return
			}
		}
		if err == nil {
			continue
		}
		if errors.Is(err, io.EOF) {
			return
		}
		p.setIOError(err)
		return
	}
}

func processForwardMergedErrors(p *Process, readers []io.ReadCloser, out *Channel) {
	var wg sync.WaitGroup
	for _, reader := range readers {
		r := reader
		wg.Add(1)
		go func() {
			defer wg.Done()
			processForwardOutput(p, r, out, false)
		}()
	}
	wg.Wait()
	out.Close()
}

func collectProcessChannel(ch *Channel, limit int64, overflow string) (string, bool) {
	if ch == nil {
		return "", false
	}
	var builder strings.Builder
	truncated := false
	for value := range ch.Ch {
		s, ok := stringArg(value)
		if !ok {
			continue
		}
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

func processEnvSnapshot(e *Evaluator, spec processSpec) []string {
	if !spec.inheritEnv {
		if len(spec.env) == 0 {
			return []string{}
		}
		return processSortedEnv(spec.env)
	}

	base := runtimeEnviron(e)
	if len(base) == 0 {
		base = os.Environ()
	}
	if len(spec.env) == 0 {
		return cloneStrings(base)
	}

	merged := makeEnvMap(base)
	for key, value := range spec.env {
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
		go func() {
			select {
			case <-cancelCh:
				cancel()
			case <-fatalCh:
				cancel()
			case <-done:
			}
		}()
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
