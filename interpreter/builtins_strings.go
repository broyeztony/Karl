package interpreter

import "strings"

func registerStringBuiltins() {
	builtins["split"] = &Builtin{Name: "split", Fn: builtinSplit}
	builtins["chars"] = &Builtin{Name: "chars", Fn: builtinChars}
	builtins["trim"] = &Builtin{Name: "trim", Fn: builtinTrim}
	builtins["toLower"] = &Builtin{Name: "toLower", Fn: builtinToLower}
	builtins["toUpper"] = &Builtin{Name: "toUpper", Fn: builtinToUpper}
	builtins["contains"] = &Builtin{Name: "contains", Fn: builtinContains}
	builtins["startsWith"] = &Builtin{Name: "startsWith", Fn: builtinStartsWith}
	builtins["endsWith"] = &Builtin{Name: "endsWith", Fn: builtinEndsWith}
	builtins["replace"] = &Builtin{Name: "replace", Fn: builtinReplace}
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
