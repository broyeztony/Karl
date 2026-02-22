//go:build js

package interpreter

func registerSignalBuiltins() {
	builtins["signalWatch"] = &Builtin{Name: "signalWatch", Fn: builtinSignalWatch}
}

func builtinSignalWatch(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "signalWatch expects 1 argument"}
	}
	return nil, recoverableError("signalWatch", "signalWatch is not supported in this runtime")
}
