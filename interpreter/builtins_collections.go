package interpreter

import "sort"

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

func builtinSetAdd(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "add expects set and value"}
	}
	s, ok := args[0].(*Set)
	if !ok {
		return nil, &RuntimeError{Message: "add expects set as first argument"}
	}
	key, err := setKeyForValue(args[1])
	if err != nil {
		return nil, err
	}
	s.Elements[key] = struct{}{}
	return s, nil
}

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

func builtinSort(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "sort expects array and comparator"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "sort expects array as first argument"}
	}
	cmp := args[1]
	out := append([]Value{}, arr.Elements...)
	var cmpErr error
	sort.Slice(out, func(i, j int) bool {
		if cmpErr != nil {
			return false
		}
		val, _, err := e.applyFunction(cmp, []Value{out[i], out[j]})
		if err != nil {
			cmpErr = err
			return false
		}
		num, _, ok := numberArg(val)
		if !ok {
			cmpErr = &RuntimeError{Message: "sort comparator must return number"}
			return false
		}
		return num < 0
	})
	if cmpErr != nil {
		return nil, cmpErr
	}
	return &Array{Elements: out}, nil
}

func builtinFilter(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "filter expects array and function"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "filter expects array"}
	}
	fn := args[1]
	out := []Value{}
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{el})
		if err != nil {
			return nil, err
		}
		b, ok := val.(*Boolean)
		if !ok {
			return nil, &RuntimeError{Message: "filter predicate must return bool"}
		}
		if b.Value {
			out = append(out, el)
		}
	}
	return &Array{Elements: out}, nil
}

func builtinReduce(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "reduce expects array, function, initial"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "reduce expects array"}
	}
	fn := args[1]
	acc := args[2]
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{acc, el})
		if err != nil {
			return nil, err
		}
		acc = val
	}
	return acc, nil
}

func builtinSum(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sum expects array"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "sum expects array"}
	}
	var total float64
	allInts := true
	for _, el := range arr.Elements {
		switch v := el.(type) {
		case *Integer:
			total += float64(v.Value)
		case *Float:
			allInts = false
			total += v.Value
		default:
			return nil, &RuntimeError{Message: "sum expects numeric array"}
		}
	}
	if allInts {
		return &Integer{Value: int64(total)}, nil
	}
	return &Float{Value: total}, nil
}

func builtinFind(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "find expects array and function"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "find expects array"}
	}
	fn := args[1]
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{el})
		if err != nil {
			return nil, err
		}
		b, ok := val.(*Boolean)
		if !ok {
			return nil, &RuntimeError{Message: "find predicate must return bool"}
		}
		if b.Value {
			return el, nil
		}
	}
	return NullValue, nil
}
