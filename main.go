package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"karl/ast"
	"karl/interpreter"
	"karl/kernel"
	"karl/lexer"
	"karl/notebook"
	"karl/parser"
	"karl/playground"
	"karl/repl"
	"karl/spreadsheet"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	sub := os.Args[1]
	switch sub {
	case "-h", "--help", "help":
		usage()
		return
	case "parse":
		os.Exit(parseCommand(os.Args[2:]))
	case "run":
		os.Exit(runCommand(os.Args[2:]))
	case "debug":
		os.Exit(debugCommand(os.Args[2:]))
	case "loom":
		os.Exit(loomCommand(os.Args[2:]))
	case "repl":
		os.Exit(replCommand(os.Args[2:]))
	case "repl-server":
		os.Exit(replServerCommand(os.Args[2:]))
	case "repl-client":
		os.Exit(replClientCommand(os.Args[2:]))
	case "notebook", "nb":
		os.Exit(notebookCommand(os.Args[2:]))

	case "kernel": // Added kernel subcommand
		os.Exit(kernelCommand(os.Args[2:]))
	case "spreadsheet":
		os.Exit(spreadsheetCommand(os.Args[2:]))
	case "playground":
		os.Exit(playgroundCommand(os.Args[2:]))
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", sub)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl <command> [arguments]\n")
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  parse <file.k>           parse a file and print the AST\n")
	fmt.Fprintf(os.Stderr, "  run <file.k>             run a file using the interpreter (program args after --)\n")
	fmt.Fprintf(os.Stderr, "  debug <file.k>           run a file with the CLI debugger\n")
	fmt.Fprintf(os.Stderr, "  loom <file.k>            run a file using the Loom runtime\n")
	fmt.Fprintf(os.Stderr, "  repl                     start the REPL\n")
	fmt.Fprintf(os.Stderr, "  repl-server              start the REPL server\n")
	fmt.Fprintf(os.Stderr, "  repl-client              start the REPL client\n")
	fmt.Fprintf(os.Stderr, "  notebook <file.knb>      run a notebook\n")
	fmt.Fprintf(os.Stderr, "  notebook convert <in.ipynb> <out.knb> convert Jupyter notebook to Karl notebook\n")
	fmt.Fprintf(os.Stderr, "  kernel <connection_file> start Jupyter kernel\n")
	fmt.Fprintf(os.Stderr, "  spreadsheet [addr]       start reactive spreadsheet server (default :8080)\n")
	fmt.Fprintf(os.Stderr, "  playground [addr]        start WASM playground server (default :8081)\n")
	fmt.Fprintf(os.Stderr, "  help                     show this help message\n")
}

func parseCommand(args []string) int {
	format, positional, help, err := parseParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		parseUsage()
		return 2
	}
	if help {
		parseUsage()
		return 0
	}
	if len(positional) != 1 {
		parseUsage()
		return 2
	}
	if err := validateExtension(positional[0]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 2
	}
	data, err := readInput(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		return 1
	}
	program, err := parseProgram(data, displayName(positional[0]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	switch format {
	case "pretty":
		fmt.Print(ast.Format(program))
	case "json":
		out, err := ast.FormatJSON(program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "format error: %v\n", err)
			return 1
		}
		fmt.Print(out)
	default:
		fmt.Fprintf(os.Stderr, "unknown format: %s\n", format)
		return 2
	}
	return 0
}

func runCommand(args []string) int {
	taskFailurePolicy, positional, programArgs, help, err := parseRunArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		runUsage()
		return 2
	}
	if help {
		runUsage()
		return 0
	}
	if len(positional) == 0 {
		runUsage()
		return 2
	}
	if len(positional) > 1 {
		fmt.Fprintf(os.Stderr, "program args must follow `--`\n")
		runUsage()
		return 2
	}
	if err := validateExtension(positional[0]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 2
	}
	data, err := readInput(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		return 1
	}
	filename := displayName(positional[0])
	program, err := parseProgram(data, filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	val, err := runProgram(program, string(data), filename, taskFailurePolicy, programArgs)
	if err != nil {
		if ute, ok := err.(*interpreter.UnhandledTaskError); ok {
			fmt.Fprintln(os.Stderr, ute.Error())
			return 1
		}
		fmt.Fprintln(os.Stderr, interpreter.FormatRuntimeError(err, string(data), filename))
		return 1
	}
	if _, ok := val.(*interpreter.Unit); !ok {
		fmt.Println(val.Inspect())
	}
	return 0
}

