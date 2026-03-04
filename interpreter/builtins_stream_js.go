//go:build js

package interpreter

func registerStreamBuiltins() {
	builtins["reader"] = &Builtin{Name: "reader", Fn: builtinReader}
	builtins["writer"] = &Builtin{Name: "writer", Fn: builtinWriter}
	builtins["pipe"] = &Builtin{Name: "pipe", Fn: builtinPipe}
}

func builtinReader(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "reader expects (path, opts?)"}
	}
	return nil, recoverableError("stream_open", "stream API is not supported in this runtime")
}

func builtinWriter(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "writer expects (path, opts?)"}
	}
	return nil, recoverableError("stream_open", "stream API is not supported in this runtime")
}

func builtinPipe(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, &RuntimeError{Message: "pipe expects (srcReader, dstWriter, opts?)"}
	}
	return nil, recoverableError("stream_state", "stream API is not supported in this runtime")
}
