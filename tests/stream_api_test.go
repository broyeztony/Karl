package tests

import (
	"fmt"
	"os"
	"path/filepath"
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
let n = w.write(encodeUtf8("hé"))
w.close()

let r = reader(%q)
let [chunk, eofA] = r.read(8)
let [tail, eofB] = r.read(8)
r.close()

;
[n, len(chunk), chunk.length, decodeUtf8(chunk), eofA, tail, eofB]
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
