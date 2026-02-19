package interpreter

import "unicode/utf8"

func registerCollectionBuiltins() {
	builtins["map"] = &Builtin{Name: "map", Fn: builtinMap}
	builtins["get"] = &Builtin{Name: "get", Fn: builtinMapGet}
	builtins["set"] = &Builtin{Name: "set", Fn: builtinMapSet}
	builtins["add"] = &Builtin{Name: "add", Fn: builtinSetAdd}
	builtins["has"] = &Builtin{Name: "has", Fn: builtinMapHas}
	builtins["delete"] = &Builtin{Name: "delete", Fn: builtinMapDelete}
	builtins["keys"] = &Builtin{Name: "keys", Fn: builtinMapKeys}
	builtins["values"] = &Builtin{Name: "values", Fn: builtinMapValues}
	builtins["len"] = &Builtin{Name: "len", Fn: builtinLen}
}

func builtinLen(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "len expects 1 argument"}
	}
	switch arg := args[0].(type) {
	case *String:
		return &Integer{Value: int64(utf8.RuneCountInString(arg.Value))}, nil
	case *Array:
		return &Integer{Value: int64(len(arg.Elements))}, nil
	case *Map:
		return &Integer{Value: int64(len(arg.Pairs))}, nil
	case *Set:
		return &Integer{Value: int64(len(arg.Elements))}, nil
	case *Object:
		return &Integer{Value: int64(len(arg.Pairs))}, nil
	default:
		return nil, &RuntimeError{Message: "len expects string, array, map, set, or object"}
	}
}
