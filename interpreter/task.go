package interpreter

import (
	"os"
)

func (t *Task) addChild(child *Task) {
	if t == nil || child == nil {
		return
	}
	t.mu.Lock()
	t.children = append(t.children, child)
	t.mu.Unlock()
}

func (t *Task) cancelChildren() {
	t.mu.Lock()
	children := append([]*Task(nil), t.children...)
	t.mu.Unlock()
	for _, child := range children {
		child.Cancel()
	}
}

func (t *Task) Cancel() {
	if t == nil {
		return
	}
	t.cancelOnce.Do(func() { close(t.cancelCh) })
	t.cancelChildren()
	t.complete(nil, canceledError())
}

func (t *Task) canceled() bool {
	if t == nil {
		return false
	}
	select {
	case <-t.cancelCh:
		return true
	default:
		return false
	}
}

func exitProcess(msg string) {
	if msg != "" {
		_, _ = os.Stderr.WriteString(msg + "\n")
	}
	os.Exit(1)
}

func canceledError() *RecoverableError {
	return &RecoverableError{Kind: "canceled", Message: "task canceled"}
}
