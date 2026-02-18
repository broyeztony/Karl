package tests

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

type failingReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

func evalWithConfiguredEvaluator(t *testing.T, input string, configure func(*interpreter.Evaluator)) (Value, error) {
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
	if configure != nil {
		configure(eval)
	}

	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if sig != nil {
		return nil, fmt.Errorf("unexpected signal: %v", sig.Type)
	}
	return val, err
}

func stringsFromArray(t *testing.T, val Value) []string {
	t.Helper()
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected Array, got %T (%v)", val, val)
	}
	out := make([]string, 0, len(arr.Elements))
	for _, el := range arr.Elements {
		s, ok := el.(*String)
		if !ok {
			t.Fatalf("expected String element, got %T (%v)", el, el)
		}
		out = append(out, s.Value)
	}
	return out
}

func TestRuntimeIOArgvInRunContext(t *testing.T) {
	val, err := evalWithConfiguredEvaluator(t, `argv()`, func(eval *interpreter.Evaluator) {
		eval.SetProgramArgs([]string{"alpha", "beta"})
		eval.SetProgramPath("app.k")
	})
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	got := stringsFromArray(t, val)
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("unexpected argv: %#v", got)
	}
}

func TestRuntimeIOArgvDefaultsToEmpty(t *testing.T) {
	val, err := evalInput(t, `argv().length`)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertInteger(t, val, 0)
}

func TestRuntimeIOProgramPathInRunContext(t *testing.T) {
	val, err := evalWithConfiguredEvaluator(t, `programPath()`, func(eval *interpreter.Evaluator) {
		eval.SetProgramPath("examples/app.k")
	})
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "examples/app.k")
}

func TestRuntimeIOProgramPathInStdinRunContext(t *testing.T) {
	val, err := evalWithConfiguredEvaluator(t, `programPath()`, func(eval *interpreter.Evaluator) {
		eval.SetProgramPath("<stdin>")
	})
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "<stdin>")
}

func TestRuntimeIOProgramPathDefaultsToNull(t *testing.T) {
	val, err := evalInput(t, `programPath()`)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if !Equivalent(val, NullValue) {
		t.Fatalf("expected null, got %s", val.Inspect())
	}
}

func TestRuntimeIOEnvironReturnsSnapshot(t *testing.T) {
	key := fmt.Sprintf("KARL_TEST_ENVIRON_%d", time.Now().UnixNano())
	t.Setenv(key, "value")

	val, err := evalInput(t, `environ()`)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	entries := stringsFromArray(t, val)
	needle := key + "=value"
	found := false
	for _, entry := range entries {
		if entry == needle {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %q in environ()", needle)
	}
}

func TestRuntimeIOEnvReturnsValueWhenPresent(t *testing.T) {
	key := fmt.Sprintf("KARL_TEST_ENV_%d", time.Now().UnixNano())
	t.Setenv(key, "bar")
	val, err := evalInput(t, fmt.Sprintf(`env("%s")`, key))
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "bar")
}

func TestRuntimeIOEnvReturnsNullWhenMissing(t *testing.T) {
	key := fmt.Sprintf("KARL_TEST_ENV_MISSING_%d", time.Now().UnixNano())
	_ = os.Unsetenv(key)
	val, err := evalInput(t, fmt.Sprintf(`env("%s")`, key))
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if !Equivalent(val, NullValue) {
		t.Fatalf("expected null, got %s", val.Inspect())
	}
}

func TestRuntimeIOEnvPreservesEmptyString(t *testing.T) {
	key := fmt.Sprintf("KARL_TEST_ENV_EMPTY_%d", time.Now().UnixNano())
	t.Setenv(key, "")
	val, err := evalInput(t, fmt.Sprintf(`env("%s")`, key))
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "")
}

func TestRuntimeIOReadLineReturnsLinesWithoutNewline(t *testing.T) {
	val, err := evalWithConfiguredEvaluator(t, `[readLine(), readLine()]`, func(eval *interpreter.Evaluator) {
		eval.SetInput(strings.NewReader("first\nsecond\n"))
	})
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	got := stringsFromArray(t, val)
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("unexpected readLine values: %#v", got)
	}
}

func TestRuntimeIOReadLineReturnsNullOnEOF(t *testing.T) {
	val, err := evalWithConfiguredEvaluator(t, `readLine()`, func(eval *interpreter.Evaluator) {
		eval.SetInput(strings.NewReader(""))
	})
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if !Equivalent(val, NullValue) {
		t.Fatalf("expected null, got %s", val.Inspect())
	}
}

func TestRuntimeIOReadLineErrorIsRecoverable(t *testing.T) {
	val, err := evalWithConfiguredEvaluator(t, `readLine() ? { error.kind }`, func(eval *interpreter.Evaluator) {
		eval.SetInput(failingReader{})
	})
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "readLine")
}

func TestRuntimeIOReadLineUnavailableByDefault(t *testing.T) {
	val, err := evalInput(t, `readLine() ? { error.kind }`)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "readLine")
}
