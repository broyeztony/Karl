package dap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"karl/ast"
	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

const defaultThreadID = 1

type server struct {
	reader *bufio.Reader
	writer *writer

	mu sync.Mutex

	source      string
	filename    string
	program     *ast.Program
	controller  *interpreter.DebugController
	eval        *interpreter.Evaluator
	env         *interpreter.Environment
	started     bool
	stopOnEntry bool
	lastAction  string

	taskFailurePolicy string
	programArgs       []string

	breakpointsByFile map[string][]int
}

type launchArgs struct {
	Program           string   `json:"program"`
	Args              []string `json:"args"`
	StopOnEntry       *bool    `json:"stopOnEntry,omitempty"`
	TaskFailurePolicy string   `json:"taskFailurePolicy,omitempty"`
}

type setBreakpointsArgs struct {
	Source struct {
		Path string `json:"path"`
	} `json:"source"`
	Breakpoints []struct {
		Line int `json:"line"`
	} `json:"breakpoints,omitempty"`
	Lines []int `json:"lines,omitempty"`
}

type stackTraceArgs struct {
	StartFrame int `json:"startFrame,omitempty"`
	Levels     int `json:"levels,omitempty"`
}

type scopesArgs struct {
	FrameID int `json:"frameId"`
}

type variablesArgs struct {
	VariablesReference int `json:"variablesReference"`
}

type evaluateArgs struct {
	Expression string `json:"expression"`
	FrameID    int    `json:"frameId,omitempty"`
}

func Run(stdin io.Reader, stdout io.Writer) error {
	s := &server{
		reader:            bufio.NewReader(stdin),
		writer:            newWriter(stdout),
		stopOnEntry:       true,
		taskFailurePolicy: interpreter.TaskFailurePolicyFailFast,
		breakpointsByFile: map[string][]int{},
	}
	return s.serve()
}

