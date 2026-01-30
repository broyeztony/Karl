package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"karl/ast"
	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
	"karl/shape"
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
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", sub)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl parse <file.k|file.shape> [--format=pretty|json]\n")
	fmt.Fprintf(os.Stderr, "  karl run <file.k>\n")
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
	if err := validateParseExtension(positional[0]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 2
	}
	data, err := readInput(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		return 1
	}
	filename := displayName(positional[0])
	if isShapeFile(filename) {
		sh, err := shape.ParseFile(string(data))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
		switch format {
		case "pretty":
			fmt.Print(shape.FormatFile(sh))
		case "json":
			out, err := shape.FormatJSONFile(sh)
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
	program, err := parseProgram(data, filename)
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
	positional, help, err := parseRunArgs(args)
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
	if err := validateRunExtension(positional[0]); err != nil {
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
	val, err := runProgram(program, string(data), filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, interpreter.FormatRuntimeError(err, string(data), filename))
		return 1
	}
	if val != nil && val.Type() != interpreter.UNIT {
		fmt.Println(val.Inspect())
	}
	return 0
}

func parseUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl parse <file.k|file.shape> [--format=pretty|json]\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --format string   output format: pretty|json (default \"pretty\")\n")
}

func runUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl run <file.k>\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
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

func parseRunArgs(args []string) ([]string, bool, error) {
	positional := []string{}
	for _, arg := range args {
		switch {
		case arg == "-h" || arg == "--help":
			return positional, true, nil
		case arg == "-":
			positional = append(positional, arg)
		case strings.HasPrefix(arg, "-"):
			return positional, false, fmt.Errorf("unknown flag: %s", arg)
		default:
			positional = append(positional, arg)
		}
	}
	return positional, false, nil
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

func isShapeFile(path string) bool {
	return strings.HasSuffix(path, ".shape")
}

func validateParseExtension(path string) error {
	if path == "-" {
		return nil
	}
	if strings.HasSuffix(path, ".k") || strings.HasSuffix(path, ".shape") {
		return nil
	}
	return fmt.Errorf("file must have .k or .shape extension: %s", path)
}

func validateRunExtension(path string) error {
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

func runProgram(program *ast.Program, source string, filename string) (interpreter.Value, error) {
	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, filename)
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		return nil, err
	}
	if sig != nil {
		return nil, fmt.Errorf("break/continue outside loop")
	}
	return val, nil
}
