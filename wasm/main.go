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

func notifyDone() {
	doneFn := js.Global().Get("__karl_done")
	if doneFn.Type() == js.TypeFunction {
		doneFn.Invoke()
	}
}

func runKarl(this js.Value, args []js.Value) interface{} {
	// Important for wasm/js: do not block inside a JS callback.
	// Running evaluation in a goroutine lets the JS event loop progress,
	// which is required for async APIs such as net/http on wasm.
	go func() {
		defer notifyDone()
		if len(args) != 1 {
			fmt.Println("Error: runKarl expects 1 argument (source code)")
			return
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
			return
		}

		err := runProgram(program, source, "playground.k")
		if err != nil {
			fmt.Printf("Runtime Error: %v\n", err)
		}
	}()

	return nil
}

func runProgram(program *ast.Program, source string, filename string) error {
	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, filename)

	env := interpreter.NewBaseEnvironment()

	val, sig, err := eval.Eval(program, env)
	if err != nil {
		if ute, ok := err.(*interpreter.UnhandledTaskError); ok {
			return ute
		}
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
