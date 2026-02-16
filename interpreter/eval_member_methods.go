package interpreter

func (e *Evaluator) arrayMethod(arr *Array, name string) (Value, *Signal, error) {
	switch name {
	case "map", "filter", "reduce", "sum", "find", "sort":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, arr)}, nil, nil
	case "push":
		return &Builtin{
			Name: "push",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "push expects 1 argument"}
				}
				arr.Elements = append(arr.Elements, args[0])
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown array member: " + name}
	}
}

func (e *Evaluator) channelMethod(ch *Channel, name string) (Value, *Signal, error) {
	switch name {
	case "send", "recv", "done":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, ch)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown channel member: " + name}
	}
}

func (e *Evaluator) stringMethod(str *String, name string) (Value, *Signal, error) {
	switch name {
	case "split", "chars", "trim", "toLower", "toUpper", "contains", "startsWith", "endsWith", "replace":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, str)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown string member: " + name}
	}
}

func (e *Evaluator) mapMethod(m *Map, name string) (Value, *Signal, error) {
	switch name {
	case "get", "set", "has", "delete", "keys", "values":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, m)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown map member: " + name}
	}
}

func (e *Evaluator) setMethod(s *Set, name string) (Value, *Signal, error) {
	switch name {
	case "add", "has", "delete", "values":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, s)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown set member: " + name}
	}
}

func (e *Evaluator) taskMethod(t *Task, name string) (Value, *Signal, error) {
	switch name {
	case "then":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, t)}, nil, nil
	case "cancel":
		return &Builtin{
			Name: "cancel",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "cancel expects no arguments"}
				}
				t.Cancel()
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown task member: " + name}
	}
}