func (s *server) serve() error {
	for {
		req, err := readRequest(s.reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		shouldExit := s.handle(req)
		if shouldExit {
			return nil
		}
	}
}

func (s *server) handle(req *request) bool {
	switch req.Command {
	case "initialize":
		body := map[string]interface{}{
			"supportsConfigurationDoneRequest": true,
			"supportsEvaluateForHovers":        true,
			"supportsStepOut":                  true,
		}
		s.respondOK(req, body)
		s.emit("initialized", map[string]interface{}{})
		return false
	case "launch":
		var args launchArgs
		if err := unmarshalArgs(req.Arguments, &args); err != nil {
			s.respondErr(req, err)
			return false
		}
		if err := s.prepareLaunch(args); err != nil {
			s.respondErr(req, err)
			return false
		}
		s.respondOK(req, map[string]interface{}{})
		return false
	case "setBreakpoints":
		body, err := s.setBreakpoints(req.Arguments)
		if err != nil {
			s.respondErr(req, err)
			return false
		}
		s.respondOK(req, body)
		return false
	case "setExceptionBreakpoints":
		s.respondOK(req, map[string]interface{}{})
		return false
	case "configurationDone":
		if err := s.startProgram(); err != nil {
			s.respondErr(req, err)
			return false
		}
		s.respondOK(req, map[string]interface{}{})
		return false
	case "threads":
		s.respondOK(req, map[string]interface{}{
			"threads": []map[string]interface{}{
				{"id": defaultThreadID, "name": "main"},
			},
		})
		return false
	case "stackTrace":
		body, err := s.stackTrace(req.Arguments)
		if err != nil {
			s.respondErr(req, err)
			return false
		}
		s.respondOK(req, body)
		return false
	case "scopes":
		body, err := s.scopes(req.Arguments)
		if err != nil {
			s.respondErr(req, err)
			return false
		}
		s.respondOK(req, body)
		return false
	case "variables":
		body, err := s.variables(req.Arguments)
		if err != nil {
			s.respondErr(req, err)
			return false
		}
		s.respondOK(req, body)
		return false
	case "evaluate":
		body, err := s.evaluate(req.Arguments)
		if err != nil {
			s.respondErr(req, err)
			return false
		}
		s.respondOK(req, body)
		return false
	case "continue":
		s.withController(func(c *interpreter.DebugController) {
			s.lastAction = "continue"
			c.Continue()
		})
		s.respondOK(req, map[string]interface{}{"allThreadsContinued": true})
		return false
	case "next":
		s.withController(func(c *interpreter.DebugController) {
			s.lastAction = "next"
			c.StepOver()
		})
		s.respondOK(req, map[string]interface{}{})
		return false
	case "stepIn":
		s.withController(func(c *interpreter.DebugController) {
			s.lastAction = "stepIn"
			c.Step()
		})
		s.respondOK(req, map[string]interface{}{})
		return false
	case "stepOut":
		s.withController(func(c *interpreter.DebugController) {
			s.lastAction = "stepOut"
			c.StepOut()
		})
		s.respondOK(req, map[string]interface{}{})
		return false
	case "pause":
		s.withController(func(c *interpreter.DebugController) {
			s.lastAction = "pause"
			c.Pause()
		})
		s.respondOK(req, map[string]interface{}{})
		return false
	case "disconnect", "terminate":
		s.withController(func(c *interpreter.DebugController) {
			c.Quit()
		})
		s.respondOK(req, map[string]interface{}{})
		return true
	default:
		s.respondErr(req, fmt.Errorf("unsupported command: %s", req.Command))
		return false
	}
}

func (s *server) prepareLaunch(args launchArgs) error {
	if strings.TrimSpace(args.Program) == "" {
		return fmt.Errorf("launch.program is required")
	}
	data, err := os.ReadFile(args.Program)
	if err != nil {
		return fmt.Errorf("read program: %w", err)
	}
	program, err := parseProgram(data, args.Program)
	if err != nil {
		return err
	}

	stopOnEntry := true
	if args.StopOnEntry != nil {
		stopOnEntry = *args.StopOnEntry
	}
	policy := args.TaskFailurePolicy
	if policy == "" {
		policy = interpreter.TaskFailurePolicyFailFast
	}
	if policy != interpreter.TaskFailurePolicyFailFast && policy != interpreter.TaskFailurePolicyDefer {
		return fmt.Errorf("invalid taskFailurePolicy: %s", policy)
	}

	s.mu.Lock()
	s.source = string(data)
	s.filename = args.Program
	s.program = program
	s.stopOnEntry = stopOnEntry
	s.taskFailurePolicy = policy
	s.programArgs = append([]string(nil), args.Args...)
	s.started = false
	s.mu.Unlock()
	return nil
}

func (s *server) startProgram() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return nil
	}
	if s.program == nil {
		s.mu.Unlock()
		return fmt.Errorf("program not loaded; call launch first")
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(s.source, s.filename)
	if err := eval.SetTaskFailurePolicy(s.taskFailurePolicy); err != nil {
		s.mu.Unlock()
		return err
	}
	eval.SetProgramArgs(s.programArgs)
	eval.SetProgramPath(s.filename)
	eval.SetInput(os.Stdin)

	controller := interpreter.NewDebugController(s.filename)
	controller.SetStartPaused(s.stopOnEntry)
	s.controller = controller
	s.eval = eval
	s.env = interpreter.NewBaseEnvironment()
	s.started = true
	s.lastAction = ""
	s.applyBreakpointsLocked()
	eval.SetDebugger(controller)
	program := s.program
	env := s.env
	source := s.source
	filename := s.filename
	s.mu.Unlock()

	go s.watchStops()
	go func() {
		val, sig, runErr := eval.Eval(program, env)
		if runErr == nil && sig != nil {
			runErr = fmt.Errorf("break/continue outside loop")
		}
		if runErr == nil {
			runErr = eval.CheckUnhandledTaskFailures()
		}
		if runErr != nil {
			runErr = fmt.Errorf("%s", interpreter.FormatRuntimeError(runErr, source, filename))
		}
		controller.Finish(val, runErr)
	}()
	return nil
}

