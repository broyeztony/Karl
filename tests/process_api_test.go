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

func TestProcessAPIHelperProcess(t *testing.T) {
	if os.Getenv("KARL_PROCESS_HELPER") != "1" {
		return
	}

	mode := os.Getenv("KARL_PROCESS_MODE")
	if mode == "" && len(os.Args) > 1 {
		mode = os.Args[len(os.Args)-1]
	}

	switch mode {
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
		time.Sleep(300 * time.Millisecond)
		_, _ = os.Stdout.WriteString("late")
		os.Exit(0)
	case "bigout":
		_, _ = os.Stdout.WriteString(strings.Repeat("x", 4096))
		os.Exit(0)
	case "emit":
		_, _ = os.Stdout.WriteString("alpha\nbeta\n")
		os.Exit(0)
	case "capture":
		stdin, _ := io.ReadAll(os.Stdin)
		_, _ = fmt.Fprintf(os.Stdout, "CAP|%s", string(stdin))
		os.Exit(0)
	case "badutf8":
		_, _ = os.Stdout.Write([]byte{0xff, 0xfe, 0xfd})
		os.Exit(0)
	default:
		_, _ = os.Stderr.WriteString("unknown mode")
		os.Exit(2)
	}
}

func TestRunReturnsCapturedStatus(t *testing.T) {
	bin := mustExecutable(t)
	tempDir := t.TempDir()

	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "ok"],
    cwd: %q,
    env: {
        KARL_PROCESS_HELPER: "1",
        KARL_PROCESS_TEST_ENV: "karl-env",
    },
    inheritEnv: true,
})

let st = run(stage, {
    stdin: "ping",
})

