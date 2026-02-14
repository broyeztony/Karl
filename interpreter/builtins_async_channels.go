package interpreter

func builtinSend(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "send expects channel and value"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "send expects channel"}
	}
	if ch.Closed {
		return nil, &RuntimeError{Message: "send on closed channel"}
	}

	fatalCh := runtimeFatalSignal(e)
	cancelCh := runtimeCancelSignal(e)

	if cancelCh == nil && fatalCh == nil {
		ch.Ch <- args[1]
		return UnitValue, nil
	}
	select {
	case ch.Ch <- args[1]:
		return UnitValue, nil
	case <-cancelCh:
		return nil, canceledError()
	case <-fatalCh:
		return nil, runtimeFatalError(e)
	}
}

func builtinRecv(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "recv expects channel"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "recv expects channel"}
	}
	var val Value
	var okRecv bool
	fatalCh := runtimeFatalSignal(e)
	cancelCh := runtimeCancelSignal(e)
	if cancelCh == nil && fatalCh == nil {
		val, okRecv = <-ch.Ch
	} else {
		select {
		case val, okRecv = <-ch.Ch:
		case <-cancelCh:
			return nil, canceledError()
		case <-fatalCh:
			return nil, runtimeFatalError(e)
		}
	}
	if !okRecv {
		return &Array{Elements: []Value{NullValue, &Boolean{Value: true}}}, nil
	}
	return &Array{Elements: []Value{val, &Boolean{Value: false}}}, nil
}

func builtinDone(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "done expects channel"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "done expects channel"}
	}
	ch.Close()
	return UnitValue, nil
}
