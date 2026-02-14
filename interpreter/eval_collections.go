package interpreter

import (
	"karl/ast"
)

func (e *Evaluator) evalArrayLiteral(node *ast.ArrayLiteral, env *Environment) (Value, *Signal, error) {
	elements := make([]Value, 0, len(node.Elements))
	for _, el := range node.Elements {
		val, sig, err := e.Eval(el, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		elements = append(elements, val)
	}
	return &Array{Elements: elements}, nil, nil
}

func (e *Evaluator) evalObjectLiteral(node *ast.ObjectLiteral, env *Environment) (Value, *Signal, error) {
	obj := &Object{Pairs: make(map[string]Value)}
	for _, entry := range node.Entries {
		if entry.Spread {
			val, sig, err := e.Eval(entry.Value, env)
			if err != nil || sig != nil {
				return val, sig, err
			}
			pairs, ok := objectPairs(val)
			if !ok {
				return nil, nil, &RuntimeError{Message: "object spread requires object"}
			}
			for k, v := range pairs {
				obj.Pairs[k] = v
			}
			continue
		}

		val, sig, err := e.Eval(entry.Value, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		obj.Pairs[entry.Key] = val
	}
	return obj, nil, nil
}

func (e *Evaluator) evalStructInitExpression(node *ast.StructInitExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.evalObjectLiteral(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	return val, nil, nil
}
