package tests

import (
	"testing"

	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

func evalWithPolicy(t *testing.T, input string, policy string) (interpreter.Value, *interpreter.Evaluator, error) {
	t.Helper()

	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.Fatalf("parse failed")
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(input, "<test>")
	if policy != "" {
		if err := eval.SetTaskFailurePolicy(policy); err != nil {
			t.Fatalf("set policy failed: %v", err)
		}
	}
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig)
	}
	return val, eval, err
}

func TestEvalDeferModeObservedRecoveredTaskNotReportedUnhandled(t *testing.T) {
	input := `
let boom = () -> { let obj = {}; obj.missing }
let t = & boom()
let out = (wait t) ? { "fallback" }
out
`
	val, eval, err := evalWithPolicy(t, input, interpreter.TaskFailurePolicyDefer)
	if err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
	assertString(t, val, "fallback")

	if err := eval.CheckUnhandledTaskFailures(); err != nil {
		t.Fatalf("did not expect unhandled failures, got: %v", err)
	}
}

func TestEvalDeferCanceledDetachedTaskNotReportedUnhandled(t *testing.T) {
	input := `
let slow = () -> { sleep(1000); 1 }
let t = & slow()
t.cancel()
sleep(20)
"ok"
`
	val, eval, err := evalWithPolicy(t, input, interpreter.TaskFailurePolicyDefer)
	if err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
	assertString(t, val, "ok")

	if err := eval.CheckUnhandledTaskFailures(); err != nil {
		t.Fatalf("did not expect unhandled failures for canceled task, got: %v", err)
	}
}

func TestEvalDeferModeDoesNotInterruptMainFlow(t *testing.T) {
	input := `
let boom = () -> { sleep(20); let obj = {}; obj.missing }
& boom()
sleep(80)
1
`
	val, eval, err := evalWithPolicy(t, input, interpreter.TaskFailurePolicyDefer)
	if err != nil {
		t.Fatalf("unexpected eval error in defer mode: %v", err)
	}
	assertInteger(t, val, 1)

	if err := eval.CheckUnhandledTaskFailures(); err == nil {
		t.Fatalf("expected deferred unhandled task failure error")
	}
}
