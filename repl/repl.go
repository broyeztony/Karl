package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

const (
	PROMPT      = "karl> "
	PROMPT_CONT = "...   "
)

// Start begins the REPL session
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := interpreter.NewBaseEnvironment()
	eval := interpreter.NewEvaluatorWithSourceAndFilename("", "<repl>")

	fmt.Fprintf(out, "Karl REPL - Type expressions and press Enter\n")
	fmt.Fprintf(out, "Commands: :help, :quit, :env\n\n")

	var inputBuffer strings.Builder
	multiline := false

	for {
		// Show appropriate prompt
		if multiline {
			fmt.Fprint(out, PROMPT_CONT)
		} else {
			fmt.Fprint(out, PROMPT)
		}

		// Read input
		if !scanner.Scan() {
			return
		}

		line := scanner.Text()

		// Handle REPL commands
		if !multiline && strings.HasPrefix(line, ":") {
			if handleCommand(line, out, env) {
				return // :quit was called
			}
			continue
		}

		// Accumulate input
		if inputBuffer.Len() > 0 {
			inputBuffer.WriteString("\n")
		}
		inputBuffer.WriteString(line)

		input := inputBuffer.String()

		// Try to parse the accumulated input
		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		// Check for parse errors
		if errs := p.ErrorsDetailed(); len(errs) > 0 {
			// Check if this looks like incomplete input
			if isIncompleteInput(input, errs) {
				multiline = true
				continue
			}

			// Real parse error - show it and reset
			fmt.Fprintf(out, "Parse error:\n%s\n", parser.FormatParseErrors(errs, input, "<repl>"))
			inputBuffer.Reset()
			multiline = false
			continue
		}

		// Successfully parsed - evaluate it
		multiline = false
		inputBuffer.Reset()

		// Update evaluator source for better error messages
		eval = interpreter.NewEvaluatorWithSourceAndFilename(input, "<repl>")

		val, sig, err := eval.Eval(program, env)
		if err != nil {
			fmt.Fprintf(out, "Error: %s\n", interpreter.FormatRuntimeError(err, input, "<repl>"))
			continue
		}

		if sig != nil {
			fmt.Fprintf(out, "Error: break/continue outside loop\n")
			continue
		}

		// Check for unhandled task failures
		if err := eval.CheckUnhandledTaskFailures(); err != nil {
			fmt.Fprintf(out, "Error: %s\n", err)
			continue
		}

		// Print result (unless it's Unit)
		if _, ok := val.(*interpreter.Unit); !ok {
			fmt.Fprintf(out, "%s\n", val.Inspect())
		}
	}
}

// handleCommand processes REPL commands (starting with :)
// Returns true if the REPL should exit
func handleCommand(cmd string, out io.Writer, env *interpreter.Environment) bool {
	switch strings.TrimSpace(cmd) {
	case ":quit", ":q", ":exit":
		fmt.Fprintln(out, "Goodbye!")
		return true

	case ":help", ":h":
		fmt.Fprintln(out, "REPL Commands:")
		fmt.Fprintln(out, "  :help, :h     - Show this help")
		fmt.Fprintln(out, "  :quit, :q     - Exit the REPL")
		fmt.Fprintln(out, "  :env          - Show current environment bindings")
		fmt.Fprintln(out, "  :clear        - Clear the screen")
		fmt.Fprintln(out, "\nTips:")
		fmt.Fprintln(out, "  - Press Enter on an incomplete line to continue on the next line")
		fmt.Fprintln(out, "  - Variables persist across evaluations")
		fmt.Fprintln(out, "  - The last expression's value is printed")

	case ":env":
		fmt.Fprintln(out, "Current environment bindings:")
		printEnv(out, env)

	case ":clear":
		fmt.Fprint(out, "\033[H\033[2J")

	default:
		fmt.Fprintf(out, "Unknown command: %s (try :help)\n", cmd)
	}

	return false
}

// printEnv displays the current environment bindings
func printEnv(out io.Writer, env *interpreter.Environment) {
	// This is a simple implementation - we'd need to expose environment internals
	// to make this more useful. For now, just show a placeholder.
	fmt.Fprintln(out, "  (environment inspection not yet implemented)")
}

// isIncompleteInput checks if parse errors suggest incomplete input
func isIncompleteInput(input string, errs []parser.ParseError) bool {
	// Heuristic: if the input ends with an opening brace, bracket, or paren,
	// or if we have unclosed delimiters, treat it as incomplete
	trimmed := strings.TrimSpace(input)
	
	if strings.HasSuffix(trimmed, "{") ||
		strings.HasSuffix(trimmed, "[") ||
		strings.HasSuffix(trimmed, "(") ||
		strings.HasSuffix(trimmed, "->") {
		return true
	}

	// Check if errors mention unexpected EOF or missing closing delimiter
	for _, err := range errs {
		msg := strings.ToLower(err.Message)
		if strings.Contains(msg, "expected }") ||
			strings.Contains(msg, "expected ]") ||
			strings.Contains(msg, "expected )") ||
			strings.Contains(msg, "unexpected eof") {
			return true
		}
	}

	return false
}
