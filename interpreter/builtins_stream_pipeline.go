package interpreter

import "strings"

func registerStreamPipelineBuiltins() {
	builtins["read"] = &Builtin{Name: "read", Fn: builtinReadSource}
	builtins["lines"] = &Builtin{Name: "lines", Fn: builtinLinesStage}
	builtins["stdout"] = &Builtin{Name: "stdout", Fn: builtinStdoutSink}
	builtins["write"] = &Builtin{Name: "write", Fn: builtinWriteSink}
	builtins["collect"] = &Builtin{Name: "collect", Fn: builtinCollectSink}
}

func builtinReadSource(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "read expects (path, opts?)"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "read path must be string"}
	}
	var opts Value = NullValue
	if len(args) == 2 {
		opts = args[1]
		if !Equivalent(opts, NullValue) {
			if _, ok := objectPairs(opts); !ok {
				return nil, &RuntimeError{Message: "read opts must be object"}
			}
		}
	}

	return &StreamSourceValue{
		name: "read",
		open: func(e *Evaluator) (streamIterator, error) {
			readerArgs := []Value{&String{Value: path}}
			if !Equivalent(opts, NullValue) {
				readerArgs = append(readerArgs, opts)
			}
			v, err := builtinReader(e, readerArgs)
			if err != nil {
				return nil, err
			}
			reader, ok := v.(*StreamReader)
			if !ok {
				return nil, &RuntimeError{Message: "read source expects stream reader"}
			}
			return &streamReaderIterator{
				reader:      reader,
				size:        defaultStreamReadSize,
				closeOnExit: true,
			}, nil
		},
	}, nil
}

func builtinLinesStage(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "lines expects no arguments"}
	}
	return &StreamStageValue{
		name: "lines",
		apply: func(upstream streamIterator) streamIterator {
			return &linesIterator{upstream: upstream}
		},
	}, nil
}

func builtinStdoutSink(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "stdout expects no arguments"}
	}
	return &StreamSinkValue{
		name: "stdout",
		run: func(e *Evaluator, upstream streamIterator) (Value, error) {
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return UnitValue, nil
				}
				if err := writeLogLine(e, streamValueToLine(item)); err != nil {
					return nil, &RuntimeError{Message: "stdout write failed: " + err.Error()}
				}
			}
		},
	}, nil
}

func builtinWriteSink(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "write expects (path, opts?)"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "write path must be string"}
	}

	var opts Value = NullValue
	if len(args) == 2 {
		opts = args[1]
		if !Equivalent(opts, NullValue) {
			if _, ok := objectPairs(opts); !ok {
				return nil, &RuntimeError{Message: "write opts must be object"}
			}
		}
	}

	return &StreamSinkValue{
		name: "write",
		run: func(e *Evaluator, upstream streamIterator) (Value, error) {
			writerArgs := []Value{&String{Value: path}}
			if !Equivalent(opts, NullValue) {
				writerArgs = append(writerArgs, opts)
			} else {
				writerArgs = append(writerArgs, &Object{Pairs: map[string]Value{
					"type": &String{Value: streamTypeText},
				}})
			}

			v, err := builtinWriter(e, writerArgs)
			if err != nil {
				return nil, err
			}
			writer, ok := v.(*StreamWriter)
			if !ok {
				return nil, &RuntimeError{Message: "write sink expects stream writer"}
			}
			defer func() {
				_ = writer.Close()
			}()

			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					if err := writer.Close(); err != nil {
						return nil, recoverableError("stream_close", "stream close error: "+err.Error())
					}
					return UnitValue, nil
				}

				var payload []byte
				if writer.Mode() == streamTypeBytes {
					payload = streamValueToBytes(item)
				} else {
					text := streamValueToLine(item)
					if !strings.HasSuffix(text, "\n") {
						text += "\n"
					}
					payload = []byte(text)
				}
				if _, err := writer.WriteChunk(payload); err != nil {
					return nil, recoverableError("stream_write", "stream write error: "+err.Error())
				}
			}
		},
	}, nil
}

func builtinCollectSink(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "collect expects no arguments"}
	}
	return &StreamSinkValue{
		name: "collect",
		run: func(_ *Evaluator, upstream streamIterator) (Value, error) {
			out := make([]Value, 0, 32)
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return &Array{Elements: out}, nil
				}
				out = append(out, item)
			}
		},
	}, nil
}
