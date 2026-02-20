package tests

import (
	"testing"
	"time"

	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

func TestDebugControllerBreakpointPause(t *testing.T) {
	source := "let x = 1\nlet y = 2\nx + y\n"
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, "test.k")
	controller := interpreter.NewDebugController("test.k")
	controller.SetStartPaused(false)
	if _, err := controller.AddBreakpoint("test.k", 2); err != nil {
		t.Fatalf("add breakpoint: %v", err)
	}
	eval.SetDebugger(controller)

	env := interpreter.NewBaseEnvironment()
	done := make(chan struct{})
	go func() {
		val, sig, err := eval.Eval(program, env)
		if err == nil && sig != nil {
			err = &interpreter.RuntimeError{Message: "unexpected signal"}
		}
		controller.Finish(val, err)
		close(done)
	}()

	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopPaused {
		t.Fatalf("expected paused stop, got %v", reason)
	}
	event, ok := controller.CurrentEvent()
	if !ok {
		t.Fatalf("expected paused event")
	}
	if event.Line != 2 {
		t.Fatalf("expected pause on line 2, got %d", event.Line)
	}

	controller.Continue()
	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopDone {
		t.Fatalf("expected done stop, got %v", reason)
	}
	<-done

	val, err, doneState := controller.Result()
	if !doneState {
		t.Fatalf("expected done state")
	}
	if err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
	assertInteger(t, val, 3)
}

func TestDebugControllerStepPausesAgain(t *testing.T) {
	source := "let x = 1\nx\n"
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, "test.k")
	controller := interpreter.NewDebugController("test.k")
	eval.SetDebugger(controller)

	env := interpreter.NewBaseEnvironment()
	done := make(chan struct{})
	go func() {
		val, sig, err := eval.Eval(program, env)
		if err == nil && sig != nil {
			err = &interpreter.RuntimeError{Message: "unexpected signal"}
		}
		controller.Finish(val, err)
		close(done)
	}()

	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopPaused {
		t.Fatalf("expected initial paused stop, got %v", reason)
	}
	controller.Step()
	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopPaused {
		t.Fatalf("expected paused stop after step, got %v", reason)
	}

	controller.Continue()
	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopDone {
		t.Fatalf("expected done stop, got %v", reason)
	}
	<-done

	val, err, doneState := controller.Result()
	if !doneState {
		t.Fatalf("expected done state")
	}
	if err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
	assertInteger(t, val, 1)
}

func TestDebugControllerStepOverSkipsFunctionFrame(t *testing.T) {
	source := "let inc = x -> x + 1\nlet a = inc(1)\nlet b = a + 1\nb\n"
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, "test.k")
	controller := interpreter.NewDebugController("test.k")
	controller.SetStartPaused(false)
	if _, err := controller.AddBreakpoint("test.k", 2); err != nil {
		t.Fatalf("add breakpoint: %v", err)
	}
	eval.SetDebugger(controller)

	env := interpreter.NewBaseEnvironment()
	done := make(chan struct{})
	go func() {
		val, sig, err := eval.Eval(program, env)
		if err == nil && sig != nil {
			err = &interpreter.RuntimeError{Message: "unexpected signal"}
		}
		controller.Finish(val, err)
		close(done)
	}()

	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopPaused {
		t.Fatalf("expected paused stop at breakpoint, got %v", reason)
	}
	event, ok := controller.CurrentEvent()
	if !ok || event.Line != 2 {
		t.Fatalf("expected pause on line 2, got %+v", event)
	}

	controller.StepOver()
	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopPaused {
		t.Fatalf("expected paused stop after step over, got %v", reason)
	}
	event, ok = controller.CurrentEvent()
	if !ok {
		t.Fatalf("expected paused event")
	}
	if event.Line != 3 {
		t.Fatalf("expected step over to pause on line 3, got %d", event.Line)
	}
	if event.FrameDepth != 0 {
		t.Fatalf("expected frame depth 0 after step over, got %d", event.FrameDepth)
	}

	controller.Continue()
	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopDone {
		t.Fatalf("expected done stop, got %v", reason)
	}
	<-done

	val, err, doneState := controller.Result()
	if !doneState {
		t.Fatalf("expected done state")
	}
	if err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
	assertInteger(t, val, 3)
}

