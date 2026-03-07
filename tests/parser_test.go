package tests

import (
	"karl/ast"
	"karl/lexer"
	"karl/parser"
	"strings"
	"testing"
)

func TestLetStatement(t *testing.T) {
	input := `let x = 5`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("statement not LetStatement. got=%T", program.Statements[0])
	}
	if stmt.Value == nil {
		t.Fatalf("let value is nil")
	}
}

func TestMatchExpression(t *testing.T) {
	input := `match x { case 1..4 -> "small" case _ -> "other" }`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	match, ok := stmt.Expression.(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expression not MatchExpression. got=%T", stmt.Expression)
	}
	if len(match.Arms) != 2 {
		t.Fatalf("expected 2 match arms, got %d", len(match.Arms))
	}
}

func TestForExpression(t *testing.T) {
	input := `for i < 10 with i = 0 { i++ } then i`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	forExpr, ok := stmt.Expression.(*ast.ForExpression)
	if !ok {
		t.Fatalf("expression not ForExpression. got=%T", stmt.Expression)
	}
	if len(forExpr.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(forExpr.Bindings))
	}
	if forExpr.Then == nil {
		t.Fatalf("expected then expression")
	}
}

func TestSpawnAndRace(t *testing.T) {
	input := `let tasks = & { taskA(), taskB() }
let fastest = !& { taskA(), taskB() }`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt0 := program.Statements[0].(*ast.LetStatement)
	if _, ok := stmt0.Value.(*ast.SpawnExpression); !ok {
		t.Fatalf("expected SpawnExpression, got %T", stmt0.Value)
	}
	stmt1 := program.Statements[1].(*ast.LetStatement)
	if _, ok := stmt1.Value.(*ast.RaceExpression); !ok {
		t.Fatalf("expected RaceExpression, got %T", stmt1.Value)
	}
}

func TestSpawnAndRaceKeywordAliases(t *testing.T) {
	input := `let tasks = spawn { taskA(), taskB() }
let fastest = race { taskA(), taskB() }`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt0 := program.Statements[0].(*ast.LetStatement)
	if _, ok := stmt0.Value.(*ast.SpawnExpression); !ok {
		t.Fatalf("expected SpawnExpression, got %T", stmt0.Value)
	}
	stmt1 := program.Statements[1].(*ast.LetStatement)
	if _, ok := stmt1.Value.(*ast.RaceExpression); !ok {
		t.Fatalf("expected RaceExpression, got %T", stmt1.Value)
	}
}

func TestPipelineOperatorParsing(t *testing.T) {
	input := `let p = cmd({ command: "echo", args: ["hi"], }) | cmd({ command: "wc", args: ["-c"], })`
	p := parser.New(lexer.New(input))
	_ = p.ParseProgram()
	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatalf("expected parse error")
	}
	found := false
	for _, err := range errors {
		if strings.Contains(err, "operator '|' is reserved for stream pipelines") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected pipe reservation parse error, got %v", errors)
	}
}
