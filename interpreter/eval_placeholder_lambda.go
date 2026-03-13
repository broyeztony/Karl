package interpreter

import (
	"karl/ast"
	"karl/token"
)

const implicitPlaceholderParam = "__karl_it"

func implicitLambdaFromPlaceholder(expr ast.Expression) ast.Expression {
	param := &ast.Identifier{
		Token: token.Token{
			Type:    token.IDENT,
			Literal: implicitPlaceholderParam,
		},
		Value: implicitPlaceholderParam,
	}
	return &ast.LambdaExpression{
		Token: token.Token{
			Type:    token.ARROW,
			Literal: "->",
		},
		Params: []ast.Pattern{param},
		Body:   replacePlaceholderInExpr(expr, param),
	}
}

func expressionContainsPlaceholder(expr ast.Expression) bool {
	switch n := expr.(type) {
	case nil:
		return false
	case *ast.Placeholder:
		return true
	case *ast.LambdaExpression:
		return false
	case *ast.PrefixExpression:
		return expressionContainsPlaceholder(n.Right)
	case *ast.InfixExpression:
		return expressionContainsPlaceholder(n.Left) || expressionContainsPlaceholder(n.Right)
	case *ast.AssignExpression:
		return expressionContainsPlaceholder(n.Left) || expressionContainsPlaceholder(n.Right)
	case *ast.PostfixExpression:
		return expressionContainsPlaceholder(n.Left)
	case *ast.AwaitExpression:
		return expressionContainsPlaceholder(n.Value)
	case *ast.IfExpression:
		return expressionContainsPlaceholder(n.Condition) ||
			blockContainsPlaceholder(n.Consequence) ||
			expressionContainsPlaceholder(n.Alternative)
	case *ast.BlockExpression:
		return blockContainsPlaceholder(n)
	case *ast.MatchExpression:
		if expressionContainsPlaceholder(n.Value) {
			return true
		}
		for _, arm := range n.Arms {
			if expressionContainsPlaceholder(arm.Guard) || expressionContainsPlaceholder(arm.Body) {
				return true
			}
		}
		return false
	case *ast.ForExpression:
		if expressionContainsPlaceholder(n.Condition) || blockContainsPlaceholder(n.Body) || expressionContainsPlaceholder(n.Then) {
			return true
		}
		for _, b := range n.Bindings {
			if expressionContainsPlaceholder(b.Value) {
				return true
			}
		}
		return false
	case *ast.CallExpression:
		if expressionContainsPlaceholder(n.Function) {
			return true
		}
		for _, arg := range n.Arguments {
			if expressionContainsPlaceholder(arg) {
				return true
			}
		}
		return false
	case *ast.RecoverExpression:
		return expressionContainsPlaceholder(n.Target) || expressionContainsPlaceholder(n.Fallback)
	case *ast.MemberExpression:
		return expressionContainsPlaceholder(n.Object)
	case *ast.IndexExpression:
		return expressionContainsPlaceholder(n.Left) || expressionContainsPlaceholder(n.Index)
	case *ast.SliceExpression:
		return expressionContainsPlaceholder(n.Left) ||
			expressionContainsPlaceholder(n.Start) ||
			expressionContainsPlaceholder(n.End)
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if expressionContainsPlaceholder(el) {
				return true
			}
		}
		return false
	case *ast.ObjectLiteral:
		for _, entry := range n.Entries {
			if expressionContainsPlaceholder(entry.Value) {
				return true
			}
		}
		return false
	case *ast.StructInitExpression:
		return expressionContainsPlaceholder(n.Value)
	case *ast.RangeExpression:
		return expressionContainsPlaceholder(n.Start) ||
			expressionContainsPlaceholder(n.End) ||
			expressionContainsPlaceholder(n.Step)
	case *ast.QueryExpression:
		if expressionContainsPlaceholder(n.Source) || expressionContainsPlaceholder(n.OrderBy) || expressionContainsPlaceholder(n.Select) {
			return true
		}
		for _, w := range n.Where {
			if expressionContainsPlaceholder(w) {
				return true
			}
		}
		return false
	case *ast.RaceExpression:
		for _, task := range n.Tasks {
			if expressionContainsPlaceholder(task) {
				return true
			}
		}
		return false
	case *ast.SpawnExpression:
		if expressionContainsPlaceholder(n.Task) {
			return true
		}
		for _, task := range n.Group {
			if expressionContainsPlaceholder(task) {
				return true
			}
		}
		return false
	case *ast.BreakExpression:
		return expressionContainsPlaceholder(n.Value)
	default:
		return false
	}
}

