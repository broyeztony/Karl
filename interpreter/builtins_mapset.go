package interpreter

func builtinMapHas(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "has expects map or set and key"}
	}
	switch target := args[0].(type) {
	case *Map:
		key, err := mapKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		_, ok := target.Pairs[key]
		return &Boolean{Value: ok}, nil
	case *Set:
		key, err := setKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		_, ok := target.Elements[key]
		return &Boolean{Value: ok}, nil
	default:
		return nil, &RuntimeError{Message: "has expects map or set as first argument"}
	}
}

func builtinMapDelete(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "delete expects map or set and key"}
	}
	switch target := args[0].(type) {
	case *Map:
		key, err := mapKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		_, ok := target.Pairs[key]
		if ok {
			delete(target.Pairs, key)
		}
		return &Boolean{Value: ok}, nil
	case *Set:
		key, err := setKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		_, ok := target.Elements[key]
		if ok {
			delete(target.Elements, key)
		}
		return &Boolean{Value: ok}, nil
	default:
		return nil, &RuntimeError{Message: "delete expects map or set as first argument"}
	}
}

func builtinMapValues(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "values expects map or set"}
	}
	switch target := args[0].(type) {
	case *Map:
		out := make([]Value, 0, len(target.Pairs))
		for _, val := range target.Pairs {
			out = append(out, val)
		}
		return &Array{Elements: out}, nil
	case *Set:
		out := make([]Value, 0, len(target.Elements))
		for key := range target.Elements {
			out = append(out, mapKeyToValue(key))
		}
		return &Array{Elements: out}, nil
	default:
		return nil, &RuntimeError{Message: "values expects map or set as first argument"}
	}
}