st
`, bin, tempDir)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	obj, ok := val.(*Object)
	if !ok {
		t.Fatalf("expected object, got %T", val)
	}

	assertBoolean(t, obj.Pairs["ok"], true)
	assertInteger(t, obj.Pairs["code"], 0)
	assertBoolean(t, obj.Pairs["timedOut"], false)
	assertBoolean(t, obj.Pairs["aborted"], false)
	assertBoolean(t, obj.Pairs["outputTruncated"], false)
	assertBoolean(t, obj.Pairs["errorTruncated"], false)
	assertString(t, obj.Pairs["error"], "ERR")

	output, ok := obj.Pairs["output"].(*String)
	if !ok {
		t.Fatalf("expected output string, got %T", obj.Pairs["output"])
	}
	parts := strings.SplitN(output.Value, "|", 4)
	if len(parts) != 4 {
		t.Fatalf("unexpected output format: %q", output.Value)
	}
	if parts[0] != "OUT" || parts[1] != "karl-env" || parts[3] != "ping" {
		t.Fatalf("unexpected output segments: %#v", parts)
	}
	assertSamePath(t, tempDir, parts[2])
}

func TestRunNonZeroExitReturnsStatus(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "exit7"],
    env: { KARL_PROCESS_HELPER: "1", },
    inheritEnv: true,
})
let st = run(stage)
;
[st.ok, st.code, st.output, st.error]
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertBoolean(t, arr.Elements[0], false)
	assertInteger(t, arr.Elements[1], 7)
	assertString(t, arr.Elements[2], "OUT7")
	assertString(t, arr.Elements[3], "ERR7")
}

func TestRunTruncatesByDefault(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "bigout"],
    env: { KARL_PROCESS_HELPER: "1", },
    inheritEnv: true,
})
let st = run(stage, {
    maxOutputBytes: 64,
})
;
[st.ok, st.output.length, st.outputTruncated]
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr := val.(*Array)
	assertBoolean(t, arr.Elements[0], true)
	assertInteger(t, arr.Elements[1], 64)
	assertBoolean(t, arr.Elements[2], true)
}

func TestRunOverflowErrorMode(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "bigout"],
    env: { KARL_PROCESS_HELPER: "1", },
    inheritEnv: true,
})
run(stage, {
    maxOutputBytes: 64,
    overflow: "error",
}) ? { error.kind }
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "process_output_limit")
}

func TestProcWaitAndPipeStreams(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "ok"],
    env: {
        KARL_PROCESS_HELPER: "1",
        KARL_PROCESS_TEST_ENV: "chan",
    },
    inheritEnv: true,
})
let p = proc(stage, {
    stdIn: "pipe",
    stdOut: "pipe",
    stdErr: "pipe",
    stdinType: "text",
    stdoutType: "text",
    stderrType: "text",
})

let collect = s -> for true with acc = "" {
    let [chunk, eof] = s.read()
    if eof { break acc }
    acc = acc + chunk
} then acc

let inStream = p.stdin
let n = inStream.write("ping")
inStream.close()

let out = collect(p.stdout)
let err = collect(p.stderr)
let st = wait p
;

{ ok: st.ok, code: st.code, out: out, err: err, pid: p.pid, running: p.running, written: n, }
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
	assertInteger(t, obj.Pairs["code"], 0)
	assertString(t, obj.Pairs["err"], "ERR")
	assertBoolean(t, obj.Pairs["running"], false)
	assertInteger(t, obj.Pairs["written"], 4)
	pid, ok := obj.Pairs["pid"].(*Integer)
	if !ok || pid.Value <= 0 {
		t.Fatalf("expected pid > 0, got %v", obj.Pairs["pid"])
	}
}

func TestProcModeConstantsWork(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "ok"],
    env: {
        KARL_PROCESS_HELPER: "1",
        KARL_PROCESS_TEST_ENV: "const",
    },
    inheritEnv: true,
})
let p = proc(stage, {
    stdIn: PIPE,
    stdOut: PIPE,
    stdErr: NULL,
    stdinType: TEXT,
    stdoutType: TEXT,
})

let collect = s -> for true with acc = "" {
    let [chunk, eof] = s.read()
    if eof { break acc }
    acc = acc + chunk
} then acc

let inStream = p.stdin
let n = inStream.write("ping")
inStream.close()

let out = collect(p.stdout)
let st = wait p
;
[st.ok, out, n, PIPE, INHERIT, NULL, TEXT, BYTES]
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertBoolean(t, arr.Elements[0], true)
	output, ok := arr.Elements[1].(*String)
	if !ok {
		t.Fatalf("expected output string, got %T", arr.Elements[1])
	}
	parts := strings.SplitN(output.Value, "|", 4)
	if len(parts) != 4 {
		t.Fatalf("unexpected output format: %q", output.Value)
	}
	if parts[0] != "OUT" || parts[1] != "const" || parts[3] != "ping" {
		t.Fatalf("unexpected output segments: %#v", parts)
	}
	assertInteger(t, arr.Elements[2], 4)
	assertString(t, arr.Elements[3], "pipe")
	assertString(t, arr.Elements[4], "inherit")
	assertString(t, arr.Elements[5], "null")
	assertString(t, arr.Elements[6], "text")
	assertString(t, arr.Elements[7], "bytes")
}

func TestProcTimeoutReturnsStatus(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "sleep"],
    env: { KARL_PROCESS_HELPER: "1", },
    inheritEnv: true,
})
let p = proc(stage, {
    timeoutMs: 20,
})
let st = wait p
;
[st.ok, st.timedOut]
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr := val.(*Array)
	assertBoolean(t, arr.Elements[0], false)
	assertBoolean(t, arr.Elements[1], true)
}

func TestProcStreamsDefaultToBytes(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "ok"],
    env: {
        KARL_PROCESS_HELPER: "1",
        KARL_PROCESS_TEST_ENV: "bytes-default",
    },
    inheritEnv: true,
})
let p = proc(stage, {
    stdIn: PIPE,
    stdOut: PIPE,
    stdErr: NULL,
})

let collect = s -> for true with acc = "" {
    let [chunk, eof] = s.read()
    if eof { break acc }
    acc = acc + decodeUtf8(chunk)
} then acc

let n = p.stdin.write(encodeUtf8("ping"))
p.stdin.close()
let out = collect(p.stdout)
let st = wait p
;
[st.ok, out, n]
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr := val.(*Array)
	assertBoolean(t, arr.Elements[0], true)
	output, ok := arr.Elements[1].(*String)
	if !ok {
		t.Fatalf("expected output string, got %T", arr.Elements[1])
	}
	parts := strings.SplitN(output.Value, "|", 4)
	if len(parts) != 4 {
		t.Fatalf("unexpected output format: %q", output.Value)
	}
	if parts[0] != "OUT" || parts[1] != "bytes-default" || parts[3] != "ping" {
		t.Fatalf("unexpected output segments: %#v", parts)
	}
	assertInteger(t, arr.Elements[2], 4)
}

func TestDecodeUtf8InvalidBytesRecoverable(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "badutf8"],
    env: { KARL_PROCESS_HELPER: "1", },
    inheritEnv: true,
})
let p = proc(stage, {
    stdOut: PIPE,
    stdErr: NULL,
})
let [chunk, _] = p.stdout.read()
let _ = wait p
decodeUtf8(chunk) ? { error.kind }
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "utf8_decode")
}

func TestProcPipelineComposition(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let plan =
    cmd({ command: %q, args: ["-test.run=TestProcessAPIHelperProcess", "emit"], env: { KARL_PROCESS_HELPER: "1", }, inheritEnv: true, })
    | cmd({ command: %q, args: ["-test.run=TestProcessAPIHelperProcess", "capture"], env: { KARL_PROCESS_HELPER: "1", }, inheritEnv: true, })

let p = proc(plan, {
    stdOut: "pipe",
    stdErr: "pipe",
    stdIn: "null",
    stdoutType: "text",
    stderrType: "text",
})

let collect = s -> for true with acc = "" {
    let [chunk, eof] = s.read()
    if eof { break acc }
    acc = acc + chunk
} then acc

let out = collect(p.stdout)
let st = wait p
;
[st.ok, out]
`, bin, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr := val.(*Array)
	assertBoolean(t, arr.Elements[0], true)
	assertString(t, arr.Elements[1], "CAP|alpha\nbeta\n")
}

