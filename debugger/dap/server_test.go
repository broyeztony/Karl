package dap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRunInitializeThenDisconnect(t *testing.T) {
	client, done := newTestClient(t)
	defer done()

	if resp := client.request("initialize", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("initialize failed: %v", responseMessage(resp))
	}
	if event := client.waitEvent("initialized"); event == nil {
		t.Fatalf("expected initialized event")
	}
	if resp := client.request("disconnect", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("disconnect failed: %v", responseMessage(resp))
	}
}

func TestContinueBeforeLaunchReturnsError(t *testing.T) {
	client, done := newTestClient(t)
	defer done()

	if resp := client.request("initialize", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("initialize failed: %v", responseMessage(resp))
	}
	resp := client.request("continue", map[string]interface{}{})
	if responseSuccess(resp) {
		t.Fatalf("expected continue to fail before launch")
	}
	if !strings.Contains(responseMessage(resp), "launch/configurationDone") {
		t.Fatalf("unexpected error message: %q", responseMessage(resp))
	}
	_ = client.request("disconnect", map[string]interface{}{})
}

func TestBreakpointEvaluateAndNestedVariables(t *testing.T) {
	tmpDir := t.TempDir()
	programPath := filepath.Join(tmpDir, "debuggee.k")
	source := strings.Join([]string{
		"let user = { name: \"Ada\", nums: [1, 2] }",
		"let probe = user",
		"let x = 41",
		"let y = x + 1",
		"y",
	}, "\n")
	if err := os.WriteFile(programPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write program: %v", err)
	}

	client, done := newTestClient(t)
	defer done()

	if resp := client.request("initialize", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("initialize failed: %v", responseMessage(resp))
	}
	client.waitEvent("initialized")

	launchResp := client.request("launch", map[string]interface{}{
		"program":     programPath,
		"stopOnEntry": false,
		"args":        []string{"a", "b"},
	})
	if !responseSuccess(launchResp) {
		t.Fatalf("launch failed: %v", responseMessage(launchResp))
	}

	bpResp := client.request("setBreakpoints", map[string]interface{}{
		"source": map[string]interface{}{"path": filepath.Base(programPath)},
		"breakpoints": []map[string]interface{}{
			{"line": 4},
		},
	})
	if !responseSuccess(bpResp) {
		t.Fatalf("setBreakpoints failed: %v", responseMessage(bpResp))
	}

	if resp := client.request("configurationDone", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("configurationDone failed: %v", responseMessage(resp))
	}
	stopped := client.waitEvent("stopped")
	if stopped == nil {
		t.Fatalf("expected stopped event")
	}
	stopBody := bodyMap(stopped)
	if reason, _ := stopBody["reason"].(string); reason != "breakpoint" {
		t.Fatalf("unexpected stop reason: %v", reason)
	}

	stackResp := client.request("stackTrace", map[string]interface{}{"threadId": defaultThreadID})
	if !responseSuccess(stackResp) {
		t.Fatalf("stackTrace failed: %v", responseMessage(stackResp))
	}
	stackFrames := bodyArrayOfMaps(stackResp, "stackFrames")
	if len(stackFrames) == 0 {
		t.Fatalf("expected non-empty stack")
	}
	topFrame := stackFrames[0]
	frameID := intFromAny(topFrame["id"])
	if frameID <= 0 {
		t.Fatalf("invalid frame id: %v", topFrame["id"])
	}
	if line := intFromAny(topFrame["line"]); line != 4 {
		t.Fatalf("expected top frame line 4, got %d", line)
	}

	scopesResp := client.request("scopes", map[string]interface{}{"frameId": frameID})
	if !responseSuccess(scopesResp) {
		t.Fatalf("scopes failed: %v", responseMessage(scopesResp))
	}
	scopes := bodyArrayOfMaps(scopesResp, "scopes")
	if len(scopes) == 0 {
		t.Fatalf("expected locals scope")
	}
	localsRef := intFromAny(scopes[0]["variablesReference"])
	if localsRef <= 0 {
		t.Fatalf("invalid locals ref: %v", scopes[0]["variablesReference"])
	}

	varsResp := client.request("variables", map[string]interface{}{"variablesReference": localsRef})
	if !responseSuccess(varsResp) {
		t.Fatalf("variables failed: %v", responseMessage(varsResp))
	}
	locals := bodyArrayOfMaps(varsResp, "variables")
	xVar := findVarByName(locals, "x")
	if xVar == nil || xVar["value"] != "41" {
		t.Fatalf("expected x=41 in locals, got: %#v", locals)
	}
	probeVar := findVarByName(locals, "probe")
	if probeVar == nil {
		t.Fatalf("expected probe in locals, got: %#v", locals)
	}
	probeRef := intFromAny(probeVar["variablesReference"])
	if probeRef <= 0 {
		t.Fatalf("expected probe to be expandable, got: %#v", probeVar)
	}

	probeFieldsResp := client.request("variables", map[string]interface{}{"variablesReference": probeRef})
	if !responseSuccess(probeFieldsResp) {
		t.Fatalf("variables(probe) failed: %v", responseMessage(probeFieldsResp))
	}
	probeFields := bodyArrayOfMaps(probeFieldsResp, "variables")
	nameField := findVarByName(probeFields, "name")
	if nameField == nil || nameField["value"] != `"Ada"` {
		t.Fatalf("expected probe.name=\"Ada\", got: %#v", probeFields)
	}
	numsField := findVarByName(probeFields, "nums")
	if numsField == nil {
		t.Fatalf("expected probe.nums field, got: %#v", probeFields)
	}
	numsRef := intFromAny(numsField["variablesReference"])
	if numsRef <= 0 {
		t.Fatalf("expected probe.nums to be expandable, got: %#v", numsField)
	}

	numsResp := client.request("variables", map[string]interface{}{"variablesReference": numsRef})
	if !responseSuccess(numsResp) {
		t.Fatalf("variables(nums) failed: %v", responseMessage(numsResp))
	}
	nums := bodyArrayOfMaps(numsResp, "variables")
	if len(nums) != 2 || nums[0]["value"] != "1" || nums[1]["value"] != "2" {
		t.Fatalf("unexpected nums: %#v", nums)
	}

	evalResp := client.request("evaluate", map[string]interface{}{
		"frameId":    frameID,
		"expression": "x + 1",
	})
	if !responseSuccess(evalResp) {
		t.Fatalf("evaluate failed: %v", responseMessage(evalResp))
	}
	evalBody := bodyMap(evalResp)
	if evalBody["result"] != "42" {
		t.Fatalf("expected evaluate result 42, got: %#v", evalBody)
	}

	continueResp := client.request("continue", map[string]interface{}{"threadId": defaultThreadID})
	if !responseSuccess(continueResp) {
		t.Fatalf("continue failed: %v", responseMessage(continueResp))
	}
	if terminated := client.waitEvent("terminated"); terminated == nil {
		t.Fatalf("expected terminated event")
	}
	if exited := client.waitEvent("exited"); exited == nil {
		t.Fatalf("expected exited event")
	}
	if resp := client.request("disconnect", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("disconnect failed: %v", responseMessage(resp))
	}
}

