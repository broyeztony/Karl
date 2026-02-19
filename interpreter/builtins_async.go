package interpreter

func registerAsyncBuiltins() {
	builtins["then"] = &Builtin{Name: "then", Fn: builtinThen}
	builtins["send"] = &Builtin{Name: "send", Fn: builtinSend}
	builtins["recv"] = &Builtin{Name: "recv", Fn: builtinRecv}
	builtins["done"] = &Builtin{Name: "done", Fn: builtinDone}
	builtins["spawn"] = &Builtin{Name: "spawn", Fn: builtinSpawn}
}

func builtinSpawn(e *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, &RuntimeError{Message: "spawn expects at least 1 argument (function)"}
	}

	fn := args[0]
	
	// Internal helper to spawn the task
	spawnTask := func(targetFn Value, callArgs []Value) (Value, error) {
		task := e.newTask(e.currentTask, false)
		taskEval := e.cloneForTask(task)
		go func() {
			res, sig, err := taskEval.applyFunction(targetFn, callArgs)
			if err != nil {
				taskEval.handleAsyncError(task, err)
				return
			}
			if sig != nil {
				exitProcess("break/continue outside loop")
				return
			}
			task.complete(res, nil)
		}()
		return task, nil
	}

	// If more arguments are provided, spawn immediately with those arguments
	if len(args) > 1 {
		return spawnTask(fn, args[1:])
	}
	
	// If only 1 argument, return a wrapper function (curried style)
	return &Builtin{
		Name: "spawn_wrapper",
		Fn: func(_ *Evaluator, spawnArgs []Value) (Value, error) {
			return spawnTask(fn, spawnArgs)
		},
	}, nil
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
