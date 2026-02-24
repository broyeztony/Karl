package dap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

type request struct {
	Seq       int             `json:"seq"`
	Type      string          `json:"type"`
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type response struct {
	Seq        int         `json:"seq"`
	Type       string      `json:"type"`
	RequestSeq int         `json:"request_seq"`
	Success    bool        `json:"success"`
	Command    string      `json:"command"`
	Message    string      `json:"message,omitempty"`
	Body       interface{} `json:"body,omitempty"`
}

type event struct {
	Seq   int         `json:"seq"`
	Type  string      `json:"type"`
	Event string      `json:"event"`
	Body  interface{} `json:"body,omitempty"`
}

type writer struct {
	mu  sync.Mutex
	out io.Writer
	seq int
}

func newWriter(out io.Writer) *writer {
	return &writer{out: out}
}

func (w *writer) send(v interface{}) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seq++
	payloadWithSeq, err := withSeq(payload, w.seq)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w.out, "Content-Length: %d\r\n\r\n", len(payloadWithSeq)); err != nil {
		return err
	}
	_, err = w.out.Write(payloadWithSeq)
	return err
}

func withSeq(payload []byte, seq int) ([]byte, error) {
	var msg map[string]interface{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	msg["seq"] = seq
	return json.Marshal(msg)
}

func readRequest(r *bufio.Reader) (*request, error) {
	headers := map[string]string{}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid dap header: %q", line)
		}
		headers[strings.ToLower(strings.TrimSpace(parts[0]))] = strings.TrimSpace(parts[1])
	}

	cl, ok := headers["content-length"]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	length, err := strconv.Atoi(cl)
	if err != nil || length < 0 {
		return nil, fmt.Errorf("invalid Content-Length: %q", cl)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	var req request
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}
	if req.Type != "request" {
		return nil, fmt.Errorf("unsupported dap message type: %s", req.Type)
	}
	return &req, nil
}
