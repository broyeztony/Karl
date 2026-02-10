package interpreter

import "strings"

func stringArg(val Value) (string, bool) {
	switch v := val.(type) {
	case *String:
		return v.Value, true
	case *Char:
		return v.Value, true
	default:
		return "", false
	}
}

func builtinSplit(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "split expects string and separator"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "split expects string as first argument"}
	}
	sep, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "split expects string separator"}
	}
	if sep == "" {
		runes := []rune(str.Value)
		out := make([]Value, 0, len(runes))
		for _, r := range runes {
			out = append(out, &String{Value: string(r)})
		}
		return &Array{Elements: out}, nil
	}
	parts := strings.Split(str.Value, sep)
	out := make([]Value, 0, len(parts))
	for _, part := range parts {
		out = append(out, &String{Value: part})
	}
	return &Array{Elements: out}, nil
}

func builtinChars(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "chars expects string"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "chars expects string as first argument"}
	}
	runes := []rune(str.Value)
	out := make([]Value, 0, len(runes))
	for _, r := range runes {
		out = append(out, &Char{Value: string(r)})
	}
	return &Array{Elements: out}, nil
}

func builtinTrim(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "trim expects string"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "trim expects string as first argument"}
	}
	return &String{Value: strings.TrimSpace(str.Value)}, nil
}

func builtinToLower(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "toLower expects string"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "toLower expects string as first argument"}
	}
	return &String{Value: strings.ToLower(str.Value)}, nil
}

func builtinToUpper(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "toUpper expects string"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "toUpper expects string as first argument"}
	}
	return &String{Value: strings.ToUpper(str.Value)}, nil
}

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
