package interpreter

import (
	"fmt"
	"karl/ast"
)

func (e *Evaluator) Eval(node ast.Node, env *Environment) (Value, *Signal, error) {
	if err := e.checkRuntimeBeforeEval(); err != nil {
		return nil, nil, err
	}

	val, sig, err := e.evalNode(node, env)
	annotateErrorToken(node, err)
	if fatalErr := e.checkRuntimeAfterEval(sig, err); fatalErr != nil {
		return nil, nil, fatalErr
	}
	return val, sig, err
}

func (e *Evaluator) evalNode(node ast.Node, env *Environment) (Value, *Signal, error) {
	switch n := node.(type) {
	case *ast.Program:
		return e.evalProgram(n, env)
	case *ast.ExpressionStatement:
		return e.Eval(n.Expression, env)
	case *ast.LetStatement:
		val, sig, err := e.Eval(n.Value, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		if ok, err := bindPattern(n.Name, val, env); !ok || err != nil {
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, &RuntimeError{Message: "let pattern did not match"}
		}
		return UnitValue, nil, nil
	case *ast.Identifier:
		return e.evalIdentifier(n, env)
	case *ast.Placeholder:
		return nil, nil, &RuntimeError{Message: "placeholder is only valid in call arguments"}
	case *ast.IntegerLiteral:
		return &Integer{Value: n.Value}, nil, nil
	case *ast.FloatLiteral:
		return &Float{Value: n.Value}, nil, nil
	case *ast.StringLiteral:
		return &String{Value: n.Value}, nil, nil
	case *ast.CharLiteral:
		return &Char{Value: n.Value}, nil, nil
	case *ast.BooleanLiteral:
		return &Boolean{Value: n.Value}, nil, nil
	case *ast.NullLiteral:
		return NullValue, nil, nil
	case *ast.UnitLiteral:
		return UnitValue, nil, nil
	case *ast.PrefixExpression:
		return e.evalPrefixExpression(n, env)
	case *ast.InfixExpression:
		return e.evalInfixExpression(n, env)
	case *ast.AssignExpression:
		return e.evalAssignExpression(n, env)
	case *ast.PostfixExpression:
		return e.evalPostfixExpression(n, env)
	case *ast.AwaitExpression:
		return e.evalAwaitExpression(n, env)
	case *ast.ImportExpression:
		return e.evalImportExpression(n, env)
	case *ast.RecoverExpression:
		return e.evalRecoverExpression(n, env)
	case *ast.IfExpression:
		return e.evalIfExpression(n, env)
	case *ast.BlockExpression:
		return e.evalBlockExpression(n, env)
	case *ast.MatchExpression:
		return e.evalMatchExpression(n, env)
	case *ast.ForExpression:
		return e.evalForExpression(n, env)
	case *ast.LambdaExpression:
		return &Function{Params: n.Params, Body: n.Body, Env: env}, nil, nil
	case *ast.CallExpression:
		return e.evalCallExpression(n, env)
	case *ast.MemberExpression:
		return e.evalMemberExpression(n, env)
	case *ast.IndexExpression:
		return e.evalIndexExpression(n, env)
	case *ast.SliceExpression:
		return e.evalSliceExpression(n, env)
	case *ast.ArrayLiteral:
		return e.evalArrayLiteral(n, env)
	case *ast.ObjectLiteral:
		return e.evalObjectLiteral(n, env)
	case *ast.StructInitExpression:
		return e.evalStructInitExpression(n, env)
	case *ast.RangeExpression:
		return e.evalRangeExpression(n, env)
	case *ast.QueryExpression:
		return e.evalQueryExpression(n, env)
	case *ast.RaceExpression:
		return e.evalRaceExpression(n, env)
	case *ast.SpawnExpression:
		return e.evalSpawnExpression(n, env)
	case *ast.BreakExpression:
		return e.evalBreakExpression(n, env)
	case *ast.ContinueExpression:
		return UnitValue, &Signal{Type: SignalContinue}, nil
	default:
		return nil, nil, &RuntimeError{Message: fmt.Sprintf("unsupported node: %T", node)}
	}
}