func TestNextStepsOverFunctionCall(t *testing.T) {
	tmpDir := t.TempDir()
	programPath := filepath.Join(tmpDir, "step_over.k")
	source := strings.Join([]string{
		"let inc = (n) -> n + 1",
		"let x = 1",
		"let y = inc(x)",
		"let z = y + 1",
		"z",
	}, "\n")
	if err := os.WriteFile(programPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write program: %v", err)
	}

	client, done := newTestClient(t)
	defer done()
	if resp := client.request("initialize", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("initialize failed: %v", responseMessage(resp))
	}
	client.waitEvent("initialized")
	if resp := client.request("launch", map[string]interface{}{
		"program":     programPath,
		"stopOnEntry": false,
	}); !responseSuccess(resp) {
		t.Fatalf("launch failed: %v", responseMessage(resp))
	}
	if resp := client.request("setBreakpoints", map[string]interface{}{
		"source": map[string]interface{}{"path": programPath},
		"breakpoints": []map[string]interface{}{
			{"line": 3},
		},
	}); !responseSuccess(resp) {
		t.Fatalf("setBreakpoints failed: %v", responseMessage(resp))
	}
	if resp := client.request("configurationDone", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("configurationDone failed: %v", responseMessage(resp))
	}
	if event := client.waitEvent("stopped"); event == nil {
		t.Fatalf("expected first stop event")
	}
	lineBefore := currentTopFrameLine(client)
	if lineBefore != 3 {
		t.Fatalf("expected first stop at line 3, got %d", lineBefore)
	}
	if resp := client.request("next", map[string]interface{}{"threadId": defaultThreadID}); !responseSuccess(resp) {
		t.Fatalf("next failed: %v", responseMessage(resp))
	}
	stepEvent := client.waitEvent("stopped")
	if stepEvent == nil {
		t.Fatalf("expected stop after next")
	}
	reason, _ := bodyMap(stepEvent)["reason"].(string)
	if reason != "step" {
		t.Fatalf("expected step reason after next, got %q", reason)
	}
	lineAfter := currentTopFrameLine(client)
	if lineAfter != 4 {
		t.Fatalf("expected next to stop at line 4, got %d", lineAfter)
	}
	if resp := client.request("continue", map[string]interface{}{"threadId": defaultThreadID}); !responseSuccess(resp) {
		t.Fatalf("continue failed: %v", responseMessage(resp))
	}
	client.waitEvent("terminated")
	client.waitEvent("exited")
	if resp := client.request("disconnect", map[string]interface{}{}); !responseSuccess(resp) {
		t.Fatalf("disconnect failed: %v", responseMessage(resp))
	}
}

type testClient struct {
	t         *testing.T
	in        *io.PipeWriter
	done      <-chan error
	nextSeq   int
	pending   []map[string]interface{}
	readAfter time.Duration
	messages  chan map[string]interface{}
	readErr   chan error
}