func TestDebugControllerStepOutReturnsToCaller(t *testing.T) {
	source := "let twicePlus = x -> {\n  let y = x + 1\n  y * 2\n}\nlet out = twicePlus(2)\nout + 1\n"
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, "test.k")
	controller := interpreter.NewDebugController("test.k")
	controller.SetStartPaused(false)
	if _, err := controller.AddBreakpoint("test.k", 5); err != nil {
		t.Fatalf("add breakpoint: %v", err)
	}
	if _, err := controller.AddBreakpoint("test.k", 2); err != nil {
		t.Fatalf("add breakpoint: %v", err)
	}
	eval.SetDebugger(controller)

	env := interpreter.NewBaseEnvironment()
	done := make(chan struct{})
	go func() {
		val, sig, err := eval.Eval(program, env)
		if err == nil && sig != nil {
			err = &interpreter.RuntimeError{Message: "unexpected signal"}
		}
		controller.Finish(val, err)
		close(done)
	}()

	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopPaused {
		t.Fatalf("expected paused stop at call site, got %v", reason)
	}
	event, ok := controller.CurrentEvent()
	if !ok || event.Line != 5 {
		t.Fatalf("expected pause on line 5, got %+v", event)
	}

	controller.Continue()
	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopPaused {
		t.Fatalf("expected paused stop in callee, got %v", reason)
	}
	event, ok = controller.CurrentEvent()
	if !ok || event.Line != 2 {
		t.Fatalf("expected pause on line 2, got %+v", event)
	}
	if event.FrameDepth < 1 {
		t.Fatalf("expected frame depth >= 1 in callee, got %d", event.FrameDepth)
	}

	controller.StepOut()
	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopPaused {
		t.Fatalf("expected paused stop after step out, got %v", reason)
	}
	event, ok = controller.CurrentEvent()
	if !ok {
		t.Fatalf("expected paused event")
	}
	if event.FrameDepth != 0 {
		t.Fatalf("expected frame depth 0 after step out, got %d", event.FrameDepth)
	}
	if event.Line < 5 {
		t.Fatalf("expected pause in caller after step out, got line %d", event.Line)
	}

	controller.Continue()
	if reason := waitDebugStop(t, controller); reason != interpreter.DebugStopDone {
		t.Fatalf("expected done stop, got %v", reason)
	}
	<-done

	val, err, doneState := controller.Result()
	if !doneState {
		t.Fatalf("expected done state")
	}
	if err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
	assertInteger(t, val, 7)
}

func TestDebugControllerBreakpointIDsAndDelete(t *testing.T) {
	controller := interpreter.NewDebugController("test.k")
	id1, err := controller.AddBreakpoint("test.k", 2)
	if err != nil {
		t.Fatalf("add breakpoint #1: %v", err)
	}
	id2, err := controller.AddBreakpoint("test.k", 3)
	if err != nil {
		t.Fatalf("add breakpoint #2: %v", err)
	}
	if id1 == id2 {
		t.Fatalf("expected distinct breakpoint ids, got %d and %d", id1, id2)
	}
	id1Again, err := controller.AddBreakpoint("test.k", 2)
	if err != nil {
		t.Fatalf("re-add breakpoint: %v", err)
	}
	if id1Again != id1 {
		t.Fatalf("expected same breakpoint id on re-add, got %d want %d", id1Again, id1)
	}

	bps := controller.Breakpoints()
	if len(bps) != 2 {
		t.Fatalf("expected 2 breakpoints, got %d", len(bps))
	}

	if err := controller.RemoveBreakpoint(id1); err != nil {
		t.Fatalf("remove breakpoint: %v", err)
	}
	if err := controller.RemoveBreakpoint(id1); err == nil {
		t.Fatalf("expected error removing non-existing breakpoint")
	}

	bps = controller.Breakpoints()
	if len(bps) != 1 {
		t.Fatalf("expected 1 breakpoint after delete, got %d", len(bps))
	}
	if bps[0].ID != id2 {
		t.Fatalf("expected remaining breakpoint id %d, got %d", id2, bps[0].ID)
	}

	removed := controller.ClearBreakpoints()
	if removed != 1 {
		t.Fatalf("expected clear to remove 1 breakpoint, got %d", removed)
	}
	if got := len(controller.Breakpoints()); got != 0 {
		t.Fatalf("expected no breakpoints after clear, got %d", got)
	}
}

func waitDebugStop(t *testing.T, controller *interpreter.DebugController) interpreter.DebugStopReason {
	t.Helper()
	ch := make(chan interpreter.DebugStopReason, 1)
	go func() {
		ch <- controller.WaitForStop()
	}()
	select {
	case reason := <-ch:
		return reason
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for debugger stop")
		return 0
	}
}
