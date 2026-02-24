package tests

import (
	"testing"

	"karl/interpreter"
)

type debugRecorder struct {
	before []interpreter.DebugEvent
	after  []interpreter.DebugEvent
	push   []interpreter.DebugFrame
	pop    []interpreter.DebugFrame
}

func (r *debugRecorder) BeforeNode(event interpreter.DebugEvent) error {
	r.before = append(r.before, event)
	return nil
}

func (r *debugRecorder) AfterNode(event interpreter.DebugEvent, _ interpreter.Value, _ *interpreter.Signal, _ error) error {
	r.after = append(r.after, event)
	return nil
}

func (r *debugRecorder) OnFramePush(frame interpreter.DebugFrame) {
	r.push = append(r.push, frame)
}

func (r *debugRecorder) OnFramePop(frame interpreter.DebugFrame) {
	r.pop = append(r.pop, frame)
}

func TestDebuggerHooksReceiveNodeCallbacks(t *testing.T) {
	program := parseProgram(t, `
let add = (a, b) -> a + b
add(1, 2)
`)
	recorder := &debugRecorder{}
	eval := interpreter.NewEvaluator()
	eval.SetDebugger(recorder)

	val, sig, err := eval.Eval(program, interpreter.NewBaseEnvironment())
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig.Type)
	}
	assertInteger(t, val, 3)

	if len(recorder.before) == 0 {
		t.Fatalf("expected before callbacks, got none")
	}
	if len(recorder.after) == 0 {
		t.Fatalf("expected after callbacks, got none")
	}
	if len(recorder.before) != len(recorder.after) {
		t.Fatalf("before/after callback count mismatch: %d vs %d", len(recorder.before), len(recorder.after))
	}
	if recorder.before[0].NodeType != "*ast.Program" {
		t.Fatalf("expected first callback on Program, got %s", recorder.before[0].NodeType)
	}
}

func TestDebuggerTracksFunctionFrames(t *testing.T) {
	program := parseProgram(t, `
let inc = x -> x + 1
inc(41)
`)
	recorder := &debugRecorder{}
	eval := interpreter.NewEvaluator()
	eval.SetDebugger(recorder)

	val, sig, err := eval.Eval(program, interpreter.NewBaseEnvironment())
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig.Type)
	}
	assertInteger(t, val, 42)

	if len(recorder.push) != 1 || len(recorder.pop) != 1 {
		t.Fatalf("expected exactly one push/pop frame event, got push=%d pop=%d", len(recorder.push), len(recorder.pop))
	}
	if recorder.push[0].Depth != 1 {
		t.Fatalf("expected pushed frame depth 1, got %d", recorder.push[0].Depth)
	}
	if recorder.push[0].Name != "inc" {
		t.Fatalf("expected frame name inc, got %q", recorder.push[0].Name)
	}

	seenFrameDepth := false
	for _, event := range recorder.before {
		if event.FrameDepth > 0 {
			seenFrameDepth = true
			break
		}
	}
	if !seenFrameDepth {
		t.Fatalf("expected at least one callback inside a function frame")
	}

	if stack := eval.DebugStack(); len(stack) != 0 {
		t.Fatalf("expected empty debug stack after eval, got %d frame(s)", len(stack))
	}
}
