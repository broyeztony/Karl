package tests

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProcessRunHelperProcess(t *testing.T) {
	if os.Getenv("KARL_PROCESS_HELPER") != "1" {
		return
	}

	switch os.Getenv("KARL_PROCESS_MODE") {
	case "ok":
		cwd, _ := os.Getwd()
		stdin, _ := io.ReadAll(os.Stdin)
		_, _ = fmt.Fprintf(os.Stdout, "OUT|%s|%s|%s", os.Getenv("KARL_PROCESS_TEST_ENV"), cwd, string(stdin))
		_, _ = os.Stderr.WriteString("ERR")
		os.Exit(0)
	case "exit7":
		_, _ = os.Stdout.WriteString("OUT7")
		_, _ = os.Stderr.WriteString("ERR7")
		os.Exit(7)
	case "sleep":
		time.Sleep(250 * time.Millisecond)
		_, _ = os.Stdout.WriteString("late")
		os.Exit(0)
	case "bigout":
		_, _ = os.Stdout.WriteString(strings.Repeat("x", 4096))
		os.Exit(0)
	default:
		_, _ = os.Stderr.WriteString("unknown mode")
		os.Exit(2)
	}
}

func TestProcessRunHappyPath(t *testing.T) {
	bin, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	tempDir := t.TempDir()

	input := fmt.Sprintf(`
let res = processRun({
    command: %q,
    args: ["-test.run=TestProcessRunHelperProcess"],
    cwd: %q,
    env: {
        KARL_PROCESS_HELPER: "1",
        KARL_PROCESS_MODE: "ok",
        KARL_PROCESS_TEST_ENV: "karl-env",
    },
    inheritEnv: true,
    stdin: "ping",
    timeoutMs: 2000,
    maxOutputBytes: 4096,
})
res
`, bin, tempDir)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	obj, ok := val.(*Object)
	if !ok {
		t.Fatalf("expected object, got %T (%v)", val, val)
	}

	assertBoolean(t, obj.Pairs["ok"], true)
	assertInteger(t, obj.Pairs["exitCode"], 0)
	stdout, ok := obj.Pairs["stdout"].(*String)
	if !ok {
		t.Fatalf("expected stdout string, got %T (%v)", obj.Pairs["stdout"], obj.Pairs["stdout"])
	}
	parts := strings.SplitN(stdout.Value, "|", 4)
	if len(parts) != 4 {
		t.Fatalf("unexpected stdout format: %q", stdout.Value)
	}
	if parts[0] != "OUT" || parts[1] != "karl-env" || parts[3] != "ping" {
		t.Fatalf("unexpected stdout segments: %#v", parts)
	}
	assertSamePath(t, tempDir, parts[2])
	assertString(t, obj.Pairs["stderr"], "ERR")
	assertBoolean(t, obj.Pairs["timedOut"], false)
	assertBoolean(t, obj.Pairs["killed"], false)

	duration, ok := obj.Pairs["durationMs"].(*Integer)
	if !ok {
		t.Fatalf("expected durationMs integer, got %T (%v)", obj.Pairs["durationMs"], obj.Pairs["durationMs"])
	}
	if duration.Value < 0 || duration.Value > 5000 {
		t.Fatalf("unexpected durationMs: %d", duration.Value)
	}
}

func TestProcessRunNonZeroExitReturnsResult(t *testing.T) {
	bin, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}

	input := fmt.Sprintf(`
let res = processRun({
    command: %q,
    args: ["-test.run=TestProcessRunHelperProcess"],
    env: {
        KARL_PROCESS_HELPER: "1",
        KARL_PROCESS_MODE: "exit7",
    },
    inheritEnv: true,
    timeoutMs: 2000,
})
;
[res.ok, res.exitCode, res.stdout, res.stderr, res.timedOut]
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T (%v)", val, val)
	}
	if len(arr.Elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(arr.Elements))
	}

	assertBoolean(t, arr.Elements[0], false)
	assertInteger(t, arr.Elements[1], 7)
	assertString(t, arr.Elements[2], "OUT7")
	assertString(t, arr.Elements[3], "ERR7")
	assertBoolean(t, arr.Elements[4], false)
}

func TestProcessRunTimeoutIsRecoverable(t *testing.T) {
	bin, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}

	input := fmt.Sprintf(`
processRun({
    command: %q,
    args: ["-test.run=TestProcessRunHelperProcess"],
    env: {
        KARL_PROCESS_HELPER: "1",
        KARL_PROCESS_MODE: "sleep",
    },
    inheritEnv: true,
    timeoutMs: 20,
    maxOutputBytes: 4096,
}) ? { error.kind }
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "process_timeout")
}

func TestProcessRunSpawnFailureIsRecoverable(t *testing.T) {
	val, err := evalInput(t, `
processRun({
    command: "__karl_missing_executable_6D69A091__",
    args: [],
    timeoutMs: 200,
}) ? { error.kind }
`)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "process_spawn")
}

func TestProcessRunOutputLimitIsRecoverable(t *testing.T) {
	bin, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}

	input := fmt.Sprintf(`
processRun({
    command: %q,
    args: ["-test.run=TestProcessRunHelperProcess"],
    env: {
        KARL_PROCESS_HELPER: "1",
        KARL_PROCESS_MODE: "bigout",
    },
    inheritEnv: true,
    timeoutMs: 2000,
    maxOutputBytes: 64,
}) ? { error.kind }
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "process_output_limit")
}

func assertBoolean(t *testing.T, val Value, expected bool) {
	t.Helper()
	b, ok := val.(*Boolean)
	if !ok {
		t.Fatalf("expected Boolean, got %T (%v)", val, val)
	}
	if b.Value != expected {
		t.Fatalf("expected %v, got %v", expected, b.Value)
	}
}

func assertSamePath(t *testing.T, expected string, actual string) {
	t.Helper()
	expectedResolved, err := filepath.EvalSymlinks(expected)
	if err != nil {
		expectedResolved = filepath.Clean(expected)
	}
	actualResolved, err := filepath.EvalSymlinks(actual)
	if err != nil {
		actualResolved = filepath.Clean(actual)
	}
	if expectedResolved != actualResolved {
		t.Fatalf("expected path %q, got %q", expectedResolved, actualResolved)
	}
}
