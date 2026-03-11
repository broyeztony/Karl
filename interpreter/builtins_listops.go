package interpreter

import "sort"

func registerListBuiltins() {
	builtins["sort"] = &Builtin{Name: "sort", Fn: builtinSort}
	builtins["filter"] = &Builtin{Name: "filter", Fn: builtinFilter}
	builtins["flatMap"] = &Builtin{Name: "flatMap", Fn: builtinFlatMap}
	builtins["take"] = &Builtin{Name: "take", Fn: builtinTake}
	builtins["drop"] = &Builtin{Name: "drop", Fn: builtinDrop}
	builtins["chunk"] = &Builtin{Name: "chunk", Fn: builtinChunk}
	builtins["window"] = &Builtin{Name: "window", Fn: builtinWindow}
	builtins["throttle"] = &Builtin{Name: "throttle", Fn: builtinThrottle}
	builtins["distinct"] = &Builtin{Name: "distinct", Fn: builtinDistinct}
	builtins["partition"] = &Builtin{Name: "partition", Fn: builtinPartition}
	builtins["reduce"] = &Builtin{Name: "reduce", Fn: builtinReduce}
	builtins["count"] = &Builtin{Name: "count", Fn: builtinCount}
	builtins["group_count"] = &Builtin{Name: "group_count", Fn: builtinGroupCount}
	builtins["reduce_by_key"] = &Builtin{Name: "reduce_by_key", Fn: builtinReduceByKey}
	builtins["top"] = &Builtin{Name: "top", Fn: builtinTop}
	builtins["sum"] = &Builtin{Name: "sum", Fn: builtinSum}
	builtins["find"] = &Builtin{Name: "find", Fn: builtinFind}
}

func builtinSort(e *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return newStreamSortStage(e, nil), nil
	}
	if len(args) == 1 {
		if !isCallable(args[0]) {
			return nil, &RuntimeError{Message: "sort stage expects comparator function"}
		}
		return newStreamSortStage(e, args[0]), nil
	}
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
	if len(args) == 1 {
		if !isCallable(args[0]) {
			return nil, &RuntimeError{Message: "filter stage expects function"}
		}
		return newStreamFilterStage(e, args[0]), nil
	}
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

func builtinFlatMap(e *Evaluator, args []Value) (Value, error) {
	if len(args) == 1 {
		if !isCallable(args[0]) {
			return nil, &RuntimeError{Message: "flatMap stage expects function"}
		}
		return newStreamFlatMapStage(e, args[0]), nil
	}
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "flatMap expects array and function"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "flatMap expects array"}
	}
	fn := args[1]
	out := []Value{}
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{el})
		if err != nil {
			return nil, err
		}
		mapped, ok := val.(*Array)
		if !ok {
			return nil, &RuntimeError{Message: "flatMap mapper must return array"}
		}
		out = append(out, mapped.Elements...)
	}
	return &Array{Elements: out}, nil
}

func builtinReduce(e *Evaluator, args []Value) (Value, error) {
	if len(args) == 2 {
		if !isCallable(args[1]) {
			return nil, &RuntimeError{Message: "reduce sink expects function as second argument"}
		}
		return newStreamReduceSink(e, args[0], args[1]), nil
	}
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

func builtinCount(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "count expects no arguments"}
	}
	return newStreamCountSink(), nil
}

func builtinDistinct(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "distinct expects no arguments"}
	}
	return newStreamDistinctStage(), nil
}

func builtinGroupCount(e *Evaluator, args []Value) (Value, error) {
	if len(args) > 1 {
		return nil, &RuntimeError{Message: "group_count expects optional key function"}
	}
	if len(args) == 1 && !isCallable(args[0]) {
		return nil, &RuntimeError{Message: "group_count expects function"}
	}
	var keyFn Value
	if len(args) == 1 {
		keyFn = args[0]
	}
	return newStreamGroupCountSink(e, keyFn), nil
}

func builtinReduceByKey(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "reduce_by_key expects (keyFn, initial, reducerFn)"}
	}
	if !isCallable(args[0]) {
		return nil, &RuntimeError{Message: "reduce_by_key keyFn must be function"}
	}
	if !isCallable(args[2]) {
		return nil, &RuntimeError{Message: "reduce_by_key reducerFn must be function"}
	}
	return newStreamReduceByKeySink(e, args[0], args[1], args[2]), nil
}

