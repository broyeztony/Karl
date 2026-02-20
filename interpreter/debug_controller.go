package interpreter

import (
	"fmt"
	"strings"
	"sync"
)

type DebugStopReason int

const (
	DebugStopPaused DebugStopReason = iota + 1
	DebugStopDone
)

type debugStepMode int

const (
	debugStepNone debugStepMode = iota
	debugStepIn
	debugStepOver
	debugStepOut
)

type DebugTerminatedError struct{}

func (e *DebugTerminatedError) Error() string {
	return "debug session terminated"
}

func IsDebugTerminated(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*DebugTerminatedError)
	return ok
}

type DebugBreakpoint struct {
	ID   int
	File string
	Line int
}

// DebugController implements the evaluator Debugger interface and provides
// pause/continue/step controls for CLI debugging.
type DebugController struct {
	mu sync.Mutex

	cond *sync.Cond

	defaultFile   string
	breakpoints   map[string]map[int]int
	breakpointIDs map[int]DebugBreakpoint
	nextBpID      int

	startPaused bool
	stepMode    debugStepMode
	stepDepth   int
	pauseReq    bool
	terminate   bool
	paused      bool
	done        bool
	skipFile    string
	skipLine    int

	current DebugEvent
	stack   []DebugFrame

	result Value
	err    error

	pauseCh  chan struct{}
	doneCh   chan struct{}
	doneOnce sync.Once
}

func NewDebugController(defaultFile string) *DebugController {
	c := &DebugController{
		defaultFile:   defaultFile,
		breakpoints:   make(map[string]map[int]int),
		breakpointIDs: make(map[int]DebugBreakpoint),
		startPaused:   true,
		pauseCh:       make(chan struct{}, 1),
		doneCh:        make(chan struct{}),
	}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *DebugController) SetStartPaused(paused bool) {
	c.mu.Lock()
	c.startPaused = paused
	c.mu.Unlock()
}

func (c *DebugController) AddBreakpoint(file string, line int) (int, error) {
	if line <= 0 {
		return 0, fmt.Errorf("line must be > 0")
	}
	if strings.TrimSpace(file) == "" {
		file = c.defaultFile
	}
	if strings.TrimSpace(file) == "" {
		return 0, fmt.Errorf("file is required for breakpoint")
	}

	c.mu.Lock()
	lines := c.breakpoints[file]
	if lines == nil {
		lines = make(map[int]int)
		c.breakpoints[file] = lines
	}
	if id, ok := lines[line]; ok {
		c.mu.Unlock()
		return id, nil
	}
	c.nextBpID++
	id := c.nextBpID
	lines[line] = id
	c.breakpointIDs[id] = DebugBreakpoint{ID: id, File: file, Line: line}
	c.mu.Unlock()
	return id, nil
}

