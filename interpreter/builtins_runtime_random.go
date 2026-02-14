package interpreter

import (
	"math"
	"math/rand"
)

func builtinRand(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "rand expects no arguments"}
	}
	return &Integer{Value: rand.Int63()}, nil
}

func builtinRandInt(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "randInt expects min and max"}
	}
	min, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "randInt expects integer min"}
	}
	max, ok := args[1].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "randInt expects integer max"}
	}
	if max.Value < min.Value {
		return nil, &RuntimeError{Message: "randInt expects min <= max"}
	}
	if max.Value == min.Value {
		return &Integer{Value: min.Value}, nil
	}
	diff := max.Value - min.Value
	if diff < 0 || diff == math.MaxInt64 {
		return nil, &RuntimeError{Message: "randInt range too large"}
	}
	n := rand.Int63n(diff+1) + min.Value
	return &Integer{Value: n}, nil
}

func builtinRandFloat(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "randFloat expects min and max"}
	}
	min, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "randFloat expects numeric min"}
	}
	max, _, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "randFloat expects numeric max"}
	}
	if max < min {
		return nil, &RuntimeError{Message: "randFloat expects min <= max"}
	}
	return &Float{Value: min + rand.Float64()*(max-min)}, nil
}
