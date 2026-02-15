package interpreter

import "karl/ast"

func (e *Evaluator) evalIndexExpression(node *ast.IndexExpression, env *Environment) (Value, *Signal, error) {
	left, sig, err := e.Eval(node.Left, env)
	if err != nil || sig != nil {
		return left, sig, err
	}
	indexVal, sig, err := e.Eval(node.Index, env)
	if err != nil || sig != nil {
		return indexVal, sig, err
	}

	switch indexed := left.(type) {
	case *Array:
		idx, ok := indexVal.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "index must be integer"}
		}
		i := int(idx.Value)
		if i < 0 || i >= len(indexed.Elements) {
			return nil, nil, &RuntimeError{Message: "index out of bounds"}
		}
		return indexed.Elements[i], nil, nil
	case *Object:
		key, ok := objectIndexKey(indexVal)
		if !ok {
			return nil, nil, &RuntimeError{Message: "object index must be string or char"}
		}
		val, ok := indexed.Pairs[key]
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + key}
		}
		return val, nil, nil
	case *ModuleObject:
		if indexed.Env == nil {
			return nil, nil, &RuntimeError{Message: "indexing requires array or object"}
		}
		key, ok := objectIndexKey(indexVal)
		if !ok {
			return nil, nil, &RuntimeError{Message: "object index must be string or char"}
		}
		val, ok := indexed.Env.GetLocal(key)
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + key}
		}
		return val, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "indexing requires array or object"}
	}
}

func (e *Evaluator) evalSliceExpression(node *ast.SliceExpression, env *Environment) (Value, *Signal, error) {
	left, sig, err := e.Eval(node.Left, env)
	if err != nil || sig != nil {
		return left, sig, err
	}
	switch sliced := left.(type) {
	case *Array:
		start, end, val, sig, err := e.evalSliceBounds(node, env, len(sliced.Elements))
		if err != nil || sig != nil {
			return val, sig, err
		}
		if start >= end {
			return &Array{Elements: []Value{}}, nil, nil
		}
		out := append([]Value{}, sliced.Elements[start:end]...)
		return &Array{Elements: out}, nil, nil
	case *String:
		runes := []rune(sliced.Value)
		start, end, val, sig, err := e.evalSliceBounds(node, env, len(runes))
		if err != nil || sig != nil {
			return val, sig, err
		}
		if start >= end {
			return &String{Value: ""}, nil, nil
		}
		return &String{Value: string(runes[start:end])}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "slice requires array or string"}
	}
}

func (e *Evaluator) evalSliceBounds(node *ast.SliceExpression, env *Environment, length int) (int, int, Value, *Signal, error) {
	start := 0
	end := length

	if node.Start != nil {
		val, sig, err := e.Eval(node.Start, env)
		if err != nil || sig != nil {
			return 0, 0, val, sig, err
		}
		idx, ok := val.(*Integer)
		if !ok {
			return 0, 0, nil, nil, &RuntimeError{Message: "slice start must be integer"}
		}
		start = normalizeIndex(int(idx.Value), length)
	}

	if node.End != nil {
		val, sig, err := e.Eval(node.End, env)
		if err != nil || sig != nil {
			return 0, 0, val, sig, err
		}
		idx, ok := val.(*Integer)
		if !ok {
			return 0, 0, nil, nil, &RuntimeError{Message: "slice end must be integer"}
		}
		end = normalizeIndex(int(idx.Value), length)
	}

	if start < 0 || start > length || end < 0 || end > length {
		return 0, 0, nil, nil, &RuntimeError{Message: "slice bounds out of range"}
	}

	return start, end, nil, nil, nil
}

func normalizeIndex(idx int, length int) int {
	if idx < 0 {
		return length + idx
	}
	return idx
}
