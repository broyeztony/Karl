package interpreter

import (
	"fmt"
	"karl/ast"
	"unicode/utf8"
)

func (e *Evaluator) evalMemberExpression(node *ast.MemberExpression, env *Environment) (Value, *Signal, error) {
	object, sig, err := e.Eval(node.Object, env)
	if err != nil || sig != nil {
		return object, sig, err
	}

	switch obj := object.(type) {
	case *Object:
		val, ok := obj.Pairs[node.Property.Value]
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + node.Property.Value}
		}
		return val, nil, nil
	case *ModuleObject:
		if obj.Env == nil {
			return nil, nil, &RuntimeError{Message: "member access on invalid module object"}
		}
		val, ok := obj.Env.GetLocal(node.Property.Value)
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + node.Property.Value}
		}
		return val, nil, nil
	case *Array:
		if node.Property.Value == "length" {
			return &Integer{Value: int64(len(obj.Elements))}, nil, nil
		}
		return e.arrayMethod(obj, node.Property.Value)
	case *String:
		if node.Property.Value == "length" {
			return &Integer{Value: int64(utf8.RuneCountInString(obj.Value))}, nil, nil
		}
		return e.stringMethod(obj, node.Property.Value)
	case *Map:
		return e.mapMethod(obj, node.Property.Value)
	case *Set:
		if node.Property.Value == "size" {
			return &Integer{Value: int64(len(obj.Elements))}, nil, nil
		}
		return e.setMethod(obj, node.Property.Value)
	case *Channel:
		return e.channelMethod(obj, node.Property.Value)
	case *Task:
		return e.taskMethod(obj, node.Property.Value)
	default:
		if object == nil {
			return nil, nil, &RuntimeError{Message: "member access on non-object (got <nil>)"}
		}
		return nil, nil, &RuntimeError{Message: fmt.Sprintf("member access on non-object (%s.%s)", object.Type(), node.Property.Value)}
	}
}
