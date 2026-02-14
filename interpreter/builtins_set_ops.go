package interpreter

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