func debugCommand(args []string) int {
	taskFailurePolicy, positional, programArgs, help, err := parseDebugArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		debugUsage()
		return 2
	}
	if help {
		debugUsage()
		return 0
	}
	if len(positional) == 0 {
		debugUsage()
		return 2
	}
	if len(positional) > 1 {
		fmt.Fprintf(os.Stderr, "program args must follow `--`\n")
		debugUsage()
		return 2
	}
	if err := validateExtension(positional[0]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 2
	}

	data, err := readInput(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		return 1
	}
	filename := displayName(positional[0])
	program, err := parseProgram(data, filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(string(data), filename)
	if err := eval.SetTaskFailurePolicy(taskFailurePolicy); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	eval.SetProgramArgs(programArgs)
	eval.SetProgramPath(filename)
	eval.SetInput(os.Stdin)

	controller := interpreter.NewDebugController(filename)
	eval.SetDebugger(controller)

	env := interpreter.NewBaseEnvironment()
	done := make(chan struct{})
	go func() {
		val, sig, runErr := eval.Eval(program, env)
		if runErr == nil && sig != nil {
			runErr = fmt.Errorf("break/continue outside loop")
		}
		if runErr == nil {
			runErr = eval.CheckUnhandledTaskFailures()
		}
		controller.Finish(val, runErr)
		close(done)
	}()

	fmt.Fprintf(os.Stdout, "Karl Debugger\n")
	printDebugCommands(os.Stdout)

	lines := strings.Split(string(data), "\n")
	reader := bufio.NewReader(os.Stdin)
	state := &debugSessionState{selectedFrame: 0}
	evalInFrame := func(expr string, frameIndex int) (interpreter.Value, error) {
		env, frameErr := controller.EnvForFrame(frameIndex)
		if frameErr != nil {
			return nil, frameErr
		}
		return evalDebugExpression(expr, env)
	}

	for {
		reason := controller.WaitForStop()
		if reason == interpreter.DebugStopDone {
			break
		}
		state.selectedFrame = 0
		printPause(controller, lines)
		printWatchValues(controller, state, evalInFrame)

		for controller.IsPaused() {
			fmt.Fprint(os.Stdout, "karl(debug)> ")
			line, readErr := reader.ReadString('\n')
			if readErr != nil && readErr != io.EOF {
				fmt.Fprintf(os.Stderr, "debugger input error: %v\n", readErr)
				controller.Quit()
				break
			}
			cmd := strings.TrimSpace(line)
			if cmd == "" && readErr == io.EOF {
				controller.Quit()
				break
			}
			if cmd == "" {
				if readErr == io.EOF {
					controller.Quit()
				}
				continue
			}
			if handleDebugCommand(cmd, controller, state, evalInFrame) {
				break
			}
			if readErr == io.EOF {
				controller.Quit()
				break
			}
		}
	}
	<-done

	val, runErr, _ := controller.Result()
	if runErr != nil {
		if interpreter.IsDebugTerminated(runErr) {
			return 0
		}
		if ute, ok := runErr.(*interpreter.UnhandledTaskError); ok {
			fmt.Fprintln(os.Stderr, ute.Error())
			return 1
		}
		fmt.Fprintln(os.Stderr, interpreter.FormatRuntimeError(runErr, string(data), filename))
		return 1
	}
	if _, ok := val.(*interpreter.Unit); !ok {
		fmt.Fprintf(os.Stdout, "result: %s\n", val.Inspect())
	}
	return 0
}

func debugUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl debug <file.k> [--task-failure-policy=fail-fast|defer] [-- <program args...>]\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
	fmt.Fprintf(os.Stderr, "  program args are only accepted after `--`\n")
	fmt.Fprintf(os.Stderr, "\nDebugger commands:\n")
	printDebugHelp(os.Stderr)
}

func parseUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl parse <file.k> [--format=pretty|json]\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --format string   output format: pretty|json (default \"pretty\")\n")
}

func runUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl run <file.k> [--task-failure-policy=fail-fast|defer] [-- <program args...>]\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
	fmt.Fprintf(os.Stderr, "  program args are only accepted after `--`\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --task-failure-policy string   task failure behavior: fail-fast|defer (default \"fail-fast\")\n")
}

func parseDebugArgs(args []string) (string, []string, []string, bool, error) {
	return parseRunArgs(args)
}

func printPause(controller *interpreter.DebugController, lines []string) {
	event, ok := controller.CurrentEvent()
	if !ok {
		return
	}
	if event.Line > 0 {
		fmt.Fprintf(os.Stdout, "\npaused at %s:%d:%d (%s)\n", event.Filename, event.Line, event.Column, event.NodeType)
		if event.Line-1 < len(lines) {
			fmt.Fprintf(os.Stdout, "  %d | %s\n", event.Line, lines[event.Line-1])
		}
	} else {
		fmt.Fprintf(os.Stdout, "\npaused (%s)\n", event.NodeType)
	}
}

type debugSessionState struct {
	selectedFrame int
	watches       []string
}

func handleDebugCommand(cmd string, controller *interpreter.DebugController, state *debugSessionState, evalExpr func(string, int) (interpreter.Value, error)) bool {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return false
	}
	switch fields[0] {
	case "c", "continue":
		controller.Continue()
		return true
	case "s", "step":
		controller.Step()
		return true
	case "n", "next":
		controller.StepOver()
		return true
	case "f", "finish":
		controller.StepOut()
		return true
	case "q", "quit":
		controller.Quit()
		return true
	case "h", "help":
		printDebugHelp(os.Stdout)
		return false
	case "break":
		if len(fields) != 2 {
			fmt.Fprintf(os.Stdout, "usage: break <line|file:line>\n")
			return false
		}
		file, line, err := parseBreakpointSpec(fields[1])
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", err.Error())
			return false
		}
		id, err := controller.AddBreakpoint(file, line)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", err.Error())
			return false
		}
		if strings.TrimSpace(file) == "" {
			fmt.Fprintf(os.Stdout, "breakpoint #%d set at line %d\n", id, line)
		} else {
			fmt.Fprintf(os.Stdout, "breakpoint #%d set at %s:%d\n", id, file, line)
		}
		return false
	case "delete":
		if len(fields) != 2 {
			fmt.Fprintf(os.Stdout, "usage: delete <id>\n")
			return false
		}
		id, err := strconv.Atoi(fields[1])
		if err != nil {
			fmt.Fprintf(os.Stdout, "invalid breakpoint id: %s\n", fields[1])
			return false
		}
		if err := controller.RemoveBreakpoint(id); err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", err.Error())
			return false
		}
		fmt.Fprintf(os.Stdout, "removed breakpoint #%d\n", id)
		return false
	case "clear":
		removed := controller.ClearBreakpoints()
		if removed == 0 {
			fmt.Fprintf(os.Stdout, "(no breakpoints)\n")
			return false
		}
		fmt.Fprintf(os.Stdout, "cleared %d breakpoint(s)\n", removed)
		return false
	case "breakpoints":
		bps := controller.Breakpoints()
		if len(bps) == 0 {
			fmt.Fprintf(os.Stdout, "(no breakpoints)\n")
			return false
		}
		sort.Slice(bps, func(i, j int) bool {
			return bps[i].ID < bps[j].ID
		})
		for _, bp := range bps {
			fmt.Fprintf(os.Stdout, "#%d %s:%d\n", bp.ID, bp.File, bp.Line)
		}
		return false
	case "stack":
		stack := controller.Stack()
		if len(stack) == 0 {
			event, _ := controller.CurrentEvent()
			marker := " "
			if state != nil && state.selectedFrame == 0 {
				marker = "*"
			}
			fmt.Fprintf(os.Stdout, "%s#0 <top-level> at %s:%d:%d\n", marker, event.Filename, event.Line, event.Column)
			return false
		}
		for i := 0; i < len(stack); i++ {
			frame := stack[len(stack)-1-i]
			marker := " "
			if state != nil && state.selectedFrame == i {
				marker = "*"
			}
			fmt.Fprintf(os.Stdout, "%s#%d %s at %s:%d:%d\n", marker, i, frame.Name, frame.Filename, frame.Line, frame.Column)
		}
		return false
	case "frame":
		if state == nil {
			fmt.Fprintf(os.Stdout, "frame selection unavailable\n")
			return false
		}
		if len(fields) != 2 {
			fmt.Fprintf(os.Stdout, "usage: frame <idx>\n")
			return false
		}
		idx, err := strconv.Atoi(fields[1])
		if err != nil {
			fmt.Fprintf(os.Stdout, "invalid frame index: %s\n", fields[1])
			return false
		}
		if _, err := controller.EnvForFrame(idx); err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", err.Error())
			return false
		}
		state.selectedFrame = idx
		fmt.Fprintf(os.Stdout, "selected frame #%d\n", idx)
		return false
	case "locals":
		showBuiltins := len(fields) >= 2 && fields[1] == "all"
		if len(fields) >= 2 && fields[1] != "all" {
			fmt.Fprintf(os.Stdout, "usage: locals [all]\n")
			return false
		}
		frameIndex := 0
		if state != nil {
			frameIndex = state.selectedFrame
		}
		env, err := controller.EnvForFrame(frameIndex)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", err.Error())
			return false
		}
		locals := env.Snapshot()
		if len(locals) == 0 {
			fmt.Fprintf(os.Stdout, "(no locals)\n")
			return false
		}
		names := make([]string, 0, len(locals))
		for name := range locals {
			if !showBuiltins {
				if _, ok := locals[name].(*interpreter.Builtin); ok {
					continue
				}
			}
			names = append(names, name)
		}
		if len(names) == 0 {
			if showBuiltins {
				fmt.Fprintf(os.Stdout, "(no locals)\n")
			} else {
				fmt.Fprintf(os.Stdout, "(no user locals; use `locals all`)\n")
			}
			return false
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintf(os.Stdout, "%s = %s\n", name, locals[name].Inspect())
		}
		return false
	case "p", "print":
		if evalExpr == nil {
			fmt.Fprintf(os.Stdout, "print is not available in this debug session\n")
			return false
		}
		expr := strings.TrimSpace(strings.TrimPrefix(cmd, fields[0]))
		if expr == "" {
			fmt.Fprintf(os.Stdout, "usage: print <expr>\n")
			return false
		}
		frameIndex := 0
		if state != nil {
			frameIndex = state.selectedFrame
		}
		val, err := evalExpr(expr, frameIndex)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", err.Error())
			return false
		}
		fmt.Fprintf(os.Stdout, "%s\n", val.Inspect())
		return false
	case "watch":
		if state == nil {
			fmt.Fprintf(os.Stdout, "watch unavailable\n")
			return false
		}
		expr := strings.TrimSpace(strings.TrimPrefix(cmd, fields[0]))
		if expr == "" {
			fmt.Fprintf(os.Stdout, "usage: watch <expr>\n")
			return false
		}
		state.watches = append(state.watches, expr)
		fmt.Fprintf(os.Stdout, "watch #%d: %s\n", len(state.watches), expr)
		return false
	case "unwatch":
		if state == nil {
			fmt.Fprintf(os.Stdout, "unwatch unavailable\n")
			return false
		}
		if len(fields) != 2 {
			fmt.Fprintf(os.Stdout, "usage: unwatch <id>\n")
			return false
		}
		id, err := strconv.Atoi(fields[1])
		if err != nil || id <= 0 || id > len(state.watches) || strings.TrimSpace(state.watches[id-1]) == "" {
			fmt.Fprintf(os.Stdout, "invalid watch id: %s\n", fields[1])
			return false
		}
		state.watches[id-1] = ""
		fmt.Fprintf(os.Stdout, "removed watch #%d\n", id)
		return false
	default:
		fmt.Fprintf(os.Stdout, "unknown command: %s\n", fields[0])
		printDebugCommands(os.Stdout)
		return false
	}
}

