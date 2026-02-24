package interpreter

import (
	"fmt"
	"karl/ast"
)

// DebugEvent describes a single evaluator step.
type DebugEvent struct {
	Filename   string
	Line       int
	Column     int
	NodeType   string
	FrameDepth int
	TaskID     int
	Env        *Environment
}

// DebugFrame represents a call frame tracked by the evaluator.
type DebugFrame struct {
	ID       int
	Name     string
	Filename string
	Line     int
	Column   int
	Depth    int
	TaskID   int
	Env      *Environment
}

// Debugger receives callbacks for evaluator execution.
type Debugger interface {
	BeforeNode(event DebugEvent) error
	AfterNode(event DebugEvent, result Value, signal *Signal, evalErr error) error
}

// FrameAwareDebugger receives frame push/pop callbacks.
type FrameAwareDebugger interface {
	OnFramePush(frame DebugFrame)
	OnFramePop(frame DebugFrame)
}

func (e *Evaluator) debugBeforeNode(node ast.Node, env *Environment) error {
	if e.debugger == nil {
		return nil
	}
	return e.debugger.BeforeNode(e.debugEvent(node, env))
}

func (e *Evaluator) debugAfterNode(node ast.Node, env *Environment, result Value, signal *Signal, evalErr error) error {
	if e.debugger == nil {
		return nil
	}
	return e.debugger.AfterNode(e.debugEvent(node, env), result, signal, evalErr)
}

func (e *Evaluator) debugEvent(node ast.Node, env *Environment) DebugEvent {
	event := DebugEvent{
		Filename:   e.filename,
		NodeType:   fmt.Sprintf("%T", node),
		FrameDepth: len(e.debugFrames),
		TaskID:     e.debugTaskID(),
		Env:        env,
	}
	if tok := tokenFromNode(node); tok != nil {
		event.Line = tok.Line
		event.Column = tok.Column
	}
	return event
}

func (e *Evaluator) pushDebugFrame(name string, node ast.Node, env *Environment) {
	if e.debugger == nil {
		return
	}
	frame := DebugFrame{
		ID:       e.debugFrameSeq + 1,
		Name:     name,
		Filename: e.filename,
		Depth:    len(e.debugFrames) + 1,
		TaskID:   e.debugTaskID(),
		Env:      env,
	}
	e.debugFrameSeq++
	if tok := tokenFromNode(node); tok != nil {
		frame.Line = tok.Line
		frame.Column = tok.Column
	}
	e.debugFrames = append(e.debugFrames, frame)
	if d, ok := e.debugger.(FrameAwareDebugger); ok {
		d.OnFramePush(frame)
	}
}

func (e *Evaluator) debugTaskID() int {
	if e.currentTask != nil && e.currentTask.debugID > 0 {
		return e.currentTask.debugID
	}
	return 1
}

func (e *Evaluator) popDebugFrame() {
	if e.debugger == nil || len(e.debugFrames) == 0 {
		return
	}
	last := e.debugFrames[len(e.debugFrames)-1]
	e.debugFrames = e.debugFrames[:len(e.debugFrames)-1]
	if d, ok := e.debugger.(FrameAwareDebugger); ok {
		d.OnFramePop(last)
	}
}
