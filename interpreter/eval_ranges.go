package interpreter

import (
	"karl/ast"
	"unicode/utf8"
)

func (e *Evaluator) evalRangeExpression(node *ast.RangeExpression, env *Environment) (Value, *Signal, error) {
	startVal, sig, err := e.Eval(node.Start, env)
	if err != nil || sig != nil {
		return startVal, sig, err
	}
	endVal, sig, err := e.Eval(node.End, env)
	if err != nil || sig != nil {
		return endVal, sig, err
	}
	if _, ok := startVal.(*Float); ok {
		return nil, nil, &RuntimeError{Message: "float ranges are not allowed"}
	}
	if _, ok := endVal.(*Float); ok {
		return nil, nil, &RuntimeError{Message: "float ranges are not allowed"}
	}

	stepVal := Value(&Integer{Value: 1})
	if node.Step != nil {
		stepVal, sig, err = e.Eval(node.Step, env)
		if err != nil || sig != nil {
			return stepVal, sig, err
		}
	}

	return buildRange(startVal, endVal, stepVal)
}

func buildRange(start, end, step Value) (Value, *Signal, error) {
	switch s := start.(type) {
	case *Integer:
		e, ok := end.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		st, ok := step.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "integer range step must be integer"}
		}
		if st.Value == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildIntRange(s.Value, e.Value, st.Value), nil, nil
	case *Float:
		e, ok := end.(*Float)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		var stepVal float64
		switch st := step.(type) {
		case *Float:
			stepVal = st.Value
		case *Integer:
			stepVal = float64(st.Value)
		default:
			return nil, nil, &RuntimeError{Message: "float range step must be number"}
		}
		if stepVal == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildFloatRange(s.Value, e.Value, stepVal), nil, nil
	case *Char:
		e, ok := end.(*Char)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		st, ok := step.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "char range step must be integer"}
		}
		if st.Value == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildCharRange(s.Value, e.Value, st.Value), nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported range endpoint type"}
	}
}

func buildIntRange(start, end, step int64) Value {
	out := []Value{}
	if step > 0 {
		for i := start; i <= end; i += step {
			out = append(out, &Integer{Value: i})
		}
	} else {
		for i := start; i >= end; i += step {
			out = append(out, &Integer{Value: i})
		}
	}
	return &Array{Elements: out}
}

func buildFloatRange(start, end, step float64) Value {
	out := []Value{}
	if step > 0 {
		for i := start; i < end; i += step {
			out = append(out, &Float{Value: i})
		}
	} else {
		for i := start; i > end; i += step {
			out = append(out, &Float{Value: i})
		}
	}
	return &Array{Elements: out}
}

func buildCharRange(start, end string, step int64) Value {
	sr, _ := utf8.DecodeRuneInString(start)
	er, _ := utf8.DecodeRuneInString(end)
	out := []Value{}
	if step > 0 {
		for r := sr; r <= er; r += rune(step) {
			out = append(out, &Char{Value: string(r)})
		}
	} else {
		for r := sr; r >= er; r += rune(step) {
			out = append(out, &Char{Value: string(r)})
		}
	}
	return &Array{Elements: out}
}
