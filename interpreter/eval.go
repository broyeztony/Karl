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

func evalIntegerInfix(op string, left *Integer, right Value) (Value, *Signal, error) {
	switch r := right.(type) {
	case *Integer:
		return evalNumericInfix(op, float64(left.Value), float64(r.Value), true)
	case *Float:
		return evalNumericInfix(op, float64(left.Value), r.Value, false)
	default:
		return nil, nil, &RuntimeError{Message: "type mismatch in integer operation"}
	}
}

func evalFloatInfix(op string, left *Float, right Value) (Value, *Signal, error) {
	switch r := right.(type) {
	case *Integer:
		return evalNumericInfix(op, left.Value, float64(r.Value), false)
	case *Float:
		return evalNumericInfix(op, left.Value, r.Value, false)
	default:
		return nil, nil, &RuntimeError{Message: "type mismatch in float operation"}
	}
}

func evalNumericInfix(op string, left, right float64, intResult bool) (Value, *Signal, error) {
	switch op {
	case "+":
		return numberValue(left+right, intResult), nil, nil
	case "-":
		return numberValue(left-right, intResult), nil, nil
	case "*":
		return numberValue(left*right, intResult), nil, nil
	case "/":
		return &Float{Value: left / right}, nil, nil
	case "%":
		if intResult {
			return &Integer{Value: int64(left) % int64(right)}, nil, nil
		}
		return nil, nil, &RuntimeError{Message: "modulo requires integers"}
	case "<":
		return &Boolean{Value: left < right}, nil, nil
	case "<=":
		return &Boolean{Value: left <= right}, nil, nil
	case ">":
		return &Boolean{Value: left > right}, nil, nil
	case ">=":
		return &Boolean{Value: left >= right}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported numeric operator: " + op}
	}
}

func numberValue(val float64, intResult bool) Value {
	if intResult {
		return &Integer{Value: int64(val)}
	}
	return &Float{Value: val}
}

func evalStringInfix(op string, left *String, right Value) (Value, *Signal, error) {
	var r *String
	switch v := right.(type) {
	case *String:
		r = v
	case *Char:
		r = &String{Value: v.Value}
	default:
		return nil, nil, &RuntimeError{Message: "string operations require strings"}
	}
	switch op {
	case "+":
		return &String{Value: left.Value + r.Value}, nil, nil
	case "==":
		return &Boolean{Value: left.Value == r.Value}, nil, nil
	case "!=":
		return &Boolean{Value: left.Value != r.Value}, nil, nil
	case "<":
		return &Boolean{Value: left.Value < r.Value}, nil, nil
	case "<=":
		return &Boolean{Value: left.Value <= r.Value}, nil, nil
	case ">":
		return &Boolean{Value: left.Value > r.Value}, nil, nil
	case ">=":
		return &Boolean{Value: left.Value >= r.Value}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported string operator: " + op}
	}
}

func evalArrayInfix(op string, left *Array, right Value) (Value, *Signal, error) {
	r, ok := right.(*Array)
	if !ok {
		return nil, nil, &RuntimeError{Message: "array operation requires array"}
	}
	switch op {
	case "+":
		out := append([]Value{}, left.Elements...)
		out = append(out, r.Elements...)
		return &Array{Elements: out}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported array operator: " + op}
	}
}
