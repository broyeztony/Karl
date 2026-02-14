package interpreter

import (
	"karl/ast"
	"karl/token"
)

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
