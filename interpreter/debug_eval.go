package interpreter

import (
	"fmt"

	"karl/ast"
	"karl/lexer"
	"karl/parser"
)

// EvalDebugExpression parses and evaluates a single expression against the
// provided environment. It is intended for debugger "print/evaluate" features.
func EvalDebugExpression(input string, env *Environment) (Value, error) {
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

	eval := NewEvaluatorWithSourceAndFilename(input, "<debug>")
	val, sig, err := eval.Eval(stmt.Expression, env)
	if err != nil {
		return nil, err
	}
	if sig != nil {
		return nil, fmt.Errorf("print expression cannot produce control flow")
	}
	return val, nil
}