func printDebugCommands(w io.Writer) {
	fmt.Fprintf(w, "Commands: break <line|file:line>, delete <id>, clear, breakpoints, continue (c), step (s), next (n), finish (f), stack, frame, locals, print (p), watch, unwatch, help (h), quit (q)\n")
}

func printDebugHelp(w io.Writer) {
	fmt.Fprintf(w, "  break <line|file:line>   add breakpoint\n")
	fmt.Fprintf(w, "  delete <id>              remove breakpoint by id\n")
	fmt.Fprintf(w, "  clear                    remove all breakpoints\n")
	fmt.Fprintf(w, "  breakpoints              list active breakpoints\n")
	fmt.Fprintf(w, "  continue | c             resume execution\n")
	fmt.Fprintf(w, "  step | s                 stop at next evaluation step\n")
	fmt.Fprintf(w, "  next | n                 step over function calls\n")
	fmt.Fprintf(w, "  finish | f               step out of current function\n")
	fmt.Fprintf(w, "  stack                    show function stack\n")
	fmt.Fprintf(w, "  frame <idx>              select stack frame by index (0 = top)\n")
	fmt.Fprintf(w, "  locals [all]             show local bindings (default hides builtins)\n")
	fmt.Fprintf(w, "  print | p <expr>         evaluate expression in current frame\n")
	fmt.Fprintf(w, "  watch <expr>             add expression watch\n")
	fmt.Fprintf(w, "  unwatch <id>             remove watch by id\n")
	fmt.Fprintf(w, "  help | h                 show debugger commands\n")
	fmt.Fprintf(w, "  quit | q                 terminate debugging session\n")
}

