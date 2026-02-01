package interpreter

import (
	"fmt"
	"strings"

	"karl/ast"
	"karl/token"
)

type RuntimeError struct {
	Message string
	Token   *token.Token
}

func (e *RuntimeError) Error() string {
	return e.Message
}

type RecoverableError struct {
	Message string
	Kind    string
	Token   *token.Token
}

func (e *RecoverableError) Error() string {
	return e.Message
}

func FormatRuntimeError(err error, source string, filename string) string {
	switch e := err.(type) {
	case *RuntimeError:
		return formatRuntimeError(e.Message, e.Token, source, filename)
	case *RecoverableError:
		return formatRuntimeError(e.Message, e.Token, source, filename)
	default:
		return err.Error()
	}
}

func formatRuntimeError(message string, tok *token.Token, source string, filename string) string {
	if tok == nil || tok.Line == 0 || source == "" {
		return "runtime error: " + message
	}
	lines := strings.Split(source, "\n")
	line := tok.Line
	col := tok.Column
	if line < 1 || line > len(lines) {
		return "runtime error: " + message
	}
	lineText := strings.TrimRight(lines[line-1], "\r")
	if col < 1 {
		col = 1
	}
	if col > len(lineText)+1 {
		col = len(lineText) + 1
	}
	caret := strings.Repeat(" ", col-1) + "^"
	location := fmt.Sprintf("%d:%d", line, tok.Column)
	if filename != "" {
		location = fmt.Sprintf("%s:%s", filename, location)
	}
	return fmt.Sprintf(
		"runtime error: %s\n  at %s\n  %d | %s\n    | %s",
		message,
		location,
		line,
		lineText,
		caret,
	)
}

type ExitError struct {
	Message string
}

func (e *ExitError) Error() string {
	if e.Message == "" {
		return "exit"
	}
	return fmt.Sprintf("exit: %s", e.Message)
}

func tokenFromNode(node ast.Node) *token.Token {
	switch n := node.(type) {
	case *ast.LetStatement:
		return &n.Token
	case *ast.ExpressionStatement:
		return &n.Token
	case *ast.Identifier:
		return &n.Token
	case *ast.Placeholder:
		return &n.Token
	case *ast.IntegerLiteral:
		return &n.Token
	case *ast.FloatLiteral:
		return &n.Token
	case *ast.StringLiteral:
		return &n.Token
	case *ast.CharLiteral:
		return &n.Token
	case *ast.BooleanLiteral:
		return &n.Token
	case *ast.NullLiteral:
		return &n.Token
	case *ast.UnitLiteral:
		return &n.Token
	case *ast.PrefixExpression:
		return &n.Token
	case *ast.InfixExpression:
		return &n.Token
	case *ast.AssignExpression:
		return &n.Token
	case *ast.PostfixExpression:
		return &n.Token
	case *ast.AwaitExpression:
		return &n.Token
	case *ast.ImportExpression:
		return &n.Token
	case *ast.RecoverExpression:
		return &n.Token
	case *ast.IfExpression:
		return &n.Token
	case *ast.BlockExpression:
		return &n.Token
	case *ast.MatchExpression:
		return &n.Token
	case *ast.ForExpression:
		return &n.Token
	case *ast.LambdaExpression:
		return &n.Token
	case *ast.CallExpression:
		return &n.Token
	case *ast.MemberExpression:
		return &n.Token
	case *ast.IndexExpression:
		return &n.Token
	case *ast.SliceExpression:
		return &n.Token
	case *ast.ArrayLiteral:
		return &n.Token
	case *ast.ObjectLiteral:
		return &n.Token
	case *ast.StructInitExpression:
		return &n.Token
	case *ast.RangeExpression:
		return &n.Token
	case *ast.QueryExpression:
		return &n.Token
	case *ast.RaceExpression:
		return &n.Token
	case *ast.SpawnExpression:
		return &n.Token
	case *ast.BreakExpression:
		return &n.Token
	case *ast.ContinueExpression:
		return &n.Token
	default:
		return nil
	}
}
