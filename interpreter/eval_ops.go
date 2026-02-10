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
