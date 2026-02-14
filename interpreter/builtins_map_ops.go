package interpreter

func builtinMap(e *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return &Map{Pairs: make(map[MapKey]Value)}, nil
	}
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "map expects no arguments or array and function"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "map expects array as first argument"}
	}
	fn := args[1]
	out := make([]Value, 0, len(arr.Elements))
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{el})
		if err != nil {
			return nil, err
		}
		out = append(out, val)
	}
	return &Array{Elements: out}, nil
}

func builtinMapGet(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "get expects map and key"}
	}
	m, ok := args[0].(*Map)
	if !ok {
		return nil, &RuntimeError{Message: "get expects map as first argument"}
	}
	key, err := mapKeyForValue(args[1])
	if err != nil {
		return nil, err
	}
	val, ok := m.Pairs[key]
	if !ok {
		return NullValue, nil
	}
	return val, nil
}

func builtinMapSet(_ *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return &Set{Elements: make(map[MapKey]struct{})}, nil
	}
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "set expects no arguments or map, key, value"}
	}
	m, ok := args[0].(*Map)
	if !ok {
		return nil, &RuntimeError{Message: "set expects map as first argument"}
	}
	key, err := mapKeyForValue(args[1])
	if err != nil {
		return nil, err
	}
	m.Pairs[key] = args[2]
	return m, nil
}

func builtinMapKeys(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "keys expects map"}
	}
	m, ok := args[0].(*Map)
	if !ok {
		return nil, &RuntimeError{Message: "keys expects map as first argument"}
	}
	out := make([]Value, 0, len(m.Pairs))
	for key := range m.Pairs {
		out = append(out, mapKeyToValue(key))
	}
	return &Array{Elements: out}, nil
}
