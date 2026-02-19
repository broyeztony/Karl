//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"karl/ast"
	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
	"syscall/js"
)

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("runKarl", js.FuncOf(runKarl))
	fmt.Println("Karl WASM Runtime initialized.")
	<-c
}

func runKarl(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return "Error: runKarl expects 1 argument (source code)"
	}
	source := args[0].String()
	
	// Parse
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if errs := p.ErrorsDetailed(); len(errs) > 0 {
		msg := parser.FormatParseErrors(errs, source, "playground.k")
		if msg == "" {
			msg = "parse error"
		}
		fmt.Printf("%s\n", msg)
		return nil
	}

	// Run
	// Using a new environment for each run? Or persistent?
	// Usually playgrounds are stateless (per run).
	// But REPLs are stateful.
	// Let's make it stateless for now to avoid weird state bugs.
	// Users can restart if they want clear state.
	
	err := runProgram(program, source, "playground.k")
	if err != nil {
		fmt.Printf("Runtime Error: %v\n", err)
	}

	return nil
}

func runProgram(program *ast.Program, source string, filename string) error {
	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, filename)
	// Set Failure Policy? Default is FailFast.
	
	env := interpreter.NewBaseEnvironment()
	
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		if ute, ok := err.(*interpreter.UnhandledTaskError); ok {
			return ute
		}
		// Format error nicer? 
		// The error from Eval is usually a *RuntimeError which has Message.
		// interpreter.FormatRuntimeError is useful but not exported?
		// Wait, I saw it in main.go: interpreter.FormatRuntimeError(err, string(data), filename)
		// Check if it's exported.
		// Yes: checked `main.go`.
		return fmt.Errorf("\n%s", interpreter.FormatRuntimeError(err, source, filename))
	}
	
	if sig != nil {
		return fmt.Errorf("break/continue outside loop")
	}
	
	if err := eval.CheckUnhandledTaskFailures(); err != nil {
		return err
	}
	
	// Print result if not unit
	if _, ok := val.(*interpreter.Unit); !ok {
		fmt.Println(val.Inspect())
	}
	
	return nil
}