func printWatchValues(controller *interpreter.DebugController, state *debugSessionState, evalExpr func(string, int) (interpreter.Value, error)) {
	if state == nil || evalExpr == nil {
		return
	}
	hasWatch := false
	for _, expr := range state.watches {
		if strings.TrimSpace(expr) != "" {
			hasWatch = true
			break
		}
	}
	if !hasWatch {
		return
	}
	fmt.Fprintf(os.Stdout, "watch:\n")
	for i, expr := range state.watches {
		expr = strings.TrimSpace(expr)
		if expr == "" {
			continue
		}
		val, err := evalExpr(expr, state.selectedFrame)
		if err != nil {
			fmt.Fprintf(os.Stdout, "  #%d %s = <error: %s>\n", i+1, expr, err.Error())
			continue
		}
		fmt.Fprintf(os.Stdout, "  #%d %s = %s\n", i+1, expr, val.Inspect())
	}
}

func evalDebugExpression(input string, env *interpreter.Environment) (interpreter.Value, error) {
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.ErrorsDetailed(); len(errs) > 0 {
		return nil, fmt.Errorf("%s", parser.FormatParseErrors(errs, input, "<debug>"))
	}
	if len(program.Statements) != 1 {
		return nil, fmt.Errorf("print expects a single expression")
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		return nil, fmt.Errorf("print expects an expression, got statement")
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(input, "<debug>")
	val, sig, err := eval.Eval(stmt.Expression, env)
	if err != nil {
		return nil, err
	}
	if sig != nil {
		return nil, fmt.Errorf("print expression cannot produce control flow")
	}
	return val, nil
}

func parseBreakpointSpec(spec string) (string, int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", 0, fmt.Errorf("breakpoint spec is empty")
	}
	if !strings.Contains(spec, ":") {
		line, err := strconv.Atoi(spec)
		if err != nil {
			return "", 0, fmt.Errorf("invalid line number: %s", spec)
		}
		return "", line, nil
	}
	idx := strings.LastIndex(spec, ":")
	file := spec[:idx]
	lineText := spec[idx+1:]
	line, err := strconv.Atoi(lineText)
	if err != nil {
		return "", 0, fmt.Errorf("invalid line number: %s", lineText)
	}
	if strings.TrimSpace(file) == "" {
		return "", 0, fmt.Errorf("missing file in breakpoint spec: %s", spec)
	}
	return file, line, nil
}

