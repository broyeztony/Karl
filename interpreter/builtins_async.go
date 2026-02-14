package interpreter

func registerAsyncBuiltins() {
	builtins["then"] = &Builtin{Name: "then", Fn: builtinThen}
	builtins["send"] = &Builtin{Name: "send", Fn: builtinSend}
	builtins["recv"] = &Builtin{Name: "recv", Fn: builtinRecv}
	builtins["done"] = &Builtin{Name: "done", Fn: builtinDone}
}

func builtinThen(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "then expects task and function"}
	}
	task, ok := args[0].(*Task)
	if !ok {
		return nil, &RuntimeError{Message: "then expects task as receiver"}
	}
	// Register observation at chaining time so fail-fast does not treat this task
	// as detached/unobserved while the continuation goroutine starts.
	task.markObserved()
	fn := args[1]
	thenTask := e.newTask(e.currentTask, false)
	thenEval := e.cloneForTask(thenTask)
	go func() {
		val, sig, err := taskAwaitWithCancel(task, thenTask.cancelCh, e.runtime)
		if err != nil {
			thenEval.handleAsyncError(thenTask, err)
			return
		}
		if sig != nil {
			exitProcess("break/continue outside loop")
			return
		}
		res, sig, err := thenEval.applyFunction(fn, []Value{val})
		if err != nil {
			thenEval.handleAsyncError(thenTask, err)
			return
		}
		if sig != nil {
			exitProcess("break/continue outside loop")
			return
		}
		thenTask.complete(res, nil)
	}()
	return thenTask, nil
}
