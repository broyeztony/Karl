package interpreter

import "math"

func numberArg(val Value) (float64, bool, bool) {
	switch v := val.(type) {
	case *Integer:
		return float64(v.Value), true, true
	case *Float:
		return v.Value, false, true
	default:
		return 0, false, false
	}
}

func unaryMath(args []Value, name string, fn func(float64) float64) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: name + " expects 1 argument"}
	}
	val, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: name + " expects number"}
	}
	result := fn(val)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return nil, &RuntimeError{Message: name + " result not finite"}
	}
	return &Float{Value: result}, nil
}

func integralMath(args []Value, name string, fn func(float64) float64) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: name + " expects 1 argument"}
	}
	val, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: name + " expects number"}
	}
	result := fn(val)
	if result > float64(math.MaxInt64) || result < float64(math.MinInt64) {
		return nil, &RuntimeError{Message: name + " overflow"}
	}
	return &Integer{Value: int64(result)}, nil
}

func minMax(args []Value, name string, takeMax bool) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: name + " expects 2 arguments"}
	}
	left, leftInt, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: name + " expects number"}
	}
	right, rightInt, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: name + " expects number"}
	}
	var out float64
	if takeMax {
		out = math.Max(left, right)
	} else {
		out = math.Min(left, right)
	}
	if leftInt && rightInt {
		return &Integer{Value: int64(out)}, nil
	}
	return &Float{Value: out}, nil
}