func parseParseArgs(args []string) (string, []string, bool, error) {
	format := "pretty"
	positional := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return format, positional, true, nil
		case arg == "-":
			positional = append(positional, arg)
		case strings.HasPrefix(arg, "--format="):
			format = strings.TrimPrefix(arg, "--format=")
		case arg == "--format":
			if i+1 >= len(args) {
				return format, positional, false, fmt.Errorf("--format requires a value")
			}
			format = args[i+1]
			i++
		case strings.HasPrefix(arg, "-"):
			return format, positional, false, fmt.Errorf("unknown flag: %s", arg)
		default:
			positional = append(positional, arg)
		}
	}

	return format, positional, false, nil
}

func parseRunArgs(args []string) (string, []string, []string, bool, error) {
	taskFailurePolicy := interpreter.TaskFailurePolicyFailFast
	positional := []string{}
	programArgs := []string{}
	separatorSeen := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return taskFailurePolicy, positional, programArgs, true, nil
		case arg == "--":
			separatorSeen = true
			programArgs = append(programArgs, args[i+1:]...)
			i = len(args)
		case strings.HasPrefix(arg, "--task-failure-policy="):
			taskFailurePolicy = strings.TrimPrefix(arg, "--task-failure-policy=")
		case arg == "--task-failure-policy":
			if i+1 >= len(args) {
				return taskFailurePolicy, positional, programArgs, false, fmt.Errorf("--task-failure-policy requires a value")
			}
			taskFailurePolicy = args[i+1]
			i++
		case arg == "-":
			positional = append(positional, arg)
		case strings.HasPrefix(arg, "-"):
			return taskFailurePolicy, positional, programArgs, false, fmt.Errorf("unknown flag: %s", arg)
		default:
			positional = append(positional, arg)
		}
	}
	if !separatorSeen && len(positional) > 1 {
		return taskFailurePolicy, positional, programArgs, false, fmt.Errorf("program args must follow `--`")
	}
	if taskFailurePolicy != interpreter.TaskFailurePolicyFailFast && taskFailurePolicy != interpreter.TaskFailurePolicyDefer {
		return taskFailurePolicy, positional, programArgs, false, fmt.Errorf("invalid --task-failure-policy: %s", taskFailurePolicy)
	}
	return taskFailurePolicy, positional, programArgs, false, nil
}

func readInput(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func displayName(path string) string {
	if path == "-" {
		return "<stdin>"
	}
	return path
}

func validateExtension(path string) error {
	if path == "-" {
		return nil
	}
	if !strings.HasSuffix(path, ".k") {
		return fmt.Errorf("file must have .k extension: %s", path)
	}
	return nil
}

func parseProgram(data []byte, filename string) (*ast.Program, error) {
	p := parser.New(lexer.New(string(data)))
	program := p.ParseProgram()
	if errs := p.ErrorsDetailed(); len(errs) > 0 {
		return nil, fmt.Errorf("%s", parser.FormatParseErrors(errs, string(data), filename))
	}
	return program, nil
}

func runProgram(program *ast.Program, source string, filename string, taskFailurePolicy string, programArgs []string) (interpreter.Value, error) {
	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, filename)
	if err := eval.SetTaskFailurePolicy(taskFailurePolicy); err != nil {
		return nil, err
	}
	eval.SetProgramArgs(programArgs)
	eval.SetProgramPath(filename)
	eval.SetInput(os.Stdin)
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		return nil, err
	}
	if sig != nil {
		return nil, fmt.Errorf("break/continue outside loop")
	}
	if err := eval.CheckUnhandledTaskFailures(); err != nil {
		return nil, err
	}
	return val, nil
}

func loomCommand(args []string) int {
	if len(args) == 0 {
		return replCommand(nil)
	}

	switch args[0] {
	case "-h", "--help", "help":
		loomUsage()
		return 0
	case "serve":
		return replServerCommand(args[1:])
	case "connect":
		return replClientCommand(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown loom subcommand: %s\n", args[0])
		loomUsage()
		return 2
	}
}

func loomUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl loom\n")
	fmt.Fprintf(os.Stderr, "  karl loom serve [--addr=host:port]\n")
	fmt.Fprintf(os.Stderr, "  karl loom connect <host:port>\n")
	fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
	fmt.Fprintf(os.Stderr, "  serve      start a remote Loom REPL server\n")
	fmt.Fprintf(os.Stderr, "  connect    connect to a remote Loom REPL server\n")
}

