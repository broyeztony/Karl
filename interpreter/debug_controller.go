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
	stepTask    int
	pauseReq    bool
	terminate   bool
	paused      bool
	done        bool
	skipFile    string
	skipLine    int
	skipTask    int
	pausedTask  int

	current     DebugEvent
	stopped     DebugEvent
	stackByTask map[int][]DebugFrame

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
		stackByTask:   make(map[int][]DebugFrame),
	}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func normalizeTaskID(taskID int) int {
	if taskID <= 0 {
		return 1
	}
	return taskID
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
	c.stepTask = 0
	c.armSkipCurrentLocationLocked()
	c.paused = false
	c.pausedTask = 0
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *DebugController) Step() {
	c.mu.Lock()
	c.stepMode = debugStepIn
	c.stepDepth = c.stopped.FrameDepth
	// Step-in may cross into newly spawned task boundaries.
	c.stepTask = 0
	c.armSkipCurrentLocationLocked()
	c.paused = false
	c.pausedTask = 0
	c.cond.Broadcast()
	c.mu.Unlock()
}

// BindPendingStepInTask locks a pending step-in to a specific task. This is
// used for deterministic stepping into spawned expressions.
func (c *DebugController) BindPendingStepInTask(taskID int) {
	if taskID <= 0 {
		return
	}
	c.mu.Lock()
	if c.stepMode == debugStepIn && c.stepTask == 0 {
		c.stepTask = taskID
	}
	c.mu.Unlock()
}

