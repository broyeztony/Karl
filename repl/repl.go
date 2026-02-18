package repl

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
	"karl/token"
)

const (
	PROMPT      = "❯ "
	PROMPT_CONT = "...   "
)

const replIntroArt = `                                                                                                    

                                                                                                    
                                                 .==-:.                                             
                                                .==..=:                                             
                                                :+:  -=.                                            
                                               .==.  -*-                                            
                                               .==  .+%*:                                           
                                               .==.  :+:                                            
                                               .=+.  -+:                                            
                                                -+: .*+.                                            
                                                :+- :*-                                             
                                                :+=.=*:                                             
                                                 -*=*=.                                             
                                                 :*%*:                                              
                                                  =%=                                               
                                                 =#%*:                                              
                                               .-#==**:                                             
                                              .=*=.:+%*:                                            
                                            .:**:   .+%*:                                           
                                           :##=.     :*%*:                                          
                                        .:*#+:.        -#=.                                         
                                     .:*%#=             :*#:                                        
          .:...  . .-+:-:::::::==+*#%%*-                  :%*:                                      
                   .*%#***#***++=-:. .                      :%%%*:.         .::-::.::-==:.          
                                                              .:*%%%%%%%%%%%%#**==:....             
                                                                  .........                         
                                                                                                    
                                                                                                    `

type scannerResult struct {
	line string
	err  error
	ok   bool
}

type startOptions struct {
	showIntro  bool
	cliVersion string
}

// Start begins the REPL session
func Start(in io.Reader, out io.Writer) {
	start(in, out, startOptions{showIntro: true})
}

// StartWithVersion begins the REPL session and displays the provided CLI version.
func StartWithVersion(in io.Reader, out io.Writer, cliVersion string) {
	start(in, out, startOptions{showIntro: true, cliVersion: cliVersion})
}

