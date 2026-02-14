package interpreter

import "karl/ast"

func (e *Evaluator) applyFunction(fn Value, args []Value) (Value, *Signal, error) {
	switch f := fn.(type) {
	case *Builtin:
		val, err := f.Fn(e, args)
		return val, nil, err
	case *Function:
		if len(args) != len(f.Params) {
			return nil, nil, &RuntimeError{Message: "wrong number of arguments"}
		}
		extended := NewEnclosedEnvironment(f.Env)
		for i, param := range f.Params {
			ok, err := bindPattern(param, args[i], extended)
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				return nil, nil, &RuntimeError{Message: "parameter pattern did not match"}
			}
		}
		val, sig, err := e.Eval(f.Body, extended)
		if err != nil {
			return nil, nil, err
		}
		if sig != nil {
			return nil, nil, &RuntimeError{Message: "break/continue outside loop"}
		}
		return val, nil, nil
	case *Partial:
		filled := []Value{}
		argIndex := 0
		for _, arg := range f.Args {
			if arg == nil {
				if argIndex >= len(args) {
					return nil, nil, &RuntimeError{Message: "not enough arguments for partial"}
				}
				filled = append(filled, args[argIndex])
				argIndex++
				continue
			}
			filled = append(filled, arg)
		}
		if argIndex != len(args) {
			return nil, nil, &RuntimeError{Message: "too many arguments for partial"}
		}
		return e.applyFunction(f.Target, filled)
	default:
		return nil, nil, &RuntimeError{Message: "not a function"}
	}
}

func bindPattern(pattern ast.Pattern, value Value, env *Environment) (bool, error) {
	return matchPattern(pattern, value, env)
}
