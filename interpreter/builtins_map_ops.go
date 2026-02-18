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
		return nil, &RuntimeError{Message: "get expects map/object and key"}
	}
	
	switch val := args[0].(type) {
	case *Map:
		key, err := mapKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		if v, ok := val.Pairs[key]; ok {
			return v, nil
		}
		return NullValue, nil
	
	case *Object:
		keyStr, ok := stringArg(args[1])
		if !ok {
			return nil, &RuntimeError{Message: "get on object expects string key"}
		}
		if v, ok := val.Pairs[keyStr]; ok {
			return v, nil
		}
		return NullValue, nil
		
	default:
		return nil, &RuntimeError{Message: "get expects map or object as first argument"}
	}
}

func builtinMapSet(_ *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return &Set{Elements: make(map[MapKey]struct{})}, nil
	}
	if len(args) == 1 {
		arr, ok := args[0].(*Array)
		if !ok {
			return nil, &RuntimeError{Message: "set expects array when called with 1 argument"}
		}
		newSet := &Set{Elements: make(map[MapKey]struct{})}
		for _, e := range arr.Elements {
			key, err := setKeyForValue(e)
			if err != nil {
				return nil, err
			}
			newSet.Elements[key] = struct{}{}
		}
		return newSet, nil
	}
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "set expects: (), (array), or (map/object, key, value)"}
	}
	
	switch container := args[0].(type) {
	case *Map:
		key, err := mapKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		container.Pairs[key] = args[2]
		return container, nil
		
	case *Object:
		keyStr, ok := stringArg(args[1])
		if !ok {
			return nil, &RuntimeError{Message: "set on object expects string key"}
		}
		// If Pairs is nil, should initialize it?
		// Assuming Pairs is always initialized for Objects created via literals.
		if container.Pairs == nil {
			container.Pairs = make(map[string]Value)
		}
		container.Pairs[keyStr] = args[2]
		return container, nil
		
	default:
		return nil, &RuntimeError{Message: "set expects map or object as first argument for key insertion"}
	}
}

func builtinMapKeys(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "keys expects map/object"}
	}
	
	switch val := args[0].(type) {
	case *Map:
		out := make([]Value, 0, len(val.Pairs))
		for k := range val.Pairs {
			out = append(out, mapKeyToValue(k))
		}
		return &Array{Elements: out}, nil
		
	case *Object:
		out := make([]Value, 0, len(val.Pairs))
		for k := range val.Pairs {
			out = append(out, &String{Value: k})
		}
		return &Array{Elements: out}, nil
		
	default:
		return nil, &RuntimeError{Message: "keys expects map or object"}
	}
}
