//go:build js

package interpreter

import "fmt"

type processStageSpec struct {
	command string
	args    []string
}

type processSpec struct {
	stdinMode string
	stdinText *string
}

type Process struct{}

func (p *Process) Type() ValueType { return PROCESS }
func (p *Process) Inspect() string { return "<process>" }
func (p *Process) PID() int64      { return 0 }
func (p *Process) Running() bool   { return false }
func (p *Process) Abort() error    { return fmt.Errorf("process API is not supported in this runtime") }
func (p *Process) Kill() error     { return fmt.Errorf("process API is not supported in this runtime") }
func (p *Process) Signal(string) error {
	return fmt.Errorf("process API is not supported in this runtime")
}
func (p *Process) inputStream() (*StreamWriter, bool) { return nil, false }
func (p *Process) outputStream() (*StreamReader, bool) {
	return nil, false
}
func (p *Process) errorStream() (*StreamReader, bool) { return nil, false }

func registerProcessBuiltins() {
	builtins["proc"] = &Builtin{Name: "proc", Fn: builtinProc}
	builtins["run"] = &Builtin{Name: "run", Fn: builtinRun}
}

const processModePipe = "pipe"

func builtinProc(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "proc expects (spec, opts?)"}
	}
	return nil, recoverableError("process_spawn", "process API is not supported in this runtime")
}

func builtinRun(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "run expects (spec, opts?)"}
	}
	return nil, recoverableError("process_spawn", "process API is not supported in this runtime")
}

func processAwaitWithCancel(_ *Process, _ <-chan struct{}, _ *runtimeState) (Value, *Signal, error) {
	return nil, nil, recoverableError("process_state", "process API is not supported in this runtime")
}

func parseRunSpec(_ []Value) (processSpec, error) {
	return processSpec{}, recoverableError("process_spawn", "process API is not supported in this runtime")
}

func executeRunSpec(_ *Evaluator, _ processSpec) (Value, error) {
	return nil, recoverableError("process_spawn", "process API is not supported in this runtime")
}