func blockContainsPlaceholder(block *ast.BlockExpression) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			if expressionContainsPlaceholder(s.Value) {
				return true
			}
		case *ast.ExpressionStatement:
			if expressionContainsPlaceholder(s.Expression) {
				return true
			}
		}
	}
	return false
}

func replacePlaceholderInExpr(expr ast.Expression, param *ast.Identifier) ast.Expression {
	switch n := expr.(type) {
	case nil:
		return nil
	case *ast.Placeholder:
		return &ast.Identifier{Token: param.Token, Value: param.Value}
	case *ast.PrefixExpression:
		return &ast.PrefixExpression{Token: n.Token, Operator: n.Operator, Right: replacePlaceholderInExpr(n.Right, param)}
	case *ast.InfixExpression:
		return &ast.InfixExpression{
			Token:    n.Token,
			Left:     replacePlaceholderInExpr(n.Left, param),
			Operator: n.Operator,
			Right:    replacePlaceholderInExpr(n.Right, param),
		}
	case *ast.AssignExpression:
		return &ast.AssignExpression{
			Token:    n.Token,
			Left:     replacePlaceholderInExpr(n.Left, param),
			Operator: n.Operator,
			Right:    replacePlaceholderInExpr(n.Right, param),
		}
	case *ast.PostfixExpression:
		return &ast.PostfixExpression{Token: n.Token, Left: replacePlaceholderInExpr(n.Left, param), Operator: n.Operator}
	case *ast.AwaitExpression:
		return &ast.AwaitExpression{Token: n.Token, Value: replacePlaceholderInExpr(n.Value, param)}
	case *ast.IfExpression:
		return &ast.IfExpression{
			Token:       n.Token,
			Condition:   replacePlaceholderInExpr(n.Condition, param),
			Consequence: replacePlaceholderInBlock(n.Consequence, param),
			Alternative: replacePlaceholderInExpr(n.Alternative, param),
		}
	case *ast.BlockExpression:
		return replacePlaceholderInBlock(n, param)
	case *ast.MatchExpression:
		arms := make([]ast.MatchArm, 0, len(n.Arms))
		for _, arm := range n.Arms {
			arms = append(arms, ast.MatchArm{
				Token:   arm.Token,
				Pattern: arm.Pattern,
				Guard:   replacePlaceholderInExpr(arm.Guard, param),
				Body:    replacePlaceholderInExpr(arm.Body, param),
			})
		}
		return &ast.MatchExpression{
			Token: n.Token,
			Value: replacePlaceholderInExpr(n.Value, param),
			Arms:  arms,
		}
	case *ast.ForExpression:
		bindings := make([]ast.Binding, 0, len(n.Bindings))
		for _, b := range n.Bindings {
			bindings = append(bindings, ast.Binding{
				Pattern: b.Pattern,
				Value:   replacePlaceholderInExpr(b.Value, param),
			})
		}
		return &ast.ForExpression{
			Token:     n.Token,
			Condition: replacePlaceholderInExpr(n.Condition, param),
			Bindings:  bindings,
			Body:      replacePlaceholderInBlock(n.Body, param),
			Then:      replacePlaceholderInExpr(n.Then, param),
		}
	case *ast.CallExpression:
		args := make([]ast.Expression, 0, len(n.Arguments))
		for _, arg := range n.Arguments {
			args = append(args, replacePlaceholderInExpr(arg, param))
		}
		return &ast.CallExpression{
			Token:     n.Token,
			Function:  replacePlaceholderInExpr(n.Function, param),
			Arguments: args,
		}
	case *ast.RecoverExpression:
		return &ast.RecoverExpression{
			Token:    n.Token,
			Target:   replacePlaceholderInExpr(n.Target, param),
			Fallback: replacePlaceholderInExpr(n.Fallback, param),
		}
	case *ast.MemberExpression:
		return &ast.MemberExpression{
			Token:    n.Token,
			Object:   replacePlaceholderInExpr(n.Object, param),
			Property: n.Property,
		}
	case *ast.IndexExpression:
		return &ast.IndexExpression{
			Token: n.Token,
			Left:  replacePlaceholderInExpr(n.Left, param),
			Index: replacePlaceholderInExpr(n.Index, param),
		}
	case *ast.SliceExpression:
		return &ast.SliceExpression{
			Token: n.Token,
			Left:  replacePlaceholderInExpr(n.Left, param),
			Start: replacePlaceholderInExpr(n.Start, param),
			End:   replacePlaceholderInExpr(n.End, param),
		}
	case *ast.ArrayLiteral:
		out := make([]ast.Expression, 0, len(n.Elements))
		for _, el := range n.Elements {
			out = append(out, replacePlaceholderInExpr(el, param))
		}
		return &ast.ArrayLiteral{Token: n.Token, Elements: out}
	case *ast.ObjectLiteral:
		out := make([]ast.ObjectEntry, 0, len(n.Entries))
		for _, entry := range n.Entries {
			out = append(out, ast.ObjectEntry{
				Token:     entry.Token,
				Key:       entry.Key,
				Value:     replacePlaceholderInExpr(entry.Value, param),
				Shorthand: entry.Shorthand,
				Spread:    entry.Spread,
			})
		}
		return &ast.ObjectLiteral{Token: n.Token, Entries: out}
	case *ast.StructInitExpression:
		obj, _ := replacePlaceholderInExpr(n.Value, param).(*ast.ObjectLiteral)
		return &ast.StructInitExpression{
			Token:    n.Token,
			TypeName: n.TypeName,
			Value:    obj,
		}
	case *ast.RangeExpression:
		return &ast.RangeExpression{
			Token: n.Token,
			Start: replacePlaceholderInExpr(n.Start, param),
			End:   replacePlaceholderInExpr(n.End, param),
			Step:  replacePlaceholderInExpr(n.Step, param),
		}
	case *ast.QueryExpression:
		where := make([]ast.Expression, 0, len(n.Where))
		for _, w := range n.Where {
			where = append(where, replacePlaceholderInExpr(w, param))
		}
		return &ast.QueryExpression{
			Token:   n.Token,
			Var:     n.Var,
			Source:  replacePlaceholderInExpr(n.Source, param),
			Where:   where,
			OrderBy: replacePlaceholderInExpr(n.OrderBy, param),
			Select:  replacePlaceholderInExpr(n.Select, param),
		}
	case *ast.RaceExpression:
		tasks := make([]ast.Expression, 0, len(n.Tasks))
		for _, task := range n.Tasks {
			tasks = append(tasks, replacePlaceholderInExpr(task, param))
		}
		return &ast.RaceExpression{Token: n.Token, Tasks: tasks}
	case *ast.SpawnExpression:
		group := make([]ast.Expression, 0, len(n.Group))
		for _, task := range n.Group {
			group = append(group, replacePlaceholderInExpr(task, param))
		}
		return &ast.SpawnExpression{
			Token: n.Token,
			Task:  replacePlaceholderInExpr(n.Task, param),
			Group: group,
		}
	case *ast.BreakExpression:
		return &ast.BreakExpression{Token: n.Token, Value: replacePlaceholderInExpr(n.Value, param)}
	default:
		return expr
	}
}

func replacePlaceholderInBlock(block *ast.BlockExpression, param *ast.Identifier) *ast.BlockExpression {
	if block == nil {
		return nil
	}
	out := make([]ast.Statement, 0, len(block.Statements))
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			out = append(out, &ast.LetStatement{
				Token: s.Token,
				Name:  s.Name,
				Value: replacePlaceholderInExpr(s.Value, param),
			})
		case *ast.ExpressionStatement:
			out = append(out, &ast.ExpressionStatement{
				Token:      s.Token,
				Expression: replacePlaceholderInExpr(s.Expression, param),
			})
		default:
			out = append(out, stmt)
		}
	}
	return &ast.BlockExpression{Token: block.Token, Statements: out}
}
