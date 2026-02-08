package tests

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"karl/ast"
	"karl/lexer"
	"karl/parser"
)

func parseProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	return program
}

func parseExpression(t *testing.T, input string) ast.Expression {
	t.Helper()
	program := parseProgram(t, input)
	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}
	switch stmt := program.Statements[0].(type) {
	case *ast.ExpressionStatement:
		return stmt.Expression
	case *ast.LetStatement:
		return stmt.Value
	default:
		t.Fatalf("unexpected statement type: %T", program.Statements[0])
	}
	return nil
}

func parseFile(t *testing.T, path string) *ast.Program {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return parseProgram(t, string(data))
}

func listKarlFiles(t *testing.T, root string) []string {
	t.Helper()
	files := []string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".k") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(files)
	return files
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}
	t.Fatalf("parser had %d errors", len(errors))
}

func countNodes(node ast.Node, match func(ast.Node) bool) int {
	count := 0
	walk(node, func(n ast.Node) {
		if match(n) {
			count++
		}
	})
	return count
}

func requireCountAtLeast(t *testing.T, node ast.Node, label string, min int, match func(ast.Node) bool) {
	t.Helper()
	count := countNodes(node, match)
	if count < min {
		t.Fatalf("expected at least %d %s nodes, got %d", min, label, count)
	}
}

func walk(node ast.Node, visit func(ast.Node)) {
	if node == nil {
		return
	}
	visit(node)

	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			walk(stmt, visit)
		}
	case *ast.LetStatement:
		walk(n.Name, visit)
		walk(n.Value, visit)
	case *ast.ExpressionStatement:
		walk(n.Expression, visit)
	case *ast.Identifier, *ast.Placeholder, *ast.IntegerLiteral, *ast.FloatLiteral,
		*ast.StringLiteral, *ast.CharLiteral, *ast.BooleanLiteral, *ast.NullLiteral,
		*ast.UnitLiteral, *ast.ContinueExpression, *ast.WildcardPattern:
		return
	case *ast.PrefixExpression:
		walk(n.Right, visit)
	case *ast.InfixExpression:
		walk(n.Left, visit)
		walk(n.Right, visit)
	case *ast.AssignExpression:
		walk(n.Left, visit)
		walk(n.Right, visit)
	case *ast.PostfixExpression:
		walk(n.Left, visit)
	case *ast.AwaitExpression:
		walk(n.Value, visit)
	case *ast.ImportExpression:
		walk(n.Path, visit)
	case *ast.IfExpression:
		walk(n.Condition, visit)
		walk(n.Consequence, visit)
		walk(n.Alternative, visit)
	case *ast.BlockExpression:
		for _, stmt := range n.Statements {
			walk(stmt, visit)
		}
	case *ast.MatchExpression:
		walk(n.Value, visit)
		for _, arm := range n.Arms {
			walk(arm.Pattern, visit)
			walk(arm.Guard, visit)
			walk(arm.Body, visit)
		}
	case *ast.ForExpression:
		walk(n.Condition, visit)
		for _, b := range n.Bindings {
			walk(b.Pattern, visit)
			walk(b.Value, visit)
		}
		walk(n.Body, visit)
		walk(n.Then, visit)
	case *ast.LambdaExpression:
		for _, param := range n.Params {
			walk(param, visit)
		}
		walk(n.Body, visit)
	case *ast.CallExpression:
		walk(n.Function, visit)
		for _, arg := range n.Arguments {
			walk(arg, visit)
		}
	case *ast.MemberExpression:
		walk(n.Object, visit)
		walk(n.Property, visit)
	case *ast.IndexExpression:
		walk(n.Left, visit)
		walk(n.Index, visit)
	case *ast.SliceExpression:
		walk(n.Left, visit)
		walk(n.Start, visit)
		walk(n.End, visit)
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			walk(el, visit)
		}
	case *ast.ObjectLiteral:
		for _, entry := range n.Entries {
			walk(entry.Value, visit)
		}
	case *ast.StructInitExpression:
		walk(n.TypeName, visit)
		walk(n.Value, visit)
	case *ast.RangeExpression:
		walk(n.Start, visit)
		walk(n.End, visit)
		walk(n.Step, visit)
	case *ast.QueryExpression:
		walk(n.Var, visit)
		walk(n.Source, visit)
		for _, w := range n.Where {
			walk(w, visit)
		}
		walk(n.OrderBy, visit)
		walk(n.Select, visit)
	case *ast.RaceExpression:
		for _, task := range n.Tasks {
			walk(task, visit)
		}
	case *ast.SpawnExpression:
		walk(n.Task, visit)
		for _, task := range n.Group {
			walk(task, visit)
		}
	case *ast.BreakExpression:
		walk(n.Value, visit)
	case *ast.RangePattern:
		walk(n.Start, visit)
		walk(n.End, visit)
	case *ast.ObjectPattern:
		for _, entry := range n.Entries {
			walk(entry.Pattern, visit)
		}
	case *ast.ArrayPattern:
		for _, el := range n.Elements {
			walk(el, visit)
		}
		walk(n.Rest, visit)
	case *ast.TuplePattern:
		for _, el := range n.Elements {
			walk(el, visit)
		}
	case *ast.CallPattern:
		walk(n.Name, visit)
		for _, arg := range n.Args {
			walk(arg, visit)
		}
	}
}
