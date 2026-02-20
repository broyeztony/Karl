package dap

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestRunInitializeThenDisconnect(t *testing.T) {
	input := bytes.NewBuffer(nil)
	writeRequest(t, input, `{"seq":1,"type":"request","command":"initialize","arguments":{}}`)
	writeRequest(t, input, `{"seq":2,"type":"request","command":"disconnect","arguments":{}}`)

	output := bytes.NewBuffer(nil)
	if err := Run(input, output); err != nil {
		t.Fatalf("run dap: %v", err)
	}

	text := output.String()
	if !strings.Contains(text, `"command":"initialize"`) {
		t.Fatalf("expected initialize response in output, got: %s", text)
	}
	if !strings.Contains(text, `"event":"initialized"`) {
		t.Fatalf("expected initialized event in output, got: %s", text)
	}
	if !strings.Contains(text, `"command":"disconnect"`) {
		t.Fatalf("expected disconnect response in output, got: %s", text)
	}
}

func writeRequest(t *testing.T, out *bytes.Buffer, payload string) {
	t.Helper()
	if _, err := fmt.Fprintf(out, "Content-Length: %d\r\n\r\n%s", len(payload), payload); err != nil {
		t.Fatalf("write request: %v", err)
	}
}