func newTestClient(t *testing.T) (*testClient, func()) {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	done := make(chan error, 1)
	go func() {
		err := Run(inR, outW)
		_ = outW.Close()
		done <- err
	}()
	client := &testClient{
		t:         t,
		in:        inW,
		done:      done,
		nextSeq:   1,
		readAfter: 3 * time.Second,
		messages:  make(chan map[string]interface{}, 32),
		readErr:   make(chan error, 1),
	}
	go client.readLoop(outR)
	cleanup := func() {
		_ = inW.Close()
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("dap server error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for dap server to stop")
		}
	}
	return client, cleanup
}

func (c *testClient) readLoop(outR *io.PipeReader) {
	reader := bufio.NewReader(outR)
	for {
		msg, err := readFramedJSON(reader)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "closed pipe") {
				return
			}
			select {
			case c.readErr <- err:
			default:
			}
			return
		}
		c.messages <- msg
	}
}

func (c *testClient) request(command string, args map[string]interface{}) map[string]interface{} {
	c.t.Helper()
	seq := c.nextSeq
	c.nextSeq++
	payload := map[string]interface{}{
		"seq":     seq,
		"type":    "request",
		"command": command,
	}
	if args != nil {
		payload["arguments"] = args
	}
	c.writeMessage(payload)
	for {
		msg := c.nextMessage()
		msgType, _ := msg["type"].(string)
		if msgType == "event" {
			c.pending = append(c.pending, msg)
			continue
		}
		reqSeq := intFromAny(msg["request_seq"])
		if reqSeq != seq {
			c.t.Fatalf("received response for unexpected seq %d (expected %d): %#v", reqSeq, seq, msg)
		}
		return msg
	}
}

func (c *testClient) waitEvent(name string) map[string]interface{} {
	c.t.Helper()
	for i := 0; i < len(c.pending); i++ {
		event := c.pending[i]
		if eventName, _ := event["event"].(string); eventName == name {
			c.pending = append(c.pending[:i], c.pending[i+1:]...)
			return event
		}
	}
	for {
		msg := c.nextMessage()
		msgType, _ := msg["type"].(string)
		if msgType != "event" {
			c.t.Fatalf("expected event %q, got non-event message: %#v", name, msg)
		}
		eventName, _ := msg["event"].(string)
		if eventName == name {
			return msg
		}
		c.pending = append(c.pending, msg)
	}
}

func (c *testClient) writeMessage(payload map[string]interface{}) {
	c.t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		c.t.Fatalf("marshal payload: %v", err)
	}
	if _, err := fmt.Fprintf(c.in, "Content-Length: %d\r\n\r\n", len(raw)); err != nil {
		c.t.Fatalf("write header: %v", err)
	}
	if _, err := c.in.Write(raw); err != nil {
		c.t.Fatalf("write payload: %v", err)
	}
}

func (c *testClient) nextMessage() map[string]interface{} {
	c.t.Helper()
	select {
	case err := <-c.readErr:
		c.t.Fatalf("read dap message: %v", err)
		return nil
	case msg := <-c.messages:
		return msg
	case <-time.After(c.readAfter):
		c.t.Fatalf("timeout waiting for dap message")
		return nil
	}
}

func readFramedJSON(r *bufio.Reader) (map[string]interface{}, error) {
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
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}
	contentLengthText, ok := headers["content-length"]
	if !ok {
		return nil, fmt.Errorf("missing content-length header")
	}
	contentLength, err := strconv.Atoi(contentLengthText)
	if err != nil || contentLength < 0 {
		return nil, fmt.Errorf("invalid content-length: %q", contentLengthText)
	}
	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	msg := map[string]interface{}{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func responseSuccess(msg map[string]interface{}) bool {
	success, _ := msg["success"].(bool)
	return success
}

func responseMessage(msg map[string]interface{}) string {
	value, _ := msg["message"].(string)
	return value
}

func bodyMap(msg map[string]interface{}) map[string]interface{} {
	body, _ := msg["body"].(map[string]interface{})
	if body == nil {
		return map[string]interface{}{}
	}
	return body
}

func bodyArrayOfMaps(msg map[string]interface{}, key string) []map[string]interface{} {
	body := bodyMap(msg)
	raw, _ := body[key].([]interface{})
	out := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}

func intFromAny(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func findVarByName(vars []map[string]interface{}, name string) map[string]interface{} {
	for _, variable := range vars {
		if variable["name"] == name {
			return variable
		}
	}
	return nil
}

func currentTopFrameLine(client *testClient) int {
	resp := client.request("stackTrace", map[string]interface{}{"threadId": defaultThreadID})
	if !responseSuccess(resp) {
		client.t.Fatalf("stackTrace failed: %v", responseMessage(resp))
	}
	frames := bodyArrayOfMaps(resp, "stackFrames")
	if len(frames) == 0 {
		client.t.Fatalf("expected stack frame")
	}
	return intFromAny(frames[0]["line"])
}
