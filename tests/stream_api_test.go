package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamReadWritePipelineRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "in.txt")
	outPath := filepath.Join(tempDir, "out.txt")
	content := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write in file: %v", err)
	}

	input := fmt.Sprintf(`
let result = read(%q, { type: BYTES, }) | write(%q, { type: BYTES, })
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
	assertString(t, arr.Elements[1], content)
}

func TestStreamReadWriteAndEOF(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "stream.txt")

	input := fmt.Sprintf(`
let w = writer(%q, { type: TEXT, })
let nA = w.write("ab")
let nB = w.write("c")
w.close()

let r = reader(%q, { type: TEXT, })
let [a, eofA] = r.read(2)
let [b, eofB] = r.read(2)
let [c, eofC] = r.read(2)
r.close()

;
[nA, nB, a, eofA, b, eofB, c, eofC]
`, path, path)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertInteger(t, arr.Elements[0], 2)
	assertInteger(t, arr.Elements[1], 1)
	assertString(t, arr.Elements[2], "ab")
	assertBoolean(t, arr.Elements[3], false)
	assertString(t, arr.Elements[4], "c")
	assertBoolean(t, arr.Elements[5], false)
	if !Equivalent(arr.Elements[6], NullValue) {
		t.Fatalf("expected null on eof chunk, got %v", arr.Elements[6])
	}
	assertBoolean(t, arr.Elements[7], true)
}

func TestStreamBytesModeUsesBytesValues(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "stream-bytes.bin")

	input := fmt.Sprintf(`
let w = writer(%q)
let n = w.write(toUtf8("hé"))
w.close()

let r = reader(%q)
let [chunk, eofA] = r.read(8)
let [tail, eofB] = r.read(8)
r.close()

;
[n, len(chunk), chunk.length, fromUtf8(chunk), eofA, tail, eofB]
`, path, path)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertInteger(t, arr.Elements[0], 3)
	assertInteger(t, arr.Elements[1], 3)
	assertInteger(t, arr.Elements[2], 3)
	assertString(t, arr.Elements[3], "hé")
	assertBoolean(t, arr.Elements[4], false)
	if !Equivalent(arr.Elements[5], NullValue) {
		t.Fatalf("expected null on eof chunk, got %v", arr.Elements[5])
	}
	assertBoolean(t, arr.Elements[6], true)
}

func TestStreamCollectPreservesBytesChunks(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "collect-bytes.bin")

	input := fmt.Sprintf(`
let w = writer(%q)
let _ = w.write(toUtf8("abc"))
w.close()

let chunks = read(%q) | collect()
let bytes = bytesJoin(chunks)

;
[len(chunks), bytes.length, fromUtf8(bytes)]
`, path, path)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertInteger(t, arr.Elements[0], 1)
	assertInteger(t, arr.Elements[1], 3)
	assertString(t, arr.Elements[2], "abc")
}

func TestStreamPlanReadMemberSequentialEOF(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "stream-read-plan.txt")
	content := "alpha\nbeta\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	input := fmt.Sprintf(`
let s = read(%q, { type: TEXT, }) | lines()
let [a, eofA] = s.read()
let [b, eofB] = s.read()
let [c, eofC] = s.read()
let [d, eofD] = s.read()
;
[a, eofA, b, eofB, c, eofC, d, eofD]
`, path)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}

	assertString(t, arr.Elements[0], "alpha")
	assertBoolean(t, arr.Elements[1], false)
	assertString(t, arr.Elements[2], "beta")
	assertBoolean(t, arr.Elements[3], false)
	if !Equivalent(arr.Elements[4], NullValue) {
		t.Fatalf("expected null at eof, got %v", arr.Elements[4])
	}
	assertBoolean(t, arr.Elements[5], true)
	if !Equivalent(arr.Elements[6], NullValue) {
		t.Fatalf("expected null after eof, got %v", arr.Elements[6])
	}
	assertBoolean(t, arr.Elements[7], true)
}

func TestStreamPlanCloseMemberStopsReads(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "stream-close-plan.txt")
	content := "alpha\nbeta\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	input := fmt.Sprintf(`
let s = read(%q, { type: TEXT, }) | lines()
let [a, eofA] = s.read()
s.close()
let [b, eofB] = s.read()
;
[a, eofA, b, eofB]
`, path)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}

	assertString(t, arr.Elements[0], "alpha")
	assertBoolean(t, arr.Elements[1], false)
	if !Equivalent(arr.Elements[2], NullValue) {
		t.Fatalf("expected null after close, got %v", arr.Elements[2])
	}
	assertBoolean(t, arr.Elements[3], true)
}

func TestStreamPlanReadMemberRejectsConcurrentRead(t *testing.T) {
	input := `
let c = channel()
let s = fromChannel(c)
let t = & (() -> s.read())()
sleep(20)
let msg = s.read() ? error.message
c.send("x")
c.done()
let first = wait t
let [item, eof] = first
s.close()
;
[msg, item, eof]
`

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}

	msg, ok := arr.Elements[0].(*String)
	if !ok {
		t.Fatalf("expected concurrent-read message string, got %T", arr.Elements[0])
	}
	if !strings.Contains(msg.Value, "already in progress") {
		t.Fatalf("expected concurrent-read error, got %q", msg.Value)
	}
	assertString(t, arr.Elements[1], "x")
	assertBoolean(t, arr.Elements[2], false)
}

func TestStreamFromJsonToJsonStages(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "json-lines.log")
	content := "{\"id\":1}\n{\"id\":2}\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	input := fmt.Sprintf(`
let out = read(%q, { type: TEXT, })
  | lines()
  | fromJson()
  | map(x -> { id: x.id + 10, })
  | toJson()
  | collect()
;
[len(out), out[0], out[1]]
`, path)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertInteger(t, arr.Elements[0], 2)
	assertString(t, arr.Elements[1], "{\"id\":11}")
	assertString(t, arr.Elements[2], "{\"id\":12}")
}

func TestStreamUtf8Stages(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "in-bytes.bin")
	outPath := filepath.Join(tempDir, "out-bytes.bin")
	if err := os.WriteFile(inPath, []byte("alpha\nbeta\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	input := fmt.Sprintf(`
read(%q, { type: BYTES, })
  | fromUtf8()
  | lines()
  | map(x -> x + "!\n")
  | toUtf8()
  | write(%q, { type: BYTES, })
;
readFile(%q)
`, inPath, outPath, outPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "alpha!\nbeta!\n")
}

func TestStreamForEachSink(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "for-each.log")
	content := "a\nb\nc\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	input := fmt.Sprintf(`
let out = []
let done = read(%q, { type: TEXT, }) | lines() | forEach(x -> out.push(x + "!"))
;
[done, out]
`, path)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if arr.Elements[0] == nil || arr.Elements[0].Inspect() != "()" {
		t.Fatalf("expected Unit for forEach sink, got %v", arr.Elements[0])
	}
	out, ok := arr.Elements[1].(*Array)
	if !ok {
		t.Fatalf("expected output array, got %T", arr.Elements[1])
	}
	if len(out.Elements) != 3 {
		t.Fatalf("expected 3 output values, got %d", len(out.Elements))
	}
	assertString(t, out.Elements[0], "a!")
	assertString(t, out.Elements[1], "b!")
	assertString(t, out.Elements[2], "c!")
}
