package interpreter

import (
	"os"
)

func newTask() *Task {
	return &Task{ResultCh: make(chan taskResult, 1)}
}

func (t *Task) complete(value Value, err error) {
	if t.done {
		return
	}
	t.done = true
	t.result = value
	t.err = err
	t.ResultCh <- taskResult{value: value, err: err}
}

func taskAwait(t *Task) (Value, *Signal, error) {
	if t.done {
		if t.err != nil {
			return nil, nil, t.err
		}
		return t.result, nil, nil
	}
	res := <-t.ResultCh
	t.done = true
	t.result = res.value
	t.err = res.err
	if res.err != nil {
		return nil, nil, res.err
	}
	return res.value, nil, nil
}

func exitProcess(msg string) {
	if msg != "" {
		_, _ = os.Stderr.WriteString(msg + "\n")
	}
	os.Exit(1)
}
