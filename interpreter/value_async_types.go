package interpreter

import "sync"

type Task struct {
	// Debug thread identifier used by the debugger/DAP bridge.
	debugID int

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
	closeCh   chan struct{}
	Closed    bool
	closedMu  sync.RWMutex
	closeOnce sync.Once
	onClose   func()
	// Some channels are fed by external event sources (OS signals, etc.).
	// Top-level recv on those channels is not a deadlock even with no Karl tasks.
	allowTopLevelBlock bool
}

func (c *Channel) Type() ValueType { return CHANNEL }
func (c *Channel) Inspect() string { return "<channel>" }
func (c *Channel) Close() {
	c.closeOnce.Do(func() {
		c.closedMu.Lock()
		c.Closed = true
		c.closedMu.Unlock()
		if c.onClose != nil {
			c.onClose()
		}
		if c.closeCh != nil {
			close(c.closeCh)
		}
	})
}

func (c *Channel) IsClosed() bool {
	if c == nil {
		return true
	}
	c.closedMu.RLock()
	closed := c.Closed
	c.closedMu.RUnlock()
	return closed
}

func (c *Channel) ClosedSignal() <-chan struct{} {
	if c == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return c.closeCh
}

type taskResult struct {
	value Value
	err   error
}
