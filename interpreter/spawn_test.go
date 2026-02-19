package interpreter_test

import (
	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
	"testing"
)

func TestSpawn(t *testing.T) {
	input := `
	let c = channel()
	spawn((ch) -> {
		ch.send("Hello from thread!")
	})(c)
	let msg = c.recv()
	msg[0]
	`
	
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	eval := interpreter.NewEvaluator()
	env := interpreter.NewBaseEnvironment()
	
	val, _, err := eval.Eval(program, env)
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}

	if val.Inspect() != "\"Hello from thread!\"" {
		t.Errorf("Expected \"Hello from thread!\", got %s", val.Inspect())
	}
}
