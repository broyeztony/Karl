package interpreter

import (
	"time"
)

func builtinSend(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "send expects channel and value"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "send expects channel"}
	}
	closedCh := ch.ClosedSignal()
	fatalCh := runtimeFatalSignal(e)
	cancelCh := runtimeCancelSignal(e)

	ticker := time.NewTicker(runtimeBlockProbeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-closedCh:
			return nil, &RuntimeError{Message: "send on closed channel"}
		case ch.Ch <- args[1]:
			return UnitValue, nil
		case <-cancelCh:
			return nil, canceledError()
		case <-fatalCh:
			return nil, runtimeFatalError(e)
		case <-ticker.C:
			if isTopLevelRuntimeDeadlocked(e) {
				select {
				case <-closedCh:
					return nil, &RuntimeError{Message: "send on closed channel"}
				default:
				}
				return nil, &RuntimeError{Message: "deadlock: send would block with no runnable tasks"}
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
