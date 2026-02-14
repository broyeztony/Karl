package interpreter

import "karl/ast"

func (e *Evaluator) evalPrefixExpression(node *ast.PrefixExpression, env *Environment) (Value, *Signal, error) {
	right, sig, err := e.Eval(node.Right, env)
	if err != nil || sig != nil {
		return right, sig, err
	}
	switch node.Operator {
	case "!":
		// Support truthy/falsy evaluation for negation
		return &Boolean{Value: !isTruthy(right)}, nil, nil
	case "-":
		switch v := right.(type) {
		case *Integer:
			return &Integer{Value: -v.Value}, nil, nil
		case *Float:
			return &Float{Value: -v.Value}, nil, nil
		default:
			return nil, nil, &RuntimeError{Message: "operator - expects number"}
		}
	default:
		return nil, nil, &RuntimeError{Message: "unknown prefix operator: " + node.Operator}
	}
}

func (e *Evaluator) evalInfixExpression(node *ast.InfixExpression, env *Environment) (Value, *Signal, error) {
	left, sig, err := e.Eval(node.Left, env)
	if err != nil || sig != nil {
		return left, sig, err
	}

	if node.Operator == "&&" || node.Operator == "||" {
		// Support truthy/falsy evaluation for logical operators
		leftTruthy := isTruthy(left)
		if node.Operator == "&&" && !leftTruthy {
			return &Boolean{Value: false}, nil, nil
		}
		if node.Operator == "||" && leftTruthy {
			return &Boolean{Value: true}, nil, nil
		}
		right, sig, err := e.Eval(node.Right, env)
		if err != nil || sig != nil {
			return right, sig, err
		}
		return &Boolean{Value: isTruthy(right)}, nil, nil
	}

	right, sig, err := e.Eval(node.Right, env)
	if err != nil || sig != nil {
		return right, sig, err
	}

	switch node.Operator {
	case "==":
		return &Boolean{Value: StrictEqual(left, right)}, nil, nil
	case "!=":
		return &Boolean{Value: !StrictEqual(left, right)}, nil, nil
	case "eqv":
		return &Boolean{Value: Equivalent(left, right)}, nil, nil
	}

	switch l := left.(type) {
	case *Integer:
		return evalIntegerInfix(node.Operator, l, right)
	case *Float:
		return evalFloatInfix(node.Operator, l, right)
	case *String:
		return evalStringInfix(node.Operator, l, right)
	case *Char:
		return evalStringInfix(node.Operator, &String{Value: l.Value}, right)
	case *Array:
		return evalArrayInfix(node.Operator, l, right)
	default:
		return nil, nil, &RuntimeError{Message: "unsupported infix operator: " + node.Operator}
	}
}
