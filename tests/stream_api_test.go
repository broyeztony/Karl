package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStreamPipeFileRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "in.txt")
	outPath := filepath.Join(tempDir, "out.txt")
	content := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write in file: %v", err)
	}

	input := fmt.Sprintf(`
let src = reader(%q, { type: BYTES, })
let dst = writer(%q, { type: BYTES, })
let st = pipe(src, dst, { bufferSize: 3, })
src.close()
dst.close()
{ bytes: st.bytes, chunks: st.chunks, out: readFile(%q), }
`, inPath, outPath, outPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	obj, ok := val.(*Object)
	if !ok {
		t.Fatalf("expected object, got %T", val)
	}
	assertInteger(t, obj.Pairs["bytes"], int64(len(content)))
	chunks, ok := obj.Pairs["chunks"].(*Integer)
	if !ok || chunks.Value <= 0 {
		t.Fatalf("expected chunks > 0, got %v", obj.Pairs["chunks"])
	}
	assertString(t, obj.Pairs["out"], content)
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