func replCommand(args []string) int {
	help := false
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			help = true
			break
		}
		if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			replUsage()
			return 2
		}
	}
	if help {
		replUsage()
		return 0
	}
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "repl takes no arguments\n")
		replUsage()
		return 2
	}
	repl.StartWithVersion(os.Stdin, os.Stdout, cliVersion())
	return 0
}

func cliVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	rev := buildInfoSetting(info, "vcs.revision")
	if len(rev) > 7 {
		rev = rev[:7]
	}
	if rev == "" {
		return "dev"
	}
	if buildInfoSetting(info, "vcs.modified") == "true" {
		return "dev+" + rev + "-dirty"
	}
	return "dev+" + rev
}

func buildInfoSetting(info *debug.BuildInfo, key string) string {
	for _, s := range info.Settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}

func replUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl loom\n")
	fmt.Fprintf(os.Stderr, "  karl repl (alias)\n")
	fmt.Fprintf(os.Stderr, "\nStarts an interactive Read-Eval-Print Loop.\n")
	fmt.Fprintf(os.Stderr, "Type expressions and press Enter to evaluate them.\n")
	fmt.Fprintf(os.Stderr, "Type :help for REPL commands.\n")
}

func replServerCommand(args []string) int {
	addr := "localhost:9000"
	help := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			help = true
		case strings.HasPrefix(arg, "--addr="):
			addr = strings.TrimPrefix(arg, "--addr=")
		case arg == "--addr":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "--addr requires a value\n")
				replServerUsage()
				return 2
			}
			addr = args[i+1]
			i++
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			replServerUsage()
			return 2
		default:
			fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", arg)
			replServerUsage()
			return 2
		}
	}

	if help {
		replServerUsage()
		return 0
	}

	if err := repl.Server(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		return 1
	}

	return 0
}

func replServerUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl loom serve [--addr=host:port]\n")
	fmt.Fprintf(os.Stderr, "  karl repl-server [--addr=host:port] (alias)\n")
	fmt.Fprintf(os.Stderr, "\nStarts a Karl REPL server that clients can connect to.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --addr string   address to listen on (default \"localhost:9000\")\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  karl loom serve\n")
	fmt.Fprintf(os.Stderr, "  karl loom serve --addr=0.0.0.0:9000\n")
}

func replClientCommand(args []string) int {
	help := false
	var addr string

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			help = true
			break
		}
		if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			replClientUsage()
			return 2
		}
		if addr == "" {
			addr = arg
		} else {
			fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", arg)
			replClientUsage()
			return 2
		}
	}

	if help {
		replClientUsage()
		return 0
	}

	if addr == "" {
		fmt.Fprintf(os.Stderr, "missing required argument: host:port\n")
		replClientUsage()
		return 2
	}

	if err := repl.Client(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Client error: %v\n", err)
		return 1
	}

	return 0
}

func replClientUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl loom connect <host:port>\n")
	fmt.Fprintf(os.Stderr, "  karl repl-client <host:port> (alias)\n")
	fmt.Fprintf(os.Stderr, "\nConnects to a remote Karl REPL server.\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  karl loom connect localhost:9000\n")
	fmt.Fprintf(os.Stderr, "  karl loom connect 192.168.1.100:9000\n")
}