func (c *DebugController) StepOver() {
	c.mu.Lock()
	c.stepMode = debugStepOver
	c.stepDepth = c.stopped.FrameDepth
	c.stepTask = normalizeTaskID(c.stopped.TaskID)
	c.armSkipCurrentLocationLocked()
	c.paused = false
	c.pausedTask = 0
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *DebugController) StepOut() {
	c.mu.Lock()
	c.stepMode = debugStepOut
	c.stepDepth = c.stopped.FrameDepth
	c.stepTask = normalizeTaskID(c.stopped.TaskID)
	c.armSkipCurrentLocationLocked()
	c.paused = false
	c.pausedTask = 0
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
	c.pausedTask = 0
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *DebugController) WaitForStop() DebugStopReason {
	for {
		select {
		case <-c.pauseCh:
			c.mu.Lock()
			done := c.done
			paused := c.paused
			c.mu.Unlock()
			if done {
				return DebugStopDone
			}
			if paused {
				return DebugStopPaused
			}
		case <-c.doneCh:
			return DebugStopDone
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
	paused := c.paused
	event := c.current
	if paused {
		event = c.stopped
	}
	c.mu.Unlock()
	return event, paused
}

func (c *DebugController) Stack() []DebugFrame {
	c.mu.Lock()
	stack := c.activeStackLocked()
	out := make([]DebugFrame, len(stack))
	copy(out, stack)
	c.mu.Unlock()
	return out
}

func (c *DebugController) Locals() map[string]Value {
	c.mu.Lock()
	env := c.activeEnvLocked()
	c.mu.Unlock()
	if env == nil {
		return map[string]Value{}
	}
	return env.Snapshot()
}

func (c *DebugController) CurrentEnv() *Environment {
	c.mu.Lock()
	env := c.activeEnvLocked()
	c.mu.Unlock()
	return env
}

func (c *DebugController) EnvForFrame(displayIndex int) (*Environment, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if displayIndex < 0 {
		return nil, fmt.Errorf("frame index must be >= 0")
	}
	// For selected top frame, expose current lexical environment to include
	// block-local bindings within the frame.
	if displayIndex == 0 {
		if env := c.activeEventLocked().Env; env != nil {
			return env, nil
		}
	}
	stack := c.activeStackLocked()
	if len(stack) == 0 {
		if displayIndex != 0 {
			return nil, fmt.Errorf("frame #%d out of range", displayIndex)
		}
		env := c.activeEventLocked().Env
		if env == nil {
			return nil, fmt.Errorf("no debug environment available")
		}
		return env, nil
	}
	if displayIndex >= len(stack) {
		return nil, fmt.Errorf("frame #%d out of range", displayIndex)
	}
	idx := len(stack) - 1 - displayIndex
	env := stack[idx].Env
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
	c.pausedTask = 0
	c.cond.Broadcast()
	c.mu.Unlock()

	c.doneOnce.Do(func() {
		close(c.doneCh)
	})
}

func (c *DebugController) BeforeNode(event DebugEvent) error {
	c.mu.Lock()
	event.TaskID = normalizeTaskID(event.TaskID)

	if c.terminate {
		c.mu.Unlock()
		return &DebugTerminatedError{}
	}

	for c.paused && c.pausedTask != event.TaskID && !c.done && !c.terminate {
		c.cond.Wait()
	}
	if c.terminate {
		c.mu.Unlock()
		return &DebugTerminatedError{}
	}
	if c.done {
		c.mu.Unlock()
		return nil
	}

	c.current = event

	if c.shouldPauseLocked(event) {
		c.paused = true
		c.pausedTask = event.TaskID
		c.stopped = event
		select {
		case c.pauseCh <- struct{}{}:
		default:
		}

		for c.paused && !c.done && !c.terminate {
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
	taskID := normalizeTaskID(frame.TaskID)
	frame.TaskID = taskID
	c.stackByTask[taskID] = append(c.stackByTask[taskID], frame)
	c.mu.Unlock()
}

func (c *DebugController) OnFramePop(frame DebugFrame) {
	c.mu.Lock()
	taskID := normalizeTaskID(frame.TaskID)
	stack := c.stackByTask[taskID]
	if n := len(stack); n > 0 {
		c.stackByTask[taskID] = stack[:n-1]
		if len(c.stackByTask[taskID]) == 0 {
			delete(c.stackByTask, taskID)
		}
	}
	c.mu.Unlock()
}

func (c *DebugController) activeEventLocked() DebugEvent {
	if c.paused {
		return c.stopped
	}
	return c.current
}

func (c *DebugController) activeStackLocked() []DebugFrame {
	event := c.activeEventLocked()
	taskID := normalizeTaskID(event.TaskID)
	if c.paused && c.pausedTask > 0 {
		taskID = c.pausedTask
	}
	return c.stackByTask[taskID]
}

func (c *DebugController) activeEnvLocked() *Environment {
	// Prefer the current lexical scope (event env) so locals from nested blocks
	// (for/if/match bodies) are visible while stepping.
	if env := c.activeEventLocked().Env; env != nil {
		return env
	}
	stack := c.activeStackLocked()
	if len(stack) > 0 {
		return stack[len(stack)-1].Env
	}
	return nil
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
		taskID := normalizeTaskID(event.TaskID)
		if c.skipTask == 0 || taskID == c.skipTask {
			if event.Filename == c.skipFile && event.Line == c.skipLine {
				return false
			}
			// We left the skipped location on the same task, re-enable breakpoint checks.
			c.skipFile = ""
			c.skipLine = 0
			c.skipTask = 0
		}
	}
	if c.shouldPauseForStepLocked(event) {
		return true
	}
	if c.stepMode != debugStepNone && c.stepTask > 0 && normalizeTaskID(event.TaskID) != c.stepTask {
		// While a step operation is in progress, only pause the stepped task.
		// This avoids unrelated concurrent tasks stealing the next stop.
		return false
	}
	if lines := c.breakpoints[event.Filename]; lines != nil {
		_, ok := lines[event.Line]
		return ok
	}
	return false
}

func (c *DebugController) shouldPauseForStepLocked(event DebugEvent) bool {
	if c.stepTask > 0 && normalizeTaskID(event.TaskID) != c.stepTask {
		return false
	}
	switch c.stepMode {
	case debugStepNone:
		return false
	case debugStepIn:
		c.stepMode = debugStepNone
		c.stepDepth = 0
		c.stepTask = 0
		return true
	case debugStepOver:
		if event.FrameDepth <= c.stepDepth {
			c.stepMode = debugStepNone
			c.stepDepth = 0
			c.stepTask = 0
			return true
		}
		return false
	case debugStepOut:
		if c.stepDepth <= 0 {
			c.stepMode = debugStepNone
			c.stepDepth = 0
			c.stepTask = 0
			return false
		}
		if event.FrameDepth < c.stepDepth {
			c.stepMode = debugStepNone
			c.stepDepth = 0
			c.stepTask = 0
			return true
		}
		return false
	default:
		return false
	}
}

func (c *DebugController) armSkipCurrentLocationLocked() {
	ev := c.activeEventLocked()
	if c.paused && ev.Line > 0 {
		c.skipFile = ev.Filename
		c.skipLine = ev.Line
		c.skipTask = normalizeTaskID(ev.TaskID)
	}
}
