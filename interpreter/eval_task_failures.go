package interpreter

import "sort"

func (e *Evaluator) handleAsyncError(task *Task, err error) {
	if err == nil {
		return
	}
	if exitErr, ok := err.(*ExitError); ok {
		exitProcess(exitErr.Message)
		return
	}
	if task == nil {
		return
	}
	task.complete(nil, err)
	if e.runtime == nil {
		return
	}
	if e.runtime.getTaskFailurePolicy() != TaskFailurePolicyFailFast {
		return
	}
	if task.internal || task.isObserved() {
		return
	}
	formatted := FormatRuntimeError(err, task.source, task.filename)
	e.runtime.setFatalTaskFailure(&UnhandledTaskError{Messages: []string{formatted}})
}

func (e *Evaluator) CheckUnhandledTaskFailures() error {
	if e.runtime == nil {
		return nil
	}
	tasks := e.runtime.snapshotTasks()
	msgs := []string{}
	for _, t := range tasks {
		if t == nil || t.internal {
			continue
		}
		if !t.isDone() || t.isObserved() {
			continue
		}
		err := t.getError()
		if err == nil {
			continue
		}
		if re, ok := err.(*RecoverableError); ok && re.Kind == "canceled" {
			continue
		}
		msgs = append(msgs, FormatRuntimeError(err, t.source, t.filename))
	}
	if len(msgs) == 0 {
		return nil
	}
	sort.Strings(msgs)
	return &UnhandledTaskError{Messages: msgs}
}
