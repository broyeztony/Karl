package interpreter

import "karl/ast"

func (e *Evaluator) evalRecoverExpression(node *ast.RecoverExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.Eval(node.Target, env)
	if err == nil && sig == nil {
		return val, nil, nil
	}
	if sig != nil {
		return val, sig, err
	}
	switch err.(type) {
	case *RecoverableError, *RuntimeError:
		// recover block handles both recoverable and runtime errors.
	default:
		return nil, nil, err
	}
	fallbackEnv := NewEnclosedEnvironment(env)
	fallbackEnv.Define("error", errorValue(err))
	return e.Eval(node.Fallback, fallbackEnv)
}
