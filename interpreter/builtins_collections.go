package interpreter

import "unicode/utf8"

func registerCollectionBuiltins() {
	builtins["map"] = &Builtin{Name: "map", Fn: builtinMap}
	builtins["tree"] = &Builtin{Name: "tree", Fn: builtinTree}
	builtins["ntree"] = &Builtin{Name: "ntree", Fn: builtinNTree}
	builtins["get"] = &Builtin{Name: "get", Fn: builtinMapGet}
	builtins["set"] = &Builtin{Name: "set", Fn: builtinMapSet}
	builtins["add"] = &Builtin{Name: "add", Fn: builtinSetAdd}
	builtins["has"] = &Builtin{Name: "has", Fn: builtinMapHas}
	builtins["delete"] = &Builtin{Name: "delete", Fn: builtinMapDelete}
	builtins["keys"] = &Builtin{Name: "keys", Fn: builtinMapKeys}
	builtins["values"] = &Builtin{Name: "values", Fn: builtinMapValues}
	builtins["len"] = &Builtin{Name: "len", Fn: builtinLen}
}

func builtinTree(_ *Evaluator, args []Value) (Value, error) {
	if len(args) > 1 {
		return nil, &RuntimeError{Message: "tree expects optional kind"}
	}
	kind := "avl"
	if len(args) == 1 {
		k, ok := stringArg(args[0])
		if !ok {
			return nil, &RuntimeError{Message: "tree kind must be string"}
		}
		kind = k
	}
	return newTree(kind)
}

func builtinNTree(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "ntree expects (rootId, rootValue?)"}
	}
	rootID, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "ntree root id must be string"}
	}
	rootValue := Value(NullValue)
	if len(args) == 2 {
		rootValue = args[1]
	}
	return newNTree(rootID, rootValue)
}

func builtinLen(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "len expects 1 argument"}
	}
	switch arg := args[0].(type) {
	case *String:
		return &Integer{Value: int64(utf8.RuneCountInString(arg.Value))}, nil
	case *Bytes:
		return &Integer{Value: int64(len(arg.Value))}, nil
	case *Array:
		return &Integer{Value: int64(len(arg.Elements))}, nil
	case *Map:
		return &Integer{Value: int64(len(arg.Pairs))}, nil
	case *Set:
		return &Integer{Value: int64(len(arg.Elements))}, nil
	case *Tree:
		return &Integer{Value: int64(arg.size)}, nil
	case *NTree:
		return &Integer{Value: int64(len(arg.nodes))}, nil
	case *Object:
		return &Integer{Value: int64(len(arg.Pairs))}, nil
	default:
		return nil, &RuntimeError{Message: "len expects string, bytes, array, map, set, tree, ntree, or object"}
	}
}
