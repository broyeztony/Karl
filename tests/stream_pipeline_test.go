package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadLinesWritePipeline(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "in.log")
	outPath := filepath.Join(tempDir, "out.log")
	content := "alpha\nbeta\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let result = read(%q) | lines() | write(%q)
;
[result, readFile(%q)]
`, inPath, outPath, outPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if arr.Elements[0] == nil || arr.Elements[0].Inspect() != "()" {
		t.Fatalf("expected Unit pipeline result, got %v", arr.Elements[0])
	}
	out, ok := arr.Elements[1].(*String)
	if !ok {
		t.Fatalf("expected output string, got %T", arr.Elements[1])
	}
	if out.Value != content {
		t.Fatalf("expected %q, got %q", content, out.Value)
	}
}

func TestSpawnPipelineExpression(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "in.log")
	outPath := filepath.Join(tempDir, "out.log")
	content := "one\ntwo\nthree\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let task = & (read(%q) | lines() | write(%q))
wait task
;
readFile(%q)
`, inPath, outPath, outPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	out, ok := val.(*String)
	if !ok {
		t.Fatalf("expected output string, got %T", val)
	}
	if out.Value != content {
		t.Fatalf("expected %q, got %q", content, out.Value)
	}
}

func TestStreamPipeRejectsInvalidRightOperand(t *testing.T) {
	_, err := evalInput(t, `read("file.txt") | 1`)
	if err == nil {
		t.Fatalf("expected runtime error")
	}
	if !strings.Contains(err.Error(), "operator '|' expects stream stage or sink on the right") {
		t.Fatalf("unexpected runtime error: %v", err)
	}
}

func TestProcessStdoutAsPipelineSource(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let p = proc(cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "emit"],
    env: { KARL_PROCESS_HELPER: "1", },
    inheritEnv: true,
}), {
    stdOut: PIPE,
    stdErr: NULL,
    stdoutType: TEXT,
})
let linesOut = p.stdout | lines() | collect()
let st = wait p
;
{ ok: st.ok, out: linesOut, }
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	obj, ok := val.(*Object)
	if !ok {
		t.Fatalf("expected object, got %T", val)
	}
	assertBoolean(t, obj.Pairs["ok"], true)
	outArr, ok := obj.Pairs["out"].(*Array)
	if !ok {
		t.Fatalf("expected out array, got %T", obj.Pairs["out"])
	}
	if len(outArr.Elements) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(outArr.Elements))
	}
	assertString(t, outArr.Elements[0], "alpha")
	assertString(t, outArr.Elements[1], "beta")
}
