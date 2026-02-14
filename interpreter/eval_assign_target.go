package interpreter

import "karl/ast"

func (e *Evaluator) resolveAssignable(node ast.Expression, env *Environment) (Value, func(Value), error) {
	switch n := node.(type) {
	case *ast.Identifier:
		val, ok := env.Get(n.Value)
		if !ok {
			return nil, nil, &RuntimeError{Message: "undefined identifier: " + n.Value}
		}
		return val, func(v Value) { env.Set(n.Value, v) }, nil
	case *ast.MemberExpression:
		objVal, sig, err := e.Eval(n.Object, env)
		if err != nil || sig != nil {
			return nil, nil, err
		}
		switch obj := objVal.(type) {
		case *Object:
			return obj.Pairs[n.Property.Value], func(v Value) { obj.Pairs[n.Property.Value] = v }, nil
		case *ModuleObject:
			if obj.Env == nil {
				return nil, nil, &RuntimeError{Message: "member assignment requires object"}
			}
			val, _ := obj.Env.GetLocal(n.Property.Value)
			return val, func(v Value) { obj.Env.Define(n.Property.Value, v) }, nil
		default:
			return nil, nil, &RuntimeError{Message: "member assignment requires object"}
		}
	case *ast.IndexExpression:
		left, sig, err := e.Eval(n.Left, env)
		if err != nil || sig != nil {
			return nil, nil, err
		}
		indexVal, sig, err := e.Eval(n.Index, env)
		if err != nil || sig != nil {
			return nil, nil, err
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
			return indexed.Elements[i], func(v Value) { indexed.Elements[i] = v }, nil
		case *Object:
			key, ok := objectIndexKey(indexVal)
			if !ok {
				return nil, nil, &RuntimeError{Message: "object index must be string or char"}
			}
			return indexed.Pairs[key], func(v Value) { indexed.Pairs[key] = v }, nil
		case *ModuleObject:
			if indexed.Env == nil {
				return nil, nil, &RuntimeError{Message: "index assignment requires array or object"}
			}
			key, ok := objectIndexKey(indexVal)
			if !ok {
				return nil, nil, &RuntimeError{Message: "object index must be string or char"}
			}
			val, _ := indexed.Env.GetLocal(key)
			return val, func(v Value) { indexed.Env.Define(key, v) }, nil
		default:
			return nil, nil, &RuntimeError{Message: "index assignment requires array or object"}
		}
	default:
		return nil, nil, &RuntimeError{Message: "invalid assignment target"}
	}
}

func objectIndexKey(index Value) (string, bool) {
	switch v := index.(type) {
	case *String:
		return v.Value, true
	case *Char:
		return v.Value, true
	default:
		return "", false
	}
}
