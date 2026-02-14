package interpreter

import "sort"

func registerListBuiltins() {
	builtins["sort"] = &Builtin{Name: "sort", Fn: builtinSort}
	builtins["filter"] = &Builtin{Name: "filter", Fn: builtinFilter}
	builtins["reduce"] = &Builtin{Name: "reduce", Fn: builtinReduce}
	builtins["sum"] = &Builtin{Name: "sum", Fn: builtinSum}
	builtins["find"] = &Builtin{Name: "find", Fn: builtinFind}
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
		matches, err := applyListPredicate(e, fn, el, "filter")
		if err != nil {
			return nil, err
		}
		if matches {
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
		matches, err := applyListPredicate(e, fn, el, "find")
		if err != nil {
			return nil, err
		}
		if matches {
			return el, nil
		}
	}
	return NullValue, nil
}
