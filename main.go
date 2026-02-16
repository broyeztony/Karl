package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"

	"karl/ast"
	"karl/interpreter"
	"karl/kernel"
	"karl/lexer"
	"karl/notebook"
	"karl/parser"
	"karl/repl"
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
	case "kernel":
		os.Exit(kernelCommand(os.Args[2:]))
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", sub)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Karl CLI version: %s\n\n", cliVersion())
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl parse <file.k> [--format=pretty|json]\n")
	fmt.Fprintf(os.Stderr, "  karl run <file.k> [--task-failure-policy=fail-fast|defer]\n")
	fmt.Fprintf(os.Stderr, "  karl loom\n")
	fmt.Fprintf(os.Stderr, "  karl loom serve [--addr=host:port]\n")
	fmt.Fprintf(os.Stderr, "  karl loom connect <host:port>\n")
	fmt.Fprintf(os.Stderr, "  karl notebook <file.knb> [--output=file.json]\n")
	fmt.Fprintf(os.Stderr, "  karl kernel <connection_file.json>\n")
	fmt.Fprintf(os.Stderr, "\nCompatibility aliases:\n")
	fmt.Fprintf(os.Stderr, "  karl repl\n")
	fmt.Fprintf(os.Stderr, "  karl repl-server [--addr=host:port]\n")
	fmt.Fprintf(os.Stderr, "  karl repl-client <host:port>\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
	fmt.Fprintf(os.Stderr, "  Use \"karl <command> --help\" for command help\n")
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
	taskFailurePolicy, positional, help, err := parseRunArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		runUsage()
		return 2
	}
	if help {
		runUsage()
		return 0
	}
	if len(positional) != 1 {
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
	val, err := runProgram(program, string(data), filename, taskFailurePolicy)
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

func parseUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl parse <file.k> [--format=pretty|json]\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --format string   output format: pretty|json (default \"pretty\")\n")
}

func runUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl run <file.k> [--task-failure-policy=fail-fast|defer]\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --task-failure-policy string   task failure behavior: fail-fast|defer (default \"fail-fast\")\n")
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

func parseRunArgs(args []string) (string, []string, bool, error) {
	taskFailurePolicy := interpreter.TaskFailurePolicyFailFast
	positional := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return taskFailurePolicy, positional, true, nil
		case strings.HasPrefix(arg, "--task-failure-policy="):
			taskFailurePolicy = strings.TrimPrefix(arg, "--task-failure-policy=")
		case arg == "--task-failure-policy":
			if i+1 >= len(args) {
				return taskFailurePolicy, positional, false, fmt.Errorf("--task-failure-policy requires a value")
			}
			taskFailurePolicy = args[i+1]
			i++
		case arg == "-":
			positional = append(positional, arg)
		case strings.HasPrefix(arg, "-"):
			return taskFailurePolicy, positional, false, fmt.Errorf("unknown flag: %s", arg)
		default:
			positional = append(positional, arg)
		}
	}
	if taskFailurePolicy != interpreter.TaskFailurePolicyFailFast && taskFailurePolicy != interpreter.TaskFailurePolicyDefer {
		return taskFailurePolicy, positional, false, fmt.Errorf("invalid --task-failure-policy: %s", taskFailurePolicy)
	}
	return taskFailurePolicy, positional, false, nil
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

func runProgram(program *ast.Program, source string, filename string, taskFailurePolicy string) (interpreter.Value, error) {
	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, filename)
	if err := eval.SetTaskFailurePolicy(taskFailurePolicy); err != nil {
		return nil, err
	}
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
	help := false
	outputFile := ""
	var inputFile string

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			help = true
			break
		}
		if strings.HasPrefix(arg, "--output=") {
			outputFile = strings.TrimPrefix(arg, "--output=")
		} else if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			notebookUsage()
			return 2
		} else if inputFile == "" {
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

	// Load the notebook
	nb, err := notebook.LoadNotebook(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load notebook: %v\n", err)
		return 1
	}

	// Create a runner and execute the notebook
	runner := notebook.NewRunner()
	outputs, err := runner.ExecuteNotebook(nb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "notebook execution failed: %v\n", err)
		return 1
	}

	// Print results
	fmt.Printf("Notebook: %s\n", nb.Title)
	fmt.Printf("Executed %d cells\n", len(outputs))
	fmt.Println()

	for _, output := range outputs {
		if output.Error != nil {
			fmt.Printf("Cell %d [ERROR]: %s\n", output.CellIndex, output.Error.Message)
		} else if output.Value != "" {
			fmt.Printf("Cell %d: %s\n", output.CellIndex, output.Value)
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
		if err := ioutil.WriteFile(outputFile, jsonData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write output file: %v\n", err)
			return 1
		}
		fmt.Printf("\nOutput saved to: %s\n", outputFile)
	}

	return 0
}

func notebookUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl notebook <file.knb> [--output=output.json]\n")
	fmt.Fprintf(os.Stderr, "  karl nb <file.knb> [--output=output.json] (alias)\n")
	fmt.Fprintf(os.Stderr, "\nExecutes a Karl notebook and displays results.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --output string   save execution results to JSON file\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  karl notebook example.knb\n")
	fmt.Fprintf(os.Stderr, "  karl notebook example.knb --output=results.json\n")
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

