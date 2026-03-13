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
		stream, ok := p.inputStream()
		if !ok {
			return nil, nil, recoverableError("process_state", "stdin is only available when stdin mode is \"pipe\"")
		}
		return stream, nil, nil
	case "stdout":
		stream, ok := p.outputStream()
		if !ok {
			return nil, nil, recoverableError("process_state", "stdout is only available when stdout mode is \"pipe\"")
		}
		return stream, nil, nil
	case "stderr":
		stream, ok := p.errorStream()
		if !ok {
			return nil, nil, recoverableError("process_state", "stderr is only available when stderr mode is \"pipe\"")
		}
		return stream, nil, nil
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

func (e *Evaluator) streamReaderMethod(s *StreamReader, name string) (Value, *Signal, error) {
	switch name {
	case "read":
		return &Builtin{
			Name: "read",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) > 1 {
					return nil, &RuntimeError{Message: "read expects optional size"}
				}
				size := int64(defaultStreamReadSize)
				if len(args) == 1 {
					i, ok := args[0].(*Integer)
					if !ok {
						return nil, &RuntimeError{Message: "read size must be integer"}
					}
					if i.Value <= 0 {
						return nil, &RuntimeError{Message: "read size must be > 0"}
					}
					size = i.Value
				}

				chunk, eof, err := s.ReadChunk(int(size))
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return &Array{Elements: []Value{NullValue, &Boolean{Value: true}}}, nil
				}
				if s.Mode() == streamTypeBytes {
					copied := append([]byte{}, chunk...)
					return &Array{Elements: []Value{&Bytes{Value: copied}, &Boolean{Value: false}}}, nil
				}
				return &Array{Elements: []Value{&String{Value: string(chunk)}, &Boolean{Value: false}}}, nil
			},
		}, nil, nil
	case "close":
		return &Builtin{
			Name: "close",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "close expects no arguments"}
				}
				if err := s.Close(); err != nil {
					return nil, recoverableError("stream_close", "stream close error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown stream reader member: " + name}
	}
}

func (e *Evaluator) streamWriterMethod(s *StreamWriter, name string) (Value, *Signal, error) {
	switch name {
	case "write":
		return &Builtin{
			Name: "write",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "write expects 1 argument"}
				}
				mode := s.Mode()
				var payload []byte
				if mode == streamTypeBytes {
					bytesPayload, ok := bytesArg(args[0])
					if !ok {
						return nil, &RuntimeError{Message: "write expects bytes payload in BYTES mode"}
					}
					payload = bytesPayload
				} else {
					textPayload, ok := stringArg(args[0])
					if !ok {
						return nil, &RuntimeError{Message: "write expects string payload in TEXT mode"}
					}
					payload = []byte(textPayload)
				}
				n, err := s.WriteChunk(payload)
				if err != nil {
					return nil, recoverableError("stream_write", "stream write error: "+err.Error())
				}
				return &Integer{Value: int64(n)}, nil
			},
		}, nil, nil
	case "close":
		return &Builtin{
			Name: "close",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "close expects no arguments"}
				}
				if err := s.Close(); err != nil {
					return nil, recoverableError("stream_close", "stream close error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown stream writer member: " + name}
	}
}

func (e *Evaluator) streamValueMethod(value Value, name string) (Value, *Signal, error) {
	switch name {
	case "read":
		return &Builtin{
			Name: "read",
			Fn: func(eval *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "read expects no arguments"}
				}
				return streamReadValue(eval, value)
			},
		}, nil, nil
	case "close":
		return &Builtin{
			Name: "close",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "close expects no arguments"}
				}
				if err := streamCloseValue(value); err != nil {
					if _, ok := err.(*RecoverableError); ok {
						return nil, err
					}
					return nil, recoverableError("stream_close", "stream close error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown stream member: " + name}
	}
}
