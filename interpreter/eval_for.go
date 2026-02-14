package interpreter

import "karl/ast"

func (e *Evaluator) evalForExpression(node *ast.ForExpression, env *Environment) (Value, *Signal, error) {
	loopEnv := NewEnclosedEnvironment(env)
	for _, binding := range node.Bindings {
		val, sig, err := e.Eval(binding.Value, loopEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		ok, err := bindPattern(binding.Pattern, val, loopEnv)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, &RuntimeError{Message: "for binding pattern did not match"}
		}
	}

	for {
		condVal, sig, err := e.Eval(node.Condition, loopEnv)
		if err != nil || sig != nil {
			return condVal, sig, err
		}
		if !isTruthy(condVal) {
			break
		}

		_, sig, err = e.Eval(node.Body, loopEnv)
		if err != nil {
			return nil, nil, err
		}
		if sig != nil {
			switch sig.Type {
			case SignalContinue:
				continue
			case SignalBreak:
				if sig.Value != nil {
					return sig.Value, nil, nil
				}
				return e.evalThen(node, loopEnv)
			default:
				return nil, sig, nil
			}
		}
	}

	return e.evalThen(node, loopEnv)
}

func (e *Evaluator) evalThen(node *ast.ForExpression, env *Environment) (Value, *Signal, error) {
	if node.Then == nil {
		return UnitValue, nil, nil
	}
	return e.Eval(node.Then, env)
}

func (e *Evaluator) evalBreakExpression(node *ast.BreakExpression, env *Environment) (Value, *Signal, error) {
	if node.Value == nil {
		return UnitValue, &Signal{Type: SignalBreak}, nil
	}
	val, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	return val, &Signal{Type: SignalBreak, Value: val}, nil
}
