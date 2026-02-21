package interpreter

import "karl/ast"

func (e *Evaluator) evalRaceExpression(node *ast.RaceExpression, env *Environment) (Value, *Signal, error) {
	raceTask := e.newTask(e.currentTask, false)

	children := make([]*Task, 0, len(node.Tasks))
	for _, taskExpr := range node.Tasks {
		child, err := e.spawnTask(taskExpr, env, raceTask, true)
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
			value Value
			sig   *Signal
			err   error
		}
		results := make(chan result, len(children))
		for _, child := range children {
			go func(t *Task) {
				val, sig, err := taskAwaitWithCancel(t, raceTask.cancelCh, e.runtime)
				results <- result{value: val, sig: sig, err: err}
			}(child)
		}

		select {
		case <-raceTask.cancelCh:
			// canceled by user or parent; Cancel() already completed the task.
			return
		case first := <-results:
			// Cancel losers. Cancellation is cooperative; losers will stop at the next yield point.
			raceTask.cancelChildren()
			if first.sig != nil {
				raceTask.complete(nil, &RuntimeError{Message: "break/continue outside loop"})
				return
			}
			raceTask.complete(first.value, first.err)
			return
		}
	}()

	return raceTask, nil, nil
}
