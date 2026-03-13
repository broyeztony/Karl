package interpreter

import (
	"time"
)

func builtinSend(e *Evaluator, args []Value) (Value, error) {
	if len(args) == 1 {
		ch, ok := args[0].(*Channel)
		if !ok {
			return nil, &RuntimeError{Message: "send expects channel"}
		}
		return &StreamSinkValue{
			name: "send",
			run: func(e *Evaluator, upstream streamIterator) (Value, error) {
				defer ch.Close()
				for {
					item, eof, err := upstream.Next()
					if err != nil {
						return nil, recoverableError("stream_read", "stream read error: "+err.Error())
					}
					if eof {
						return UnitValue, nil
					}
					if err := channelSendBlocking(e, ch, item); err != nil {
						return nil, err
					}
				}
			},
		}, nil
	}
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "send expects channel and value"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "send expects channel"}
	}
	if err := channelSendBlocking(e, ch, args[1]); err != nil {
		return nil, err
	}
	return UnitValue, nil
}

func channelSendBlocking(e *Evaluator, ch *Channel, value Value) error {
	if ch == nil {
		return &RuntimeError{Message: "send expects channel"}
	}
	closedCh := ch.ClosedSignal()
	fatalCh := runtimeFatalSignal(e)
	cancelCh := runtimeCancelSignal(e)

	ticker := time.NewTicker(runtimeBlockProbeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-closedCh:
			return &RuntimeError{Message: "send on closed channel"}
		case ch.Ch <- value:
			return nil
		case <-cancelCh:
			return canceledError()
		case <-fatalCh:
			return runtimeFatalError(e)
		case <-ticker.C:
			if isTopLevelRuntimeDeadlocked(e) {
				select {
				case <-closedCh:
					return &RuntimeError{Message: "send on closed channel"}
				default:
				}
				return &RuntimeError{Message: "deadlock: send would block with no runnable tasks"}
			}
		}
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
	closedCh := ch.ClosedSignal()
	fatalCh := runtimeFatalSignal(e)
	cancelCh := runtimeCancelSignal(e)

	ticker := time.NewTicker(runtimeBlockProbeInterval)
	defer ticker.Stop()

	for {
		// Prefer draining queued values before reporting closed.
		select {
		case val := <-ch.Ch:
			return &Array{Elements: []Value{val, &Boolean{Value: false}}}, nil
		default:
		}

		select {
		case <-closedCh:
			return &Array{Elements: []Value{NullValue, &Boolean{Value: true}}}, nil
		default:
		}

		select {
		case val := <-ch.Ch:
			return &Array{Elements: []Value{val, &Boolean{Value: false}}}, nil
		case <-closedCh:
			select {
			case val := <-ch.Ch:
				return &Array{Elements: []Value{val, &Boolean{Value: false}}}, nil
			default:
				return &Array{Elements: []Value{NullValue, &Boolean{Value: true}}}, nil
			}
		case <-cancelCh:
			return nil, canceledError()
		case <-fatalCh:
			return nil, runtimeFatalError(e)
		case <-ticker.C:
			if !isTopLevelRuntimeDeadlocked(e) {
				continue
			}
			// If data/close became ready while ticker fired, consume it first.
			select {
			case val := <-ch.Ch:
				return &Array{Elements: []Value{val, &Boolean{Value: false}}}, nil
			case <-closedCh:
				return &Array{Elements: []Value{NullValue, &Boolean{Value: true}}}, nil
			default:
			}
			return nil, &RuntimeError{Message: "deadlock: recv would block with no runnable tasks"}
		}
	}
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

func isTopLevelRuntimeDeadlocked(e *Evaluator) bool {
	if e == nil || e.currentTask != nil || e.runtime == nil {
		return false
	}
	return !e.runtime.hasUndoneTasks()
}
