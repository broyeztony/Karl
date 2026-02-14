package interpreter

import (
	"time"
)

func registerRuntimeCoreBuiltins() {
	builtins["exit"] = &Builtin{Name: "exit", Fn: builtinExit}
	builtins["fail"] = &Builtin{Name: "fail", Fn: builtinFail}
	builtins["rendezvous"] = &Builtin{Name: "rendezvous", Fn: builtinChannel}
	builtins["channel"] = &Builtin{Name: "channel", Fn: builtinChannel}
	builtins["buffered"] = &Builtin{Name: "buffered", Fn: builtinBufferedChannel}
	builtins["sleep"] = &Builtin{Name: "sleep", Fn: builtinSleep}
	builtins["log"] = &Builtin{Name: "log", Fn: builtinLog}
	builtins["str"] = &Builtin{Name: "str", Fn: builtinStr}
}

func runtimeFatalSignal(e *Evaluator) <-chan struct{} {
	if e == nil || e.runtime == nil {
		return nil
	}
	return e.runtime.fatalSignal()
}

func runtimeCancelSignal(e *Evaluator) <-chan struct{} {
	if e == nil || e.currentTask == nil {
		return nil
	}
	return e.currentTask.cancelCh
}

func runtimeFatalError(e *Evaluator) error {
	if e != nil && e.runtime != nil {
		if err := e.runtime.getFatalTaskFailure(); err != nil {
			return err
		}
	}
	return &RuntimeError{Message: "runtime terminated"}
}

func builtinExit(_ *Evaluator, args []Value) (Value, error) {
	msg := ""
	if len(args) > 0 {
		msg = args[0].Inspect()
	}
	exitProcess(msg)
	return nil, &ExitError{Message: msg}
}

func builtinFail(_ *Evaluator, args []Value) (Value, error) {
	if len(args) > 1 {
		return nil, &RuntimeError{Message: "fail expects 0 or 1 argument"}
	}
	msg := ""
	if len(args) == 1 {
		s, ok := args[0].(*String)
		if !ok {
			return nil, &RuntimeError{Message: "fail expects string message"}
		}
		msg = s.Value
	}
	return nil, recoverableError("fail", msg)
}

func builtinSleep(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sleep expects 1 argument"}
	}
	ms, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "sleep expects integer milliseconds"}
	}

	d := time.Duration(ms.Value) * time.Millisecond
	if d <= 0 {
		return UnitValue, nil
	}
	fatalCh := runtimeFatalSignal(e)
	cancelCh := runtimeCancelSignal(e)
	if cancelCh == nil && fatalCh == nil {
		time.Sleep(d)
		return UnitValue, nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return UnitValue, nil
	case <-cancelCh:
		return nil, canceledError()
	case <-fatalCh:
		return nil, runtimeFatalError(e)
	}
}