func (c *DebugController) RemoveBreakpoint(id int) error {
	if id <= 0 {
		return fmt.Errorf("breakpoint id must be > 0")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	bp, ok := c.breakpointIDs[id]
	if !ok {
		return fmt.Errorf("breakpoint #%d not found", id)
	}
	delete(c.breakpointIDs, id)
	if lines := c.breakpoints[bp.File]; lines != nil {
		delete(lines, bp.Line)
		if len(lines) == 0 {
			delete(c.breakpoints, bp.File)
		}
	}
	return nil
}

func (c *DebugController) Breakpoints() []DebugBreakpoint {
	c.mu.Lock()
	out := make([]DebugBreakpoint, 0, len(c.breakpointIDs))
	for _, bp := range c.breakpointIDs {
		out = append(out, bp)
	}
	c.mu.Unlock()
	return out
}

func (c *DebugController) ClearBreakpoints() int {
	c.mu.Lock()
	count := len(c.breakpointIDs)
	c.breakpoints = make(map[string]map[int]int)
	c.breakpointIDs = make(map[int]DebugBreakpoint)
	c.mu.Unlock()
	return count
}

func (c *DebugController) Continue() {
	c.mu.Lock()
	c.stepMode = debugStepNone
	c.stepDepth = 0
	c.armSkipCurrentLocationLocked()
	c.paused = false
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *DebugController) Step() {
	c.mu.Lock()
	c.stepMode = debugStepIn
	c.stepDepth = c.current.FrameDepth
	c.armSkipCurrentLocationLocked()
	c.paused = false
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *DebugController) StepOver() {
	c.mu.Lock()
	c.stepMode = debugStepOver
	c.stepDepth = c.current.FrameDepth
	c.armSkipCurrentLocationLocked()
	c.paused = false
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *DebugController) StepOut() {
	c.mu.Lock()
	c.stepMode = debugStepOut
	c.stepDepth = c.current.FrameDepth
	c.armSkipCurrentLocationLocked()
	c.paused = false
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *DebugController) Pause() {
	c.mu.Lock()
	c.pauseReq = true
	c.mu.Unlock()
}

func (c *DebugController) Quit() {
	c.mu.Lock()
	c.terminate = true
	c.paused = false
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *DebugController) WaitForStop() DebugStopReason {
	for {
		c.mu.Lock()
		if c.done {
			c.mu.Unlock()
			return DebugStopDone
		}
		if c.paused {
			c.mu.Unlock()
			return DebugStopPaused
		}
		c.mu.Unlock()

		select {
		case <-c.pauseCh:
		case <-c.doneCh:
		}
	}
}

func (c *DebugController) IsPaused() bool {
	c.mu.Lock()
	paused := c.paused
	c.mu.Unlock()
	return paused
}

func (c *DebugController) CurrentEvent() (DebugEvent, bool) {
	c.mu.Lock()
	event := c.current
	paused := c.paused
	c.mu.Unlock()
	return event, paused
}

func (c *DebugController) Stack() []DebugFrame {
	c.mu.Lock()
	out := make([]DebugFrame, len(c.stack))
	copy(out, c.stack)
	c.mu.Unlock()
	return out
}

func (c *DebugController) Locals() map[string]Value {
	c.mu.Lock()
	var env *Environment
	if len(c.stack) > 0 {
		env = c.stack[len(c.stack)-1].Env
	} else {
		env = c.current.Env
	}
	c.mu.Unlock()
	if env == nil {
		return map[string]Value{}
	}
	return env.Snapshot()
}

func (c *DebugController) CurrentEnv() *Environment {
	c.mu.Lock()
	var env *Environment
	if len(c.stack) > 0 {
		env = c.stack[len(c.stack)-1].Env
	} else {
		env = c.current.Env
	}
	c.mu.Unlock()
	return env
}

func (c *DebugController) EnvForFrame(displayIndex int) (*Environment, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if displayIndex < 0 {
		return nil, fmt.Errorf("frame index must be >= 0")
	}
	if len(c.stack) == 0 {
		if displayIndex != 0 {
			return nil, fmt.Errorf("frame #%d out of range", displayIndex)
		}
		if c.current.Env == nil {
			return nil, fmt.Errorf("no debug environment available")
		}
		return c.current.Env, nil
	}
	if displayIndex >= len(c.stack) {
		return nil, fmt.Errorf("frame #%d out of range", displayIndex)
	}
	idx := len(c.stack) - 1 - displayIndex
	env := c.stack[idx].Env
	if env == nil {
		return nil, fmt.Errorf("frame #%d has no environment", displayIndex)
	}
	return env, nil
}

func (c *DebugController) Result() (Value, error, bool) {
	c.mu.Lock()
	result := c.result
	err := c.err
	done := c.done
	c.mu.Unlock()
	return result, err, done
}

func (c *DebugController) Finish(result Value, err error) {
	c.mu.Lock()
	c.result = result
	c.err = err
	c.done = true
	c.paused = false
	c.cond.Broadcast()
	c.mu.Unlock()

	c.doneOnce.Do(func() {
		close(c.doneCh)
	})
}

func (c *DebugController) BeforeNode(event DebugEvent) error {
	c.mu.Lock()
	c.current = event

	if c.terminate {
		c.mu.Unlock()
		return &DebugTerminatedError{}
	}

	if c.shouldPauseLocked(event) {
		c.paused = true
		select {
		case c.pauseCh <- struct{}{}:
		default:
		}

		for c.paused && !c.done {
			c.cond.Wait()
		}
		if c.terminate {
			c.mu.Unlock()
			return &DebugTerminatedError{}
		}
	}

	c.mu.Unlock()
	return nil
}

func (c *DebugController) AfterNode(_ DebugEvent, _ Value, _ *Signal, _ error) error {
	return nil
}

func (c *DebugController) OnFramePush(frame DebugFrame) {
	c.mu.Lock()
	c.stack = append(c.stack, frame)
	c.mu.Unlock()
}

func (c *DebugController) OnFramePop(frame DebugFrame) {
	c.mu.Lock()
	if n := len(c.stack); n > 0 {
		c.stack = c.stack[:n-1]
	}
	c.mu.Unlock()
}

func (c *DebugController) shouldPauseLocked(event DebugEvent) bool {
	if c.pauseReq {
		c.pauseReq = false
		return true
	}
	if c.startPaused && event.Line > 0 {
		c.startPaused = false
		return true
	}
	if event.Line <= 0 {
		return false
	}
	if c.skipLine > 0 {
		if event.Filename == c.skipFile && event.Line == c.skipLine {
			return false
		}
		c.skipFile = ""
		c.skipLine = 0
	}
	if c.shouldPauseForStepLocked(event) {
		return true
	}
	if lines := c.breakpoints[event.Filename]; lines != nil {
		_, ok := lines[event.Line]
		return ok
	}
	return false
}

func (c *DebugController) shouldPauseForStepLocked(event DebugEvent) bool {
	switch c.stepMode {
	case debugStepNone:
		return false
	case debugStepIn:
		c.stepMode = debugStepNone
		c.stepDepth = 0
		return true
	case debugStepOver:
		if event.FrameDepth <= c.stepDepth {
			c.stepMode = debugStepNone
			c.stepDepth = 0
			return true
		}
		return false
	case debugStepOut:
		if c.stepDepth <= 0 {
			c.stepMode = debugStepNone
			c.stepDepth = 0
			return false
		}
		if event.FrameDepth < c.stepDepth {
			c.stepMode = debugStepNone
			c.stepDepth = 0
			return true
		}
		return false
	default:
		return false
	}
}

func (c *DebugController) armSkipCurrentLocationLocked() {
	if c.paused && c.current.Line > 0 {
		c.skipFile = c.current.Filename
		c.skipLine = c.current.Line
	}
}
