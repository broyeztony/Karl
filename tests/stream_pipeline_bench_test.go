package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"karl/ast"
	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

const (
	streamBenchPayloadBytes = int64(64 * 1024 * 1024) // 64 MiB
)

func BenchmarkPipelineFileToFile64MiB(b *testing.B) {
	benchPipelineFileToFile(b, streamBenchPayloadBytes)
}

func BenchmarkPipelineProcessStdoutToFile64MiB(b *testing.B) {
	if _, err := exec.LookPath("cat"); err != nil {
		b.Skip("cat not found in PATH")
	}
	benchPipelineProcessStdoutToFile(b, streamBenchPayloadBytes)
}

func benchPipelineFileToFile(b *testing.B, payloadBytes int64) {
	tempDir := b.TempDir()
	inPath := filepath.Join(tempDir, "in.bin")
	outPath := filepath.Join(tempDir, "out.bin")
	mustWriteBenchFixture(b, inPath, payloadBytes)

	source := fmt.Sprintf(`
read(%q, { type: BYTES, }) | write(%q, { type: BYTES, })
`, inPath, outPath)
	program := mustParseBenchProgram(b, source)

	b.ReportAllocs()
	b.SetBytes(payloadBytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val := mustEvalBenchProgram(b, source, program)
		if val == nil || val.Inspect() != "()" {
			b.Fatalf("expected unit result, got %v", val)
		}
		info, err := os.Stat(outPath)
		if err != nil {
			b.Fatalf("stat output: %v", err)
		}
		if info.Size() != payloadBytes {
			b.Fatalf("expected %d bytes, got %d", payloadBytes, info.Size())
		}
	}
}

func benchPipelineProcessStdoutToFile(b *testing.B, payloadBytes int64) {
	tempDir := b.TempDir()
	inPath := filepath.Join(tempDir, "in.bin")
	outPath := filepath.Join(tempDir, "out.bin")
	mustWriteBenchFixture(b, inPath, payloadBytes)

	source := fmt.Sprintf(`
let p = proc({ command: "cat", args: [%q], }, { stdout: PIPE, stderr: NULL, })
let _ = p.stdout | write(%q, { type: BYTES, })
let ps = wait p
if !ps.ok { fail("cat failed") }
()
`, inPath, outPath)
	program := mustParseBenchProgram(b, source)

	b.ReportAllocs()
	b.SetBytes(payloadBytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val := mustEvalBenchProgram(b, source, program)
		if val == nil || val.Inspect() != "()" {
			b.Fatalf("expected unit result, got %v", val)
		}
		info, err := os.Stat(outPath)
		if err != nil {
			b.Fatalf("stat output: %v", err)
		}
		if info.Size() != payloadBytes {
			b.Fatalf("expected %d bytes, got %d", payloadBytes, info.Size())
		}
	}
}

func mustParseBenchProgram(b *testing.B, source string) *ast.Program {
	b.Helper()
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, err := range errs {
			b.Logf("parse error: %s", err)
		}
		b.Fatalf("parse failed")
	}
	return program
}

func mustEvalBenchProgram(b *testing.B, source string, program *ast.Program) interpreter.Value {
	b.Helper()
	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, "<bench>")
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		b.Fatalf("eval error: %v", err)
	}
	if sig != nil {
		b.Fatalf("unexpected signal: %v", sig.Type)
	}
	return val
}

func mustWriteBenchFixture(b *testing.B, path string, bytes int64) {
	b.Helper()
	f, err := os.Create(path)
	if err != nil {
		b.Fatalf("create fixture: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	chunk := make([]byte, 1024*1024) // 1 MiB deterministic pattern
	for i := range chunk {
		chunk[i] = byte(i % 251)
	}

	var written int64
	for written < bytes {
		n := int64(len(chunk))
		if remaining := bytes - written; remaining < n {
			n = remaining
		}
		if _, err := f.Write(chunk[:int(n)]); err != nil {
			b.Fatalf("write fixture: %v", err)
		}
		written += n
	}
	if err := f.Sync(); err != nil {
		b.Fatalf("sync fixture: %v", err)
	}
}
