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

func (e *Evaluator) processMethod(p *Process, name string) (Value, *Signal, error) {
	switch name {
	case "pid":
		return &Integer{Value: p.PID()}, nil, nil
	case "running":
		return &Boolean{Value: p.Running()}, nil, nil
	case "stdin":
		ch, ok := p.inputChannel()
		if !ok {
			return nil, nil, recoverableError("process_state", "stdin is only available when stdIn mode is \"pipe\"")
		}
		return ch, nil, nil
	case "stdout":
		ch, ok := p.outputChannel()
		if !ok {
			return nil, nil, recoverableError("process_state", "stdout is only available when stdOut mode is \"pipe\"")
		}
		return ch, nil, nil
	case "stderr":
		ch, ok := p.errorChannel()
		if !ok {
			return nil, nil, recoverableError("process_state", "stderr is only available when stdErr mode is \"pipe\"")
		}
		return ch, nil, nil
	case "abort":
		return &Builtin{
			Name: "abort",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "abort expects no arguments"}
				}
				if err := p.Abort(); err != nil {
					return nil, recoverableError("process_state", "process abort error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	case "kill":
		return &Builtin{
			Name: "kill",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "kill expects no arguments"}
				}
				if err := p.Kill(); err != nil {
					return nil, recoverableError("process_state", "process kill error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	case "signal":
		return &Builtin{
			Name: "signal",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "signal expects signal name"}
				}
				name, ok := stringArg(args[0])
				if !ok {
					return nil, &RuntimeError{Message: "signal expects string signal name"}
				}
				if err := p.Signal(name); err != nil {
					return nil, recoverableError("process_state", "process signal error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown process member: " + name}
	}
}
