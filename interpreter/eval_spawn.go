package interpreter

import "karl/ast"

func (e *Evaluator) evalAwaitExpression(node *ast.AwaitExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	task, ok := val.(*Task)
	if !ok {
		return nil, nil, &RuntimeError{Message: "wait expects task"}
	}
	var cancelCh <-chan struct{}
	if e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}
	return taskAwaitWithCancel(task, cancelCh, e.runtime)
}

func (e *Evaluator) evalSpawnExpression(node *ast.SpawnExpression, env *Environment) (Value, *Signal, error) {
	if node.Task != nil {
		task, err := e.spawnTask(node.Task, env, e.currentTask, false)
		if err != nil {
			return nil, nil, err
		}
		e.bindPendingStepInTask(task)
		return task, nil, nil
	}

	join := e.newTask(e.currentTask, false)

	children := make([]*Task, 0, len(node.Group))
	for _, expr := range node.Group {
		child, err := e.spawnTask(expr, env, join, true)
		if err != nil {
			return nil, nil, err
		}
		if len(children) == 0 {
			e.bindPendingStepInTask(child)
		}
		children = append(children, child)
	}

	go func() {
		type result struct {
			idx   int
			value Value
			sig   *Signal
			err   error
		}

		resultsCh := make(chan result, len(children))
		for i, child := range children {
			go func(idx int, t *Task) {
				val, sig, err := taskAwaitWithCancel(t, join.cancelCh, e.runtime)
				resultsCh <- result{idx: idx, value: val, sig: sig, err: err}
			}(i, child)
		}

		out := make([]Value, len(children))
		remaining := len(children)
		for remaining > 0 {
			select {
			case <-join.cancelCh:
				// canceled by user or parent; Cancel() already completed the task.
				return
			case r := <-resultsCh:
				if r.err != nil {
					// Fail fast: cancel remaining children and surface the error on the join task.
					join.cancelChildren()
					join.complete(nil, r.err)
					return
				}
				if r.sig != nil {
					join.cancelChildren()
					join.complete(nil, &RuntimeError{Message: "break/continue outside loop"})
					return
				}
				out[r.idx] = r.value
				remaining--
			}
		}

		join.complete(&Array{Elements: out}, nil)
	}()

	return join, nil, nil
}

func (e *Evaluator) spawnTask(expr ast.Expression, env *Environment, parent *Task, internal bool) (*Task, error) {
	task := e.newTask(parent, internal)
	taskEval := e.cloneForTask(task)
	go func() {
		val, sig, err := taskEval.Eval(expr, env)
		if err != nil {
			taskEval.handleAsyncError(task, err)
			return
		}
		if sig != nil {
			task.complete(nil, &RuntimeError{Message: "break/continue outside loop"})
			return
		}
		task.complete(val, nil)
	}()
	return task, nil
}

func (e *Evaluator) bindPendingStepInTask(task *Task) {
	if task == nil || e.debugger == nil {
		return
	}
	controller, ok := e.debugger.(*DebugController)
	if !ok {
		return
	}
	controller.BindPendingStepInTask(task.debugID)
}
