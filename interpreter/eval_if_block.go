package interpreter

import "karl/ast"

func (e *Evaluator) evalIfExpression(node *ast.IfExpression, env *Environment) (Value, *Signal, error) {
	cond, sig, err := e.Eval(node.Condition, env)
	if err != nil || sig != nil {
		return cond, sig, err
	}
	if isTruthy(cond) {
		return e.Eval(node.Consequence, env)
	}
	if node.Alternative != nil {
		return e.Eval(node.Alternative, env)
	}
	return UnitValue, nil, nil
}

func (e *Evaluator) evalBlockExpression(block *ast.BlockExpression, env *Environment) (Value, *Signal, error) {
	blockEnv := NewEnclosedEnvironment(env)
	var result Value = UnitValue
	for _, stmt := range block.Statements {
		val, sig, err := e.Eval(stmt, blockEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		result = val
	}
	return result, nil, nil
}
