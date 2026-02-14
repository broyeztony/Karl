package interpreter

import "karl/ast"

func errorValue(err error) Value {
	switch e := err.(type) {
	case *RecoverableError:
		kind := e.Kind
		if kind == "" {
			kind = "error"
		}
		return &Object{Pairs: map[string]Value{
			"kind":    &String{Value: kind},
			"message": &String{Value: e.Message},
		}}
	case *RuntimeError:
		return &Object{Pairs: map[string]Value{
			"kind":    &String{Value: "runtime"},
			"message": &String{Value: e.Message},
		}}
	default:
		return &Object{Pairs: map[string]Value{
			"kind":    &String{Value: "error"},
			"message": &String{Value: err.Error()},
		}}
	}
}

func (e *Evaluator) evalProgram(program *ast.Program, env *Environment) (Value, *Signal, error) {
	var result Value = UnitValue
	for _, stmt := range program.Statements {
		val, sig, err := e.Eval(stmt, env)
		if err != nil {
			return nil, nil, err
		}
		if sig != nil {
			return nil, nil, &RuntimeError{Message: "break/continue outside loop"}
		}
		result = val
	}
	return result, nil, nil
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier, env *Environment) (Value, *Signal, error) {
	val, ok := env.Get(node.Value)
	if !ok {
		return nil, nil, &RuntimeError{Message: "undefined identifier: " + node.Value}
	}
	return val, nil, nil
}
