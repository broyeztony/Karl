package interpreter

import "strings"

func builtinContains(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "contains expects string and substring"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "contains expects string as first argument"}
	}
	sub, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "contains expects string substring"}
	}
	return &Boolean{Value: strings.Contains(str.Value, sub)}, nil
}

func builtinStartsWith(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "startsWith expects string and prefix"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "startsWith expects string as first argument"}
	}
	prefix, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "startsWith expects string prefix"}
	}
	return &Boolean{Value: strings.HasPrefix(str.Value, prefix)}, nil
}

func builtinEndsWith(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "endsWith expects string and suffix"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "endsWith expects string as first argument"}
	}
	suffix, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "endsWith expects string suffix"}
	}
	return &Boolean{Value: strings.HasSuffix(str.Value, suffix)}, nil
}

func builtinReplace(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "replace expects string, old, new"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "replace expects string as first argument"}
	}
	oldVal, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "replace expects string old value"}
	}
	newVal, ok := stringArg(args[2])
	if !ok {
		return nil, &RuntimeError{Message: "replace expects string new value"}
	}
	return &String{Value: strings.ReplaceAll(str.Value, oldVal, newVal)}, nil
}
