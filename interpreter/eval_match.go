package interpreter

import "karl/ast"

func (e *Evaluator) evalMatchExpression(node *ast.MatchExpression, env *Environment) (Value, *Signal, error) {
	value, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return value, sig, err
	}
	for _, arm := range node.Arms {
		armEnv := NewEnclosedEnvironment(env)
		ok, err := matchPattern(arm.Pattern, value, armEnv)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			continue
		}
		if arm.Guard != nil {
			guardVal, sig, err := e.Eval(arm.Guard, armEnv)
			if err != nil || sig != nil {
				return guardVal, sig, err
			}
			if !isTruthy(guardVal) {
				continue
			}
		}
		return e.Eval(arm.Body, armEnv)
	}
	return nil, nil, &RuntimeError{Message: "non-exhaustive match"}
}