func start(in io.Reader, out io.Writer, opts startOptions) {
	env := interpreter.NewBaseEnvironment()
	eval := interpreter.NewEvaluatorWithSourceAndFilename("", "<repl>")
	eval.SetInputUnavailableMessage("readLine is not available in REPL; use `karl run` for stdin input")

	var (
		scanCh chan scannerResult
		tty    *ttyInput
	)
	if ti, ok := newTTYInput(in, out); ok {
		tty = ti
		defer tty.Close()
	} else if _, ok := in.(net.Conn); ok {
		tty = newStreamInput(in, out)
		defer tty.Close()
	} else {
		scanner := bufio.NewScanner(in)
		scanCh = make(chan scannerResult)
		go scanInput(scanner, scanCh)
	}

	sessionOut := out
	if tty != nil {
		// In raw TTY mode, normalize LF to CRLF so lines start in column 0.
		sessionOut = newTTYLineWriter(out)
	}

	if opts.showIntro {
		fmt.Fprintf(sessionOut, "%s\n", replIntroArt)
		if v := strings.TrimSpace(opts.cliVersion); v != "" {
			fmt.Fprintf(sessionOut, "Karl CLI version: %s\n", v)
		}
		fmt.Fprintf(sessionOut, "\n")
		fmt.Fprintf(sessionOut, "Type expressions and press Enter to evaluate.\n")
		fmt.Fprintf(sessionOut, "Commands: :help, :quit, :clear, :env\n")
		fmt.Fprintf(sessionOut, "See repl/EXAMPLES.md for ideas!\n\n")
	}

	var inputBuffer strings.Builder
	multiline := false

	for {
		// Show appropriate prompt
		prompt := PROMPT
		if multiline {
			prompt = PROMPT_CONT
		}

		var (
			line string
			ok   bool
		)
		if tty != nil {
			line, ok = tty.readLine(prompt, eval)
			if !ok {
				return
			}
		} else {
			fmt.Fprint(out, prompt)
			line, ok = waitForInput(scanCh, eval, out)
			if !ok {
				return
			}
		}

		// Support Ctrl+L (form feed) as a clear-screen shortcut.
		if isCtrlL(line) {
			fmt.Fprint(out, "\033[H\033[2J")
			continue
		}

		// Handle REPL commands
		if !multiline && strings.HasPrefix(line, ":") {
			if handleCommand(line, sessionOut, env) {
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

		// Check whether we should continue collecting multiline input.
		errs := p.ErrorsDetailed()
		if IsIncompleteInput(input, errs) {
			multiline = true
			continue
		}
		if len(errs) > 0 {
			// Real parse error - show it and reset
			fmt.Fprintf(sessionOut, "Parse error:\n%s\n", parser.FormatParseErrors(errs, input, "<repl>"))
			inputBuffer.Reset()
			multiline = false
			continue
		}

		// Successfully parsed - evaluate it
		multiline = false
		inputBuffer.Reset()

		// Keep one evaluator/runtime for the whole REPL session; only refresh
		// source metadata for diagnostics on each submitted input.
		eval.SetSourceAndFilename(input, "<repl>")

		val, sig, err := eval.Eval(program, env)
		if err != nil {
			fmt.Fprintf(sessionOut, "Error: %s\n", interpreter.FormatRuntimeError(err, input, "<repl>"))
			if isFatalREPLError(err) {
				return
			}
			continue
		}

		if sig != nil {
			fmt.Fprintf(sessionOut, "Error: break/continue outside loop\n")
			continue
		}

		// Check for unhandled task failures
		if err := eval.CheckUnhandledTaskFailures(); err != nil {
			fmt.Fprintf(sessionOut, "Error: %s\n", err)
			if isFatalREPLError(err) {
				return
			}
			continue
		}

		// Print result (unless it's Unit)
		if _, ok := val.(*interpreter.Unit); !ok {
			fmt.Fprintf(sessionOut, "%s\n", val.Inspect())
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
		fmt.Fprintln(out, "  :examples     - Show example code snippets")
		fmt.Fprintln(out, "  :clear        - Clear the screen (same as Ctrl+L)")
		fmt.Fprintln(out, "\nTips:")
		fmt.Fprintln(out, "  - Press Enter on an incomplete line to continue on the next line")
		fmt.Fprintln(out, "  - Variables persist across evaluations")
		fmt.Fprintln(out, "  - The last expression's value is printed")

	case ":env":
		fmt.Fprintln(out, "Current environment bindings:")
		printEnv(out, env)

	case ":examples", ":ex":
		showExamples(out)

	case ":clear":
		clearScreen(out)

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

// IsIncompleteInput checks if parse errors suggest incomplete input
func IsIncompleteInput(input string, errs []parser.ParseError) bool {
	// If there are unclosed delimiters, keep collecting lines.
	if hasUnclosedDelimiters(input) {
		return true
	}

	// Heuristic: if the input ends with an opening delimiter or arrow, treat it as incomplete.
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

func hasUnclosedDelimiters(input string) bool {
	l := lexer.New(input)
	parenDepth := 0
	braceDepth := 0
	bracketDepth := 0

	for {
		tok := l.NextToken()
		switch tok.Type {
		case token.LPAREN:
			parenDepth++
		case token.RPAREN:
			parenDepth--
		case token.LBRACE:
			braceDepth++
		case token.RBRACE:
			braceDepth--
		case token.LBRACKET:
			bracketDepth++
		case token.RBRACKET:
			bracketDepth--
		case token.EOF:
			return parenDepth > 0 || braceDepth > 0 || bracketDepth > 0
		}
	}
}

func isFatalREPLError(err error) bool {
	if err == nil {
		return false
	}
	_, fatal := err.(*interpreter.UnhandledTaskError)
	return fatal
}

func isCtrlL(line string) bool {
	return strings.TrimRight(line, "\r") == "\f"
}

func scanInput(scanner *bufio.Scanner, out chan<- scannerResult) {
	defer close(out)
	for scanner.Scan() {
		out <- scannerResult{line: scanner.Text(), ok: true}
	}
	if err := scanner.Err(); err != nil {
		out <- scannerResult{err: err}
	}
}

func waitForInput(scanCh <-chan scannerResult, eval *interpreter.Evaluator, out io.Writer) (string, bool) {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case in, ok := <-scanCh:
			if !ok {
				return "", false
			}
			if in.err != nil {
				fmt.Fprintf(out, "Input error: %v\n", in.err)
				return "", false
			}
			return in.line, in.ok
		case <-ticker.C:
			err := eval.CheckUnhandledTaskFailures()
			if err == nil {
				continue
			}
			fmt.Fprintf(out, "Error: %s\n", err)
			if isFatalREPLError(err) {
				return "", false
			}
		}
	}
}

// showExamples displays the EXAMPLES.md file content
func showExamples(out io.Writer) {
	// Try to find EXAMPLES.md in the repl directory
	examplesPath := findExamplesFile()

	if examplesPath == "" {
		fmt.Fprintln(out, "Examples file not found.")
		fmt.Fprintln(out, "See repl/EXAMPLES.md in the Karl repository for examples.")
		return
	}

	content, err := os.ReadFile(examplesPath)
	if err != nil {
		fmt.Fprintf(out, "Error reading examples file: %v\n", err)
		return
	}

	fmt.Fprintln(out, string(content))
}

// findExamplesFile attempts to locate EXAMPLES.md
func findExamplesFile() string {
	// Try common locations relative to where karl might be run from
	candidates := []string{
		"repl/EXAMPLES.md",
		"./repl/EXAMPLES.md",
		"../repl/EXAMPLES.md",
	}

	// Also try relative to the executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(exeDir, "repl", "EXAMPLES.md"))
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
