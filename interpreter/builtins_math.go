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

func builtinAbs(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "abs expects 1 argument"}
	}
	switch v := args[0].(type) {
	case *Integer:
		if v.Value == math.MinInt64 {
			return nil, &RuntimeError{Message: "abs overflow"}
		}
		if v.Value < 0 {
			return &Integer{Value: -v.Value}, nil
		}
		return v, nil
	case *Float:
		return &Float{Value: math.Abs(v.Value)}, nil
	default:
		return nil, &RuntimeError{Message: "abs expects number"}
	}
}

func builtinSqrt(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sqrt expects 1 argument"}
	}
	val, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "sqrt expects number"}
	}
	if val < 0 {
		return nil, &RuntimeError{Message: "sqrt expects non-negative number"}
	}
	return &Float{Value: math.Sqrt(val)}, nil
}

func builtinPow(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "pow expects base and exponent"}
	}
	base, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "pow expects numeric base"}
	}
	exp, _, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "pow expects numeric exponent"}
	}
	result := math.Pow(base, exp)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return nil, &RuntimeError{Message: "pow result not finite"}
	}
	return &Float{Value: result}, nil
}

func builtinSin(_ *Evaluator, args []Value) (Value, error) {
	return unaryMath(args, "sin", math.Sin)
}

func builtinCos(_ *Evaluator, args []Value) (Value, error) {
	return unaryMath(args, "cos", math.Cos)
}

func builtinTan(_ *Evaluator, args []Value) (Value, error) {
	return unaryMath(args, "tan", math.Tan)
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

func builtinFloor(_ *Evaluator, args []Value) (Value, error) {
	return integralMath(args, "floor", math.Floor)
}

func builtinCeil(_ *Evaluator, args []Value) (Value, error) {
	return integralMath(args, "ceil", math.Ceil)
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

func builtinMin(_ *Evaluator, args []Value) (Value, error) {
	return minMax(args, "min", false)
}

func builtinMax(_ *Evaluator, args []Value) (Value, error) {
	return minMax(args, "max", true)
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

func builtinClamp(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "clamp expects value, min, max"}
	}
	val, valInt, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "clamp expects numeric value"}
	}
	min, minInt, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "clamp expects numeric min"}
	}
	max, maxInt, ok := numberArg(args[2])
	if !ok {
		return nil, &RuntimeError{Message: "clamp expects numeric max"}
	}
	if min > max {
		return nil, &RuntimeError{Message: "clamp expects min <= max"}
	}
	out := val
	if out < min {
		out = min
	}
	if out > max {
		out = max
	}
	if valInt && minInt && maxInt {
		return &Integer{Value: int64(out)}, nil
	}
	return &Float{Value: out}, nil
}
