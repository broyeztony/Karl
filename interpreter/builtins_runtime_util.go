package interpreter

import (
	"strconv"
	"time"
)

func registerRuntimeUtilityBuiltins() {
	builtins["rand"] = &Builtin{Name: "rand", Fn: builtinRand}
	builtins["randInt"] = &Builtin{Name: "randInt", Fn: builtinRandInt}
	builtins["randFloat"] = &Builtin{Name: "randFloat", Fn: builtinRandFloat}
	builtins["parseInt"] = &Builtin{Name: "parseInt", Fn: builtinParseInt}
	builtins["now"] = &Builtin{Name: "now", Fn: builtinNow}
}

func builtinParseInt(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "parseInt expects 1 argument"}
	}
	s, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "parseInt expects string"}
	}
	n, err := strconv.ParseInt(s.Value, 10, 64)
	if err != nil {
		return nil, &RuntimeError{Message: "invalid integer: " + s.Value}
	}
	return &Integer{Value: n}, nil
}

func builtinNow(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "now expects no arguments"}
	}
	return &Integer{Value: time.Now().UnixNano() / int64(time.Millisecond)}, nil
}
