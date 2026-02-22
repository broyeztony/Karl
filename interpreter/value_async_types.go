package interpreter

import "sync"

type Task struct {
	ResultCh chan taskResult

	mu       sync.Mutex
	done     bool
	result   Value
	err      error
	observed bool

	internal bool

	// Cancellation is cooperative. A task only stops when it reaches a yield
	// point (wait/recv/sleep/http/...) where we check cancelCh.
	cancelOnce sync.Once
	cancelCh   chan struct{}

	// Bookkeeping for structured cancellation (parent cancels children).
	parent   *Task
	children []*Task

	// Used for formatting task errors (each task captures the file it was spawned from).
	source   string
	filename string
}

func (t *Task) Type() ValueType { return TASK }
func (t *Task) Inspect() string { return "<task>" }

type Channel struct {
	Ch        chan Value
	Closed    bool
	closeOnce sync.Once
	onClose   func()
}

func (c *Channel) Type() ValueType { return CHANNEL }
func (c *Channel) Inspect() string { return "<channel>" }
func (c *Channel) Close() {
	c.closeOnce.Do(func() {
		c.Closed = true
		if c.onClose != nil {
			c.onClose()
		}
		close(c.Ch)
	})
}

type taskResult struct {
	value Value
	err   error
}
