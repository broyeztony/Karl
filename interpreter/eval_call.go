package interpreter

import "karl/ast"

func (e *Evaluator) evalCallExpression(node *ast.CallExpression, env *Environment) (Value, *Signal, error) {
	function, sig, err := e.Eval(node.Function, env)
	if err != nil || sig != nil {
		return function, sig, err
	}

	args := make([]Value, 0, len(node.Arguments))
	hasDirectPlaceholder := false
	for _, arg := range node.Arguments {
		if _, ok := arg.(*ast.Placeholder); ok {
			args = append(args, nil)
			hasDirectPlaceholder = true
			continue
		}
		toEval := arg
		if expressionContainsPlaceholder(arg) {
			toEval = implicitLambdaFromPlaceholder(arg)
		}

		val, sig, err := e.Eval(toEval, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		args = append(args, val)
	}

	if hasDirectPlaceholder {
		return &Partial{Target: function, Args: args}, nil, nil
	}
	return e.applyFunction(function, args)
}
