package interpreter

import "karl/ast"

func (e *Evaluator) evalCallExpression(node *ast.CallExpression, env *Environment) (Value, *Signal, error) {
	function, sig, err := e.Eval(node.Function, env)
	if err != nil || sig != nil {
		return function, sig, err
	}

	args := make([]Value, 0, len(node.Arguments))
	hasPlaceholder := false
	for _, arg := range node.Arguments {
		if _, ok := arg.(*ast.Placeholder); ok {
			args = append(args, nil)
			hasPlaceholder = true
			continue
		}
		val, sig, err := e.Eval(arg, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		args = append(args, val)
	}

	if hasPlaceholder {
		return &Partial{Target: function, Args: args}, nil, nil
	}
	return e.applyFunction(function, args)
}
