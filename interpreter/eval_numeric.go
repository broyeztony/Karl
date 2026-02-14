package interpreter

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