func notebookCommand(args []string) int {
	if len(args) < 1 {
		notebookUsage()
		return 2
	}

	if args[0] == "convert" {
		return convertCommand(args[1:])
	}

	help := false
	outputFile := ""
	var inputFile string
	step := false
	replMode := false
	quiet := false

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			help = true
			break
		}
		if strings.HasPrefix(arg, "--output=") {
			outputFile = strings.TrimPrefix(arg, "--output=")
			continue
		}
		if arg == "--step" || arg == "-s" {
			step = true
			continue
		}
		if arg == "--repl" || arg == "-r" {
			replMode = true
			continue
		}
		if arg == "--quiet" || arg == "-q" {
			quiet = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			notebookUsage()
			return 2
		}

		if inputFile == "" {
			inputFile = arg
		} else {
			fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", arg)
			notebookUsage()
			return 2
		}
	}

	if help {
		notebookUsage()
		return 0
	}

	if inputFile == "" {
		fmt.Fprintf(os.Stderr, "missing required argument: <file.knb>\n")
		notebookUsage()
		return 2
	}

	nb, err := notebook.LoadNotebook(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading notebook: %v\n", err)
		return 1
	}

	runner := notebook.NewRunner()

	if step || replMode {
		if err := runner.RunInteractive(nb, step, replMode, inputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Interactive execution error: %v\n", err)
			return 1
		}
		return 0
	}

	outputs, err := runner.ExecuteNotebook(nb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		// We still output what we have
	}

	// Print results
	if !quiet {
		fmt.Printf("Notebook: %s\n", nb.Title)
		fmt.Printf("Executed %d cells\n", len(outputs))
		fmt.Println()
	}

	for i, output := range outputs {
		if quiet {
			// In quiet mode, only print the last output if it has a value and no error
			if i == len(outputs)-1 {
				if output.Error != nil {
					fmt.Printf("Error: %s\n", output.Error.Message)
				} else if output.Value != "" {
					fmt.Println(output.Value)
				}
			}
		} else {
			if output.Error != nil {
				fmt.Printf("Cell %d [ERROR]: %s\n", output.CellIndex, output.Error.Message)
			} else if output.Value != "" {
				fmt.Printf("Cell %d: %s\n", output.CellIndex, output.Value)
			}
		}
	}

	// Save output if requested
	if outputFile != "" {
		data := map[string]interface{}{
			"notebook": nb,
			"outputs":  outputs,
		}
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal output: %v\n", err)
			return 1
		}
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write output file: %v\n", err)
			return 1
		}
		fmt.Printf("\nOutput saved to: %s\n", outputFile)
	}

	return 0
}

func notebookUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl notebook <file.knb> [--output=output.json] [--step] [--repl]\n")
	fmt.Fprintf(os.Stderr, "  karl notebook convert <in.ipynb> <out.knb>\n")
	fmt.Fprintf(os.Stderr, "  karl nb <file.knb> ... (alias)\n")
	fmt.Fprintf(os.Stderr, "\nExecutes a Karl notebook and displays results.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --output string   save execution results to JSON file\n")
	fmt.Fprintf(os.Stderr, "  --step, -s        run step-by-step with confirmation\n")
	fmt.Fprintf(os.Stderr, "  --repl, -r        run all cells then enter REPL mode\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  karl notebook nums.knb --step\n")
	fmt.Fprintf(os.Stderr, "  karl notebook nums.knb --repl\n")
}

func convertCommand(args []string) int {
	return notebook.ConvertCommand(args)
}

func kernelCommand(args []string) int {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: karl kernel <connection_file>\n")
		return 2
	}

	configFile := args[0]
	k, err := kernel.NewKernel(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize kernel: %v\n", err)
		return 1
	}

	if err := k.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Kernel error: %v\n", err)
		return 1
	}

	return 0
}

func spreadsheetCommand(args []string) int {
	addr := ":8080" // Default to :8080
	if len(args) > 0 {
		addr = args[0]
		// Binding to "localhost" can cause issues with IPv4/IPv6 mismatch.
		// Prefer binding to all interfaces (e.g. ":8082").
		addr = strings.Replace(addr, "localhost", "", 1)
		
		// If port only (e.g. "8081"), prepend ":"
		if !strings.Contains(addr, ":") {
			addr = ":" + addr
		}
	}

	srv := spreadsheet.NewServer()
	if err := srv.Start(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Spreadsheet server error: %v\n", err)
		return 1
	}
	return 0
}

func playgroundCommand(args []string) int {
	addr := ":8081" // Default to :8081
	if len(args) > 0 {
		addr = args[0]
		// Binding to "localhost" can cause issues with IPv4/IPv6 mismatch.
		// Prefer binding to all interfaces (e.g. ":8082").
		addr = strings.Replace(addr, "localhost", "", 1)
		
		// If port only (e.g. "8081"), prepend ":"
		if !strings.Contains(addr, ":") {
			addr = ":" + addr
		}
	}

	srv := playground.NewServer()
	if err := srv.Start(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Playground server error: %v\n", err)
		return 1
	}
	return 0
}
