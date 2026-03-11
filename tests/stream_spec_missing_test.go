package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

func evalInputWithStdin(t *testing.T, input string, stdin string) (Value, error) {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.Fatalf("parse failed")
	}

	eval := interpreter.NewEvaluator()
	eval.SetInput(strings.NewReader(stdin))
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if sig != nil {
		return nil, fmt.Errorf("unexpected signal: %v", sig.Type)
	}
	return val, err
}

func TestStreamJSONStage(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "json-lines.log")
	content := "{\"id\":1}\n{\"id\":2}\n{\"id\":3}\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
read(%q, { type: TEXT, })
| lines()
| json()
| map(x -> x.id)
| collect()
`, inPath)

	val, err := evalInput(t, input)
	if err == nil {
		// Keep as failing test until json() stage is implemented.
		arr, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected array, got %T", val)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 values, got %d", len(arr.Elements))
		}
		assertInteger(t, arr.Elements[0], 1)
		assertInteger(t, arr.Elements[1], 2)
		assertInteger(t, arr.Elements[2], 3)
		return
	}
	t.Fatalf("eval error: %v", err)
}

func TestStreamDistinctAndGroupCount(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "group.log")
	content := "a\na\nb\nb\nc\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let unique = read(%q, { type: TEXT, }) | lines() | distinct() | collect()
let counts = read(%q, { type: TEXT, }) | lines() | group_count()
;
[unique, counts.get("a"), counts.get("b"), counts.get("c")]
`, inPath, inPath)

	val, err := evalInput(t, input)
	if err == nil {
		out, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected array, got %T", val)
		}
		unique, ok := out.Elements[0].(*Array)
		if !ok {
			t.Fatalf("expected unique array, got %T", out.Elements[0])
		}
		if len(unique.Elements) != 3 {
			t.Fatalf("expected 3 unique values, got %d", len(unique.Elements))
		}
		assertString(t, unique.Elements[0], "a")
		assertString(t, unique.Elements[1], "b")
		assertString(t, unique.Elements[2], "c")
		assertInteger(t, out.Elements[1], 2)
		assertInteger(t, out.Elements[2], 2)
		assertInteger(t, out.Elements[3], 1)
		return
	}
	t.Fatalf("eval error: %v", err)
}

func TestStreamReduceByKeySink(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "reduce-by-key.log")
	content := "cpu|2\ncpu|3\nmem|5\ncpu|7\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let totals = read(%q, { type: TEXT, })
    | lines()
    | map(line -> {
        let p = line.split("|")
        { key: p[0], value: parseInt(p[1]), }
    })
    | reduce_by_key(x -> x.key, 0, (acc, x) -> acc + x.value)
;
[totals.get("cpu"), totals.get("mem")]
`, inPath)

	val, err := evalInput(t, input)
	if err == nil {
		out, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected array, got %T", val)
		}
		assertInteger(t, out.Elements[0], 12)
		assertInteger(t, out.Elements[1], 5)
		return
	}
	t.Fatalf("eval error: %v", err)
}

func TestStreamSortStageAndTopSink(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "sort-top.log")
	content := "7\n2\n9\n1\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let sorted = read(%q, { type: TEXT, })
    | lines()
    | map(x -> parseInt(x))
    | sort((a, b) -> a - b)
    | collect()
let best = read(%q, { type: TEXT, })
    | lines()
    | map(x -> parseInt(x))
    | top(2)
;
[sorted, best]
`, inPath, inPath)

	val, err := evalInput(t, input)
	if err == nil {
		out, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected array, got %T", val)
		}
		sorted, ok := out.Elements[0].(*Array)
		if !ok {
			t.Fatalf("expected sorted array, got %T", out.Elements[0])
		}
		if len(sorted.Elements) != 4 {
			t.Fatalf("expected 4 sorted values, got %d", len(sorted.Elements))
		}
		assertInteger(t, sorted.Elements[0], 1)
		assertInteger(t, sorted.Elements[1], 2)
		assertInteger(t, sorted.Elements[2], 7)
		assertInteger(t, sorted.Elements[3], 9)

		best, ok := out.Elements[1].(*Array)
		if !ok {
			t.Fatalf("expected top array, got %T", out.Elements[1])
		}
		if len(best.Elements) != 2 {
			t.Fatalf("expected 2 top values, got %d", len(best.Elements))
		}
		assertInteger(t, best.Elements[0], 9)
		assertInteger(t, best.Elements[1], 7)
		return
	}
	t.Fatalf("eval error: %v", err)
}

func TestStreamSplitSink(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "split.log")
	content := "ok-1\nerr-1\nok-2\nerr-2\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let out = read(%q, { type: TEXT, })
    | lines()
    | split(x -> x.startsWith("ok"))
;
[out.left, out.right]
`, inPath)

	val, err := evalInput(t, input)
	if err == nil {
		arr, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected array, got %T", val)
		}
		left, ok := arr.Elements[0].(*Array)
		if !ok {
			t.Fatalf("expected left array, got %T", arr.Elements[0])
		}
		right, ok := arr.Elements[1].(*Array)
		if !ok {
			t.Fatalf("expected right array, got %T", arr.Elements[1])
		}
		if len(left.Elements) != 2 || len(right.Elements) != 2 {
			t.Fatalf("unexpected split sizes: left=%d right=%d", len(left.Elements), len(right.Elements))
		}
		assertString(t, left.Elements[0], "ok-1")
		assertString(t, left.Elements[1], "ok-2")
		assertString(t, right.Elements[0], "err-1")
		assertString(t, right.Elements[1], "err-2")
		return
	}
	t.Fatalf("eval error: %v", err)
}

func TestStreamExecSink(t *testing.T) {
	bin := mustExecutable(t)
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "exec.log")
	if err := os.WriteFile(inPath, []byte("hello from exec sink"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let st = read(%q, { type: TEXT, })
    | exec({
        command: %q,
        args: ["-test.run=TestProcessAPIHelperProcess", "capture"],
        env: { KARL_PROCESS_HELPER: "1", },
        inheritEnv: true,
    })
;
[st.ok, st.code, st.output]
`, inPath, bin)

	val, err := evalInput(t, input)
	if err == nil {
		out, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected array, got %T", val)
		}
		assertBoolean(t, out.Elements[0], true)
		assertInteger(t, out.Elements[1], 0)
		assertString(t, out.Elements[2], "CAP|hello from exec sink")
		return
	}
	t.Fatalf("eval error: %v", err)
}

func TestStreamStdinSource(t *testing.T) {
	input := `stdin({ type: TEXT, }) | lines() | take(2) | collect()`
	val, err := evalInputWithStdin(t, input, "line-1\nline-2\nline-3\n")
	if err == nil {
		arr, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected array, got %T", val)
		}
		if len(arr.Elements) != 2 {
			t.Fatalf("expected 2 values, got %d", len(arr.Elements))
		}
		assertString(t, arr.Elements[0], "line-1")
		assertString(t, arr.Elements[1], "line-2")
		return
	}
	t.Fatalf("eval error: %v", err)
}

func TestHTTPStringSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("a\nb\n"))
	}))
	defer srv.Close()

	input := fmt.Sprintf(`http(%q) | lines() | collect()`, srv.URL)
	val, err := evalInput(t, input)
	if err == nil {
		arr, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected array, got %T", val)
		}
		if len(arr.Elements) != 2 {
			t.Fatalf("expected 2 values, got %d", len(arr.Elements))
		}
		assertString(t, arr.Elements[0], "a")
		assertString(t, arr.Elements[1], "b")
		return
	}
	t.Fatalf("eval error: %v", err)
}
