//go:build js

package interpreter

func builtinHTTPServe(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "httpServe expects config object"}
	}
	return nil, recoverableError("httpServe", "httpServe is not supported in this runtime")
}

func builtinHTTPServerStop(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "httpServerStop expects server"}
	}
	return UnitValue, nil
}