func builtinTop(e *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "top expects (n, scoreFn?)"}
	}
	n, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "top expects integer n"}
	}
	if n.Value < 0 {
		return nil, &RuntimeError{Message: "top expects non-negative n"}
	}
	var scoreFn Value
	if len(args) == 2 {
		if !isCallable(args[1]) {
			return nil, &RuntimeError{Message: "top scoreFn must be function"}
		}
		scoreFn = args[1]
	}
	return newStreamTopSink(e, n.Value, scoreFn), nil
}

func builtinTake(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "take expects 1 argument"}
	}
	n, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "take expects integer"}
	}
	if n.Value < 0 {
		return nil, &RuntimeError{Message: "take expects non-negative integer"}
	}
	return newStreamTakeStage(n.Value), nil
}

func builtinDrop(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "drop expects 1 argument"}
	}
	n, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "drop expects integer"}
	}
	if n.Value < 0 {
		return nil, &RuntimeError{Message: "drop expects non-negative integer"}
	}
	return newStreamDropStage(n.Value), nil
}

func builtinChunk(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "chunk expects 1 argument"}
	}
	size, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "chunk expects integer size"}
	}
	if size.Value <= 0 {
		return nil, &RuntimeError{Message: "chunk expects size > 0"}
	}
	return newStreamChunkStage(size.Value), nil
}

func builtinWindow(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "window expects (size, step)"}
	}
	size, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "window size must be integer"}
	}
	step, ok := args[1].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "window step must be integer"}
	}
	if size.Value <= 0 {
		return nil, &RuntimeError{Message: "window size must be > 0"}
	}
	if step.Value <= 0 {
		return nil, &RuntimeError{Message: "window step must be > 0"}
	}
	return newStreamWindowStage(size.Value, step.Value), nil
}

func builtinThrottle(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "throttle expects 1 argument"}
	}
	ms, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "throttle expects integer milliseconds"}
	}
	if ms.Value < 0 {
		return nil, &RuntimeError{Message: "throttle expects non-negative milliseconds"}
	}
	return newStreamThrottleStage(e, ms.Value), nil
}

func builtinPartition(e *Evaluator, args []Value) (Value, error) {
	if len(args) == 1 {
		if !isCallable(args[0]) {
			return nil, &RuntimeError{Message: "partition expects function"}
		}
		return newStreamPartitionSink(e, args[0], nil), nil
	}
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "partition expects (array, fn) or (selectorFn, branchKeys)"}
	}
	if isCallable(args[0]) {
		keys, err := parsePartitionBranchKeys(args[1])
		if err != nil {
			return nil, err
		}
		return newStreamPartitionSink(e, args[0], keys), nil
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "partition expects array"}
	}
	fn := args[1]
	pass := []Value{}
	fail := []Value{}
	for _, el := range arr.Elements {
		matches, err := applyListPredicate(e, fn, el, "partition")
		if err != nil {
			return nil, err
		}
		if matches {
			pass = append(pass, el)
		} else {
			fail = append(fail, el)
		}
	}
	return &Object{Pairs: map[string]Value{
		"pass": &Array{Elements: pass},
		"fail": &Array{Elements: fail},
	}}, nil
}

func parsePartitionBranchKeys(v Value) ([]string, error) {
	arr, ok := v.(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "partition branchKeys must be array of strings"}
	}
	if len(arr.Elements) == 0 {
		return nil, &RuntimeError{Message: "partition branchKeys must not be empty"}
	}
	keys := make([]string, 0, len(arr.Elements))
	seen := map[string]struct{}{}
	for _, el := range arr.Elements {
		s, ok := el.(*String)
		if !ok {
			return nil, &RuntimeError{Message: "partition branchKeys must contain only strings"}
		}
		name := s.Value
		if name == "" {
			return nil, &RuntimeError{Message: "partition branchKeys must not contain empty string"}
		}
		if _, exists := seen[name]; exists {
			return nil, &RuntimeError{Message: "partition branchKeys must be unique"}
		}
		seen[name] = struct{}{}
		keys = append(keys, name)
	}
	return keys, nil
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
