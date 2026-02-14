package interpreter

import "karl/ast"

func (e *Evaluator) evalAssignExpression(node *ast.AssignExpression, env *Environment) (Value, *Signal, error) {
	target, setter, err := e.resolveAssignable(node.Left, env)
	if err != nil {
		return nil, nil, err
	}
	right, sig, err := e.Eval(node.Right, env)
	if err != nil || sig != nil {
		return right, sig, err
	}

	var newVal Value
	switch node.Operator {
	case "=":
		newVal = right
	case "+=":
		newVal, err = e.applyBinary("+", target, right)
	case "-=":
		newVal, err = e.applyBinary("-", target, right)
	case "*=":
		newVal, err = e.applyBinary("*", target, right)
	case "/=":
		newVal, err = e.applyBinary("/", target, right)
	case "%=":
		newVal, err = e.applyBinary("%", target, right)
	default:
		return nil, nil, &RuntimeError{Message: "unknown assignment operator: " + node.Operator}
	}
	if err != nil {
		return nil, nil, err
	}
	setter(newVal)
	return newVal, nil, nil
}

func (e *Evaluator) evalPostfixExpression(node *ast.PostfixExpression, env *Environment) (Value, *Signal, error) {
	switch node.Operator {
	case "++", "--":
		target, setter, err := e.resolveAssignable(node.Left, env)
		if err != nil {
			return nil, nil, err
		}
		var delta float64 = 1
		if node.Operator == "--" {
			delta = -1
		}
		switch v := target.(type) {
		case *Integer:
			newVal := &Integer{Value: v.Value + int64(delta)}
			setter(newVal)
			return newVal, nil, nil
		case *Float:
			newVal := &Float{Value: v.Value + delta}
			setter(newVal)
			return newVal, nil, nil
		default:
			return nil, nil, &RuntimeError{Message: "increment/decrement requires number"}
		}
	default:
		return nil, nil, &RuntimeError{Message: "unknown postfix operator: " + node.Operator}
	}
}

func (e *Evaluator) applyBinary(op string, left, right Value) (Value, error) {
	switch op {
	case "+":
		switch l := left.(type) {
		case *Integer:
			val, _, err := evalIntegerInfix(op, l, right)
			return val, err
		case *Float:
			val, _, err := evalFloatInfix(op, l, right)
			return val, err
		case *String:
			val, _, err := evalStringInfix(op, l, right)
			return val, err
		case *Array:
			val, _, err := evalArrayInfix(op, l, right)
			return val, err
		default:
			return nil, &RuntimeError{Message: "unsupported +="}
		}
	case "-", "*", "/", "%":
		switch l := left.(type) {
		case *Integer:
			val, _, err := evalIntegerInfix(op, l, right)
			return val, err
		case *Float:
			val, _, err := evalFloatInfix(op, l, right)
			return val, err
		default:
			return nil, &RuntimeError{Message: "unsupported assignment operator"}
		}
	default:
		return nil, &RuntimeError{Message: "unknown assignment operator"}
	}
}
