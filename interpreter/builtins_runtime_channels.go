package interpreter

func builtinChannel(_ *Evaluator, _ []Value) (Value, error) {
	return &Channel{Ch: make(chan Value)}, nil
}

func builtinBufferedChannel(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "buffered expects 1 argument (buffer size)"}
	}
	size, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "buffered expects integer buffer size"}
	}
	if size.Value < 0 {
		return nil, &RuntimeError{Message: "buffered expects non-negative buffer size"}
	}
	if size.Value > 1000000 {
		return nil, &RuntimeError{Message: "buffered buffer size too large (max 1000000)"}
	}
	return &Channel{Ch: make(chan Value, size.Value)}, nil
}