func TestProcPipelineCompositionRepeatedRuns(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let plan =
    cmd({ command: %q, args: ["-test.run=TestProcessAPIHelperProcess", "emit"], env: { KARL_PROCESS_HELPER: "1", }, inheritEnv: true, })
    | cmd({ command: %q, args: ["-test.run=TestProcessAPIHelperProcess", "capture"], env: { KARL_PROCESS_HELPER: "1", }, inheritEnv: true, })

let once = () -> {
    let p = proc(plan, {
        stdOut: "pipe",
        stdErr: "pipe",
        stdIn: "null",
        stdoutType: "text",
        stderrType: "text",
    })

    let collect = s -> for true with acc = "" {
        let [chunk, eof] = s.read()
        if eof { break acc }
        acc = acc + chunk
    } then acc

    let out = collect(p.stdout)
    let st = wait p
    if !st.ok { fail("pipeline failed") }
    out
}

for i < 20 with i = 0, last = "" {
    last = once()
    i++
} then last
`, bin, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "CAP|alpha\nbeta\n")
}

func TestStdOutRequiresPipeMode(t *testing.T) {
	bin := mustExecutable(t)
	input := fmt.Sprintf(`
let stage = cmd({
    command: %q,
    args: ["-test.run=TestProcessAPIHelperProcess", "ok"],
    env: { KARL_PROCESS_HELPER: "1", KARL_PROCESS_TEST_ENV: "state", },
    inheritEnv: true,
})
let p = proc(stage, {
    stdOut: "null",
    stdErr: "null",
})
p.stdout ? { error.kind }
`, bin)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "process_state")
}

func mustExecutable(t *testing.T) string {
	t.Helper()
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	return path
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
