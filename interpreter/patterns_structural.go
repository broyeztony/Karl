package interpreter

import (
	"fmt"
	"karl/ast"
)

func matchObjectPattern(p *ast.ObjectPattern, value Value, env *Environment) (bool, error) {
	obj, ok := objectPairs(value)
	if !ok {
		return false, nil
	}
	for _, entry := range p.Entries {
		val, ok := obj[entry.Key]
		if !ok {
			return false, nil
		}
		ok, err := matchPattern(entry.Pattern, val, env)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func matchArrayPattern(p *ast.ArrayPattern, value Value, env *Environment) (bool, error) {
	arr, ok := value.(*Array)
	if !ok {
		return false, nil
	}
	if p.Rest == nil && len(arr.Elements) != len(p.Elements) {
		return false, nil
	}
	if p.Rest != nil && len(arr.Elements) < len(p.Elements) {
		return false, nil
	}
	for i, el := range p.Elements {
		ok, err := matchPattern(el, arr.Elements[i], env)
		if err != nil || !ok {
			return ok, err
		}
	}
	if p.Rest != nil {
		rest := &Array{Elements: append([]Value{}, arr.Elements[len(p.Elements):]...)}
		ok, err := matchPattern(p.Rest, rest, env)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func matchTuplePattern(p *ast.TuplePattern, value Value, env *Environment) (bool, error) {
	arr, ok := value.(*Array)
	if !ok {
		return false, nil
	}
	if len(arr.Elements) != len(p.Elements) {
		return false, nil
	}
	for i, el := range p.Elements {
		ok, err := matchPattern(el, arr.Elements[i], env)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func matchCallPattern(p *ast.CallPattern, value Value, env *Environment) (bool, error) {
	return false, fmt.Errorf("call patterns are not supported")
}
