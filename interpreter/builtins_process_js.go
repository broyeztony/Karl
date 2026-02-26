//go:build js

package interpreter

func registerProcessBuiltins() {
	builtins["processRun"] = &Builtin{Name: "processRun", Fn: builtinProcessRun}
}

func builtinProcessRun(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "processRun expects options object"}
	}
	return nil, recoverableError("process_spawn", "processRun is not supported in this runtime")
}
