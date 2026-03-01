//go:build js

package interpreter

import "fmt"

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
func (p *Process) inputChannel() (*Channel, bool)  { return nil, false }
func (p *Process) outputChannel() (*Channel, bool) { return nil, false }
func (p *Process) errorChannel() (*Channel, bool)  { return nil, false }

func registerProcessBuiltins() {
	builtins["cmd"] = &Builtin{Name: "cmd", Fn: builtinCmd}
	builtins["proc"] = &Builtin{Name: "proc", Fn: builtinProc}
	builtins["run"] = &Builtin{Name: "run", Fn: builtinRun}
}

func builtinCmd(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "cmd expects 1 object argument"}
	}
	return nil, recoverableError("process_spawn", "process API is not supported in this runtime")
}

func builtinProc(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "proc expects (cmdOrPipeline, opts?)"}
	}
	return nil, recoverableError("process_spawn", "process API is not supported in this runtime")
}

func builtinRun(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "run expects (cmdOrPipeline, opts?)"}
	}
	return nil, recoverableError("process_spawn", "process API is not supported in this runtime")
}

func processPipeInfix(left Value, right Value) (Value, error) {
	return nil, recoverableError("process_state", "process API is not supported in this runtime")
}

func processAwaitWithCancel(_ *Process, _ <-chan struct{}, _ *runtimeState) (Value, *Signal, error) {
	return nil, nil, recoverableError("process_state", "process API is not supported in this runtime")
}