func (s *server) watchStops() {
	s.mu.Lock()
	controller := s.controller
	stopOnEntry := s.stopOnEntry
	s.mu.Unlock()
	if controller == nil {
		return
	}

	firstStop := true
	for {
		reason := controller.WaitForStop()
		if reason == interpreter.DebugStopDone {
			_, err, _ := controller.Result()
			exitCode := 0
			if err != nil {
				exitCode = 1
				s.emit("output", map[string]interface{}{
					"category": "stderr",
					"output":   err.Error() + "\n",
				})
			}
			s.emit("terminated", map[string]interface{}{})
			s.emit("exited", map[string]interface{}{"exitCode": exitCode})
			return
		}

		stopReason := "breakpoint"
		s.mu.Lock()
		lastAction := s.lastAction
		s.lastAction = ""
		s.mu.Unlock()
		switch {
		case firstStop && stopOnEntry:
			stopReason = "entry"
		case lastAction == "pause":
			stopReason = "pause"
		case lastAction == "stepIn" || lastAction == "next" || lastAction == "stepOut":
			stopReason = "step"
		}
		firstStop = false

		s.emit("stopped", map[string]interface{}{
			"reason":            stopReason,
			"threadId":          defaultThreadID,
			"allThreadsStopped": true,
		})
	}
}

func (s *server) setBreakpoints(raw json.RawMessage) (map[string]interface{}, error) {
	var args setBreakpointsArgs
	if err := unmarshalArgs(raw, &args); err != nil {
		return nil, err
	}
	file := strings.TrimSpace(args.Source.Path)
	if file == "" {
		return nil, fmt.Errorf("setBreakpoints.source.path is required")
	}

	lines := []int{}
	for _, bp := range args.Breakpoints {
		if bp.Line > 0 {
			lines = append(lines, bp.Line)
		}
	}
	for _, line := range args.Lines {
		if line > 0 {
			lines = append(lines, line)
		}
	}
	sort.Ints(lines)
	lines = uniqLines(lines)

	s.mu.Lock()
	s.breakpointsByFile[file] = lines
	s.applyBreakpointsLocked()
	s.mu.Unlock()

	out := make([]map[string]interface{}, 0, len(lines))
	for _, line := range lines {
		out = append(out, map[string]interface{}{
			"verified": true,
			"line":     line,
		})
	}
	return map[string]interface{}{"breakpoints": out}, nil
}

func (s *server) stackTrace(raw json.RawMessage) (map[string]interface{}, error) {
	var args stackTraceArgs
	_ = unmarshalArgs(raw, &args)

	s.mu.Lock()
	controller := s.controller
	s.mu.Unlock()
	if controller == nil {
		return map[string]interface{}{"stackFrames": []map[string]interface{}{}, "totalFrames": 0}, nil
	}

	stack := controller.Stack()
	frames := make([]map[string]interface{}, 0, len(stack)+1)
	if len(stack) == 0 {
		ev, _ := controller.CurrentEvent()
		frames = append(frames, map[string]interface{}{
			"id":     1,
			"name":   "<top-level>",
			"line":   clampLine(ev.Line),
			"column": clampColumn(ev.Column),
			"source": map[string]interface{}{"path": ev.Filename, "name": filepathBase(ev.Filename)},
		})
	} else {
		for i := 0; i < len(stack); i++ {
			frame := stack[len(stack)-1-i]
			name := frame.Name
			if strings.TrimSpace(name) == "" {
				name = "<lambda>"
			}
			frames = append(frames, map[string]interface{}{
				"id":     i + 1,
				"name":   name,
				"line":   clampLine(frame.Line),
				"column": clampColumn(frame.Column),
				"source": map[string]interface{}{"path": frame.Filename, "name": filepathBase(frame.Filename)},
			})
		}
	}

	start := args.StartFrame
	if start < 0 {
		start = 0
	}
	if start > len(frames) {
		start = len(frames)
	}
	end := len(frames)
	if args.Levels > 0 && start+args.Levels < end {
		end = start + args.Levels
	}
	return map[string]interface{}{
		"stackFrames": frames[start:end],
		"totalFrames": len(frames),
	}, nil
}

func (s *server) scopes(raw json.RawMessage) (map[string]interface{}, error) {
	var args scopesArgs
	if err := unmarshalArgs(raw, &args); err != nil {
		return nil, err
	}
	if args.FrameID <= 0 {
		return nil, fmt.Errorf("frameId must be > 0")
	}
	s.mu.Lock()
	controller := s.controller
	s.mu.Unlock()
	if controller == nil {
		return nil, fmt.Errorf("debugger not started")
	}
	if _, err := controller.EnvForFrame(args.FrameID - 1); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"scopes": []map[string]interface{}{
			{
				"name":               "Locals",
				"variablesReference": args.FrameID,
				"expensive":          false,
			},
		},
	}, nil
}

func (s *server) variables(raw json.RawMessage) (map[string]interface{}, error) {
	var args variablesArgs
	if err := unmarshalArgs(raw, &args); err != nil {
		return nil, err
	}
	if args.VariablesReference <= 0 {
		return nil, fmt.Errorf("variablesReference must be > 0")
	}
	s.mu.Lock()
	controller := s.controller
	s.mu.Unlock()
	if controller == nil {
		return nil, fmt.Errorf("debugger not started")
	}
	env, err := controller.EnvForFrame(args.VariablesReference - 1)
	if err != nil {
		return nil, err
	}
	locals := env.Snapshot()
	names := make([]string, 0, len(locals))
	for name, value := range locals {
		if _, ok := value.(*interpreter.Builtin); ok {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	vars := make([]map[string]interface{}, 0, len(names))
	for _, name := range names {
		value := locals[name]
		vars = append(vars, map[string]interface{}{
			"name":               name,
			"value":              value.Inspect(),
			"type":               string(value.Type()),
			"variablesReference": 0,
		})
	}
	return map[string]interface{}{"variables": vars}, nil
}

func (s *server) evaluate(raw json.RawMessage) (map[string]interface{}, error) {
	var args evaluateArgs
	if err := unmarshalArgs(raw, &args); err != nil {
		return nil, err
	}
	if strings.TrimSpace(args.Expression) == "" {
		return nil, fmt.Errorf("expression is required")
	}
	frameID := args.FrameID
	if frameID <= 0 {
		frameID = 1
	}
	s.mu.Lock()
	controller := s.controller
	s.mu.Unlock()
	if controller == nil {
		return nil, fmt.Errorf("debugger not started")
	}
	env, err := controller.EnvForFrame(frameID - 1)
	if err != nil {
		return nil, err
	}
	val, err := interpreter.EvalDebugExpression(args.Expression, env)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result":             val.Inspect(),
		"type":               string(val.Type()),
		"variablesReference": 0,
	}, nil
}

func (s *server) withController(fn func(c *interpreter.DebugController)) {
	s.mu.Lock()
	controller := s.controller
	s.mu.Unlock()
	if controller != nil {
		fn(controller)
	}
}

func (s *server) applyBreakpointsLocked() {
	if s.controller == nil {
		return
	}
	s.controller.ClearBreakpoints()
	for file, lines := range s.breakpointsByFile {
		for _, line := range lines {
			_, _ = s.controller.AddBreakpoint(file, line)
		}
	}
}

func (s *server) respondOK(req *request, body interface{}) {
	_ = s.writer.send(response{
		Type:       "response",
		RequestSeq: req.Seq,
		Success:    true,
		Command:    req.Command,
		Body:       body,
	})
}

func (s *server) respondErr(req *request, err error) {
	_ = s.writer.send(response{
		Type:       "response",
		RequestSeq: req.Seq,
		Success:    false,
		Command:    req.Command,
		Message:    err.Error(),
	})
}

func (s *server) emit(name string, body interface{}) {
	_ = s.writer.send(event{
		Type:  "event",
		Event: name,
		Body:  body,
	})
}

func unmarshalArgs(raw json.RawMessage, out interface{}) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func parseProgram(data []byte, filename string) (*ast.Program, error) {
	p := parser.New(lexer.New(string(data)))
	program := p.ParseProgram()
	if errs := p.ErrorsDetailed(); len(errs) > 0 {
		return nil, fmt.Errorf("%s", parser.FormatParseErrors(errs, string(data), filename))
	}
	return program, nil
}

func uniqLines(lines []int) []int {
	if len(lines) == 0 {
		return nil
	}
	out := []int{lines[0]}
	for _, line := range lines[1:] {
		if line != out[len(out)-1] {
			out = append(out, line)
		}
	}
	return out
}

func clampLine(line int) int {
	if line <= 0 {
		return 1
	}
	return line
}

func clampColumn(col int) int {
	if col <= 0 {
		return 1
	}
	return col
}

func filepathBase(path string) string {
	if path == "" {
		return "<unknown>"
	}
	idx := strings.LastIndexAny(path, `/\`)
	if idx == -1 || idx == len(path)-1 {
		return path
	}
	return path[idx+1:]
}
