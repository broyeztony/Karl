package interpreter

import (
	"io"
	"strings"
)

func registerStreamPipelineBuiltins() {
	builtins["read"] = &Builtin{Name: "read", Fn: builtinReadSource}
	builtins["stdin"] = &Builtin{Name: "stdin", Fn: builtinStdinSource}
	builtins["fromChannel"] = &Builtin{Name: "fromChannel", Fn: builtinFromChannelSource}
	builtins["join"] = &Builtin{Name: "join", Fn: builtinJoinSource}
	builtins["merge"] = &Builtin{Name: "merge", Fn: builtinMergeSource}
	builtins["zip"] = &Builtin{Name: "zip", Fn: builtinZipSource}
	builtins["debounce"] = &Builtin{Name: "debounce", Fn: builtinDebounceStage}
	builtins["lines"] = &Builtin{Name: "lines", Fn: builtinLinesStage}
	builtins["tee"] = &Builtin{Name: "tee", Fn: builtinTeeStage}
	builtins["spill"] = &Builtin{Name: "spill", Fn: builtinSpillStage}
	builtins["stdout"] = &Builtin{Name: "stdout", Fn: builtinStdoutSink}
	builtins["write"] = &Builtin{Name: "write", Fn: builtinWriteSink}
	builtins["collect"] = &Builtin{Name: "collect", Fn: builtinCollectSink}
	builtins["toChannel"] = &Builtin{Name: "toChannel", Fn: builtinToChannelSink}
	builtins["exec"] = &Builtin{Name: "exec", Fn: builtinExecSink}
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

func builtinStdinSource(_ *Evaluator, args []Value) (Value, error) {
	if len(args) > 1 {
		return nil, &RuntimeError{Message: "stdin expects optional opts object"}
	}
	mode := streamTypeBytes
	if len(args) == 1 && !Equivalent(args[0], NullValue) {
		opts, ok := objectPairs(args[0])
		if !ok {
			return nil, &RuntimeError{Message: "stdin opts must be object"}
		}
		if typeVal, ok := opts["type"]; ok && !Equivalent(typeVal, NullValue) {
			parsed, err := parseStreamType(typeVal, "type")
			if err != nil {
				return nil, err
			}
			mode = parsed
		}
	}
	return &StreamSourceValue{
		name: "stdin",
		open: func(e *Evaluator) (streamIterator, error) {
			if e == nil || e.runtime == nil {
				return nil, recoverableError("stream_open", "stdin unavailable")
			}
			reader := &StreamReader{
				reader: &runtimeInputLockedReader{runtime: e.runtime},
				mode:   mode,
			}
			return &streamReaderIterator{
				reader:      reader,
				size:        defaultStreamReadSize,
				closeOnExit: false,
			}, nil
		},
	}, nil
}

func builtinFromChannelSource(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "fromChannel expects channel"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "fromChannel expects channel"}
	}
	return &StreamSourceValue{
		name: "fromChannel",
		open: func(_ *Evaluator) (streamIterator, error) {
			return &channelStreamIterator{ch: ch}, nil
		},
	}, nil
}

func builtinMergeSource(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, &RuntimeError{Message: "merge expects at least one stream source or plan"}
	}
	streamArgs := append([]Value(nil), args...)
	return &StreamSourceValue{
		name: "merge",
		open: func(e *Evaluator) (streamIterator, error) {
			iters := make([]streamIterator, 0, len(streamArgs))
			for _, arg := range streamArgs {
				iter, err := openStreamIterator(e, arg)
				if err != nil {
					for _, it := range iters {
						_ = it.Close()
					}
					return nil, err
				}
				iters = append(iters, iter)
			}
			open := make([]bool, len(iters))
			for i := range open {
				open[i] = true
			}
			return &mergeStreamIterator{
				iters:     iters,
				open:      open,
				remaining: len(iters),
			}, nil
		},
	}, nil
}

func builtinZipSource(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "zip expects exactly 2 stream sources or plans"}
	}
	leftArg := args[0]
	rightArg := args[1]
	return &StreamSourceValue{
		name: "zip",
		open: func(e *Evaluator) (streamIterator, error) {
			left, err := openStreamIterator(e, leftArg)
			if err != nil {
				return nil, err
			}
			right, err := openStreamIterator(e, rightArg)
			if err != nil {
				_ = left.Close()
				return nil, err
			}
			return &zipStreamIterator{
				left:  left,
				right: right,
			}, nil
		},
	}, nil
}

func builtinJoinSource(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 4 {
		return nil, &RuntimeError{Message: "join expects (left, right, leftKeyFn, rightKeyFn)"}
	}
	leftArg := args[0]
	rightArg := args[1]
	leftKeyFn := args[2]
	rightKeyFn := args[3]
	if !isCallable(leftKeyFn) || !isCallable(rightKeyFn) {
		return nil, &RuntimeError{Message: "join key extractors must be functions"}
	}

	return &StreamSourceValue{
		name: "join",
		open: func(_ *Evaluator) (streamIterator, error) {
			leftIter, err := openStreamIterator(e, leftArg)
			if err != nil {
				return nil, err
			}
			rightIter, err := openStreamIterator(e, rightArg)
			if err != nil {
				_ = leftIter.Close()
				return nil, err
			}

			rightIndex := map[MapKey][]Value{}
			for {
				item, eof, err := rightIter.Next()
				if err != nil {
					_ = rightIter.Close()
					_ = leftIter.Close()
					return nil, err
				}
				if eof {
					break
				}
				keyVal, sig, err := e.applyFunction(rightKeyFn, []Value{item})
				if err != nil {
					_ = rightIter.Close()
					_ = leftIter.Close()
					return nil, err
				}
				if sig != nil {
					_ = rightIter.Close()
					_ = leftIter.Close()
					return nil, &RuntimeError{Message: "break/continue outside loop"}
				}
				key, err := mapKeyForValue(keyVal)
				if err != nil {
					_ = rightIter.Close()
					_ = leftIter.Close()
					return nil, &RuntimeError{Message: "join key must be hashable"}
				}
				rightIndex[key] = append(rightIndex[key], cloneStreamItem(item))
			}
			_ = rightIter.Close()

			return &joinStreamIterator{
				left:      leftIter,
				eval:      e,
				leftKeyFn: leftKeyFn,
				right:     rightIndex,
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

func builtinDebounceStage(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "debounce expects 1 argument"}
	}
	ms, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "debounce expects integer milliseconds"}
	}
	if ms.Value < 0 {
		return nil, &RuntimeError{Message: "debounce expects non-negative milliseconds"}
	}
	return &StreamStageValue{
		name: "debounce",
		apply: func(upstream streamIterator) streamIterator {
			return newStreamDebounceIterator(upstream, e, ms.Value)
		},
	}, nil
}

func builtinTeeStage(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "tee expects stream sink"}
	}
	side, ok := args[0].(*StreamSinkValue)
	if !ok {
		return nil, &RuntimeError{Message: "tee expects stream sink"}
	}
	return &StreamStageValue{
		name: "tee",
		apply: func(upstream streamIterator) streamIterator {
			return newTeeStreamIterator(e, upstream, side)
		},
	}, nil
}

func builtinSpillStage(e *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "spill expects (path, opts?)"}
	}
	writeSink, err := builtinWriteSink(e, args)
	if err != nil {
		return nil, err
	}
	side, ok := writeSink.(*StreamSinkValue)
	if !ok {
		return nil, &RuntimeError{Message: "spill expects writable stream sink"}
	}
	return &StreamStageValue{
		name: "spill",
		apply: func(upstream streamIterator) streamIterator {
			return newTeeStreamIterator(e, upstream, side)
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
					"type": &String{Value: streamTypeBytes},
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
					textChunk, err := streamValueToTextChunk(item)
					if err != nil {
						return nil, recoverableError("stream_write", "stream write error: "+err.Error())
					}
					text := textChunk
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
				switch v := item.(type) {
				case *Bytes:
					out = append(out, &Bytes{Value: append([]byte{}, v.Value...)})
				case *String:
					out = append(out, &String{Value: v.Value})
				default:
					out = append(out, item)
				}
			}
		},
	}, nil
}

func builtinToChannelSink(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "toChannel expects channel"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "toChannel expects channel"}
	}
	return &StreamSinkValue{
		name: "toChannel",
		run: func(e *Evaluator, upstream streamIterator) (Value, error) {
			defer ch.Close()
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return UnitValue, nil
				}
				if err := channelSendBlocking(e, ch, item); err != nil {
					return nil, err
				}
			}
		},
	}, nil
}

func builtinExecSink(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, &RuntimeError{Message: "exec expects process spec or command arguments"}
	}
	sinkArgs := append([]Value(nil), args...)
	return &StreamSinkValue{
		name: "exec",
		run: func(e *Evaluator, upstream streamIterator) (Value, error) {
			var stdinBuilder strings.Builder
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					break
				}
				text, err := streamValueToTextChunk(item)
				if err != nil {
					text = streamValueToLine(item)
				}
				stdinBuilder.WriteString(text)
			}

			spec, err := parseRunSpec(sinkArgs)
			if err != nil {
				return nil, err
			}
			stdin := stdinBuilder.String()
			spec.stdinText = &stdin
			spec.stdinMode = processModePipe
			return executeRunSpec(e, spec)
		},
	}, nil
}

type runtimeInputLockedReader struct {
	runtime *runtimeState
}

func (r *runtimeInputLockedReader) Read(p []byte) (int, error) {
	if r == nil || r.runtime == nil {
		return 0, io.EOF
	}
	r.runtime.inputMu.Lock()
	defer r.runtime.inputMu.Unlock()
	reader, err := r.runtime.inputBufReader()
	if err != nil {
		return 0, err
	}
	return reader.Read(p)
}

type channelStreamIterator struct {
	ch     *Channel
	closed bool
}

func (c *channelStreamIterator) Next() (Value, bool, error) {
	if c == nil || c.ch == nil || c.closed {
		return nil, true, nil
	}
	for {
		select {
		case val := <-c.ch.Ch:
			return val, false, nil
		default:
		}
		select {
		case val := <-c.ch.Ch:
			return val, false, nil
		case <-c.ch.ClosedSignal():
			select {
			case val := <-c.ch.Ch:
				return val, false, nil
			default:
				return nil, true, nil
			}
		}
	}
}

func (c *channelStreamIterator) Close() error {
	if c == nil || c.closed {
		return nil
	}
	c.closed = true
	return nil
}

type mergeStreamIterator struct {
	iters     []streamIterator
	open      []bool
	remaining int
	next      int
	closed    bool
}

func (m *mergeStreamIterator) Next() (Value, bool, error) {
	if m == nil || m.closed || m.remaining == 0 || len(m.iters) == 0 {
		return nil, true, nil
	}

	checked := 0
	for checked < len(m.iters) {
		idx := m.next % len(m.iters)
		m.next = (idx + 1) % len(m.iters)
		checked++
		if !m.open[idx] {
			continue
		}

		item, eof, err := m.iters[idx].Next()
		if err != nil {
			return nil, false, err
		}
		if eof {
			m.open[idx] = false
			m.remaining--
			continue
		}
		return item, false, nil
	}
	if m.remaining <= 0 {
		return nil, true, nil
	}
	return nil, true, nil
}

func (m *mergeStreamIterator) Close() error {
	if m == nil || m.closed {
		return nil
	}
	m.closed = true
	var firstErr error
	for _, it := range m.iters {
		if it == nil {
			continue
		}
		if err := it.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type zipStreamIterator struct {
	left   streamIterator
	right  streamIterator
	closed bool
}

func (z *zipStreamIterator) Next() (Value, bool, error) {
	if z == nil || z.closed || z.left == nil || z.right == nil {
		return nil, true, nil
	}
	leftItem, leftEOF, err := z.left.Next()
	if err != nil {
		return nil, false, err
	}
	if leftEOF {
		return nil, true, nil
	}
	rightItem, rightEOF, err := z.right.Next()
	if err != nil {
		return nil, false, err
	}
	if rightEOF {
		return nil, true, nil
	}
	return &Array{Elements: []Value{cloneStreamItem(leftItem), cloneStreamItem(rightItem)}}, false, nil
}

func (z *zipStreamIterator) Close() error {
	if z == nil || z.closed {
		return nil
	}
	z.closed = true
	var firstErr error
	if z.left != nil {
		if err := z.left.Close(); err != nil {
			firstErr = err
		}
	}
	if z.right != nil {
		if err := z.right.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type joinStreamIterator struct {
	left      streamIterator
	eval      *Evaluator
	leftKeyFn Value
	right     map[MapKey][]Value
	pending   []Value
	closed    bool
}

func (j *joinStreamIterator) Next() (Value, bool, error) {
	if j == nil || j.closed || j.left == nil {
		return nil, true, nil
	}
	if len(j.pending) > 0 {
		out := j.pending[0]
		j.pending = j.pending[1:]
		return out, false, nil
	}

	for {
		leftItem, eof, err := j.left.Next()
		if err != nil {
			return nil, false, err
		}
		if eof {
			return nil, true, nil
		}
		keyVal, sig, err := j.eval.applyFunction(j.leftKeyFn, []Value{leftItem})
		if err != nil {
			return nil, false, err
		}
		if sig != nil {
			return nil, false, &RuntimeError{Message: "break/continue outside loop"}
		}
		key, err := mapKeyForValue(keyVal)
		if err != nil {
			return nil, false, &RuntimeError{Message: "join key must be hashable"}
		}
		matches := j.right[key]
		if len(matches) == 0 {
			continue
		}
		leftCopy := cloneStreamItem(leftItem)
		j.pending = make([]Value, 0, len(matches))
		for _, rightItem := range matches {
			j.pending = append(j.pending, &Array{
				Elements: []Value{cloneStreamItem(leftCopy), cloneStreamItem(rightItem)},
			})
		}
		out := j.pending[0]
		j.pending = j.pending[1:]
		return out, false, nil
	}
}

func (j *joinStreamIterator) Close() error {
	if j == nil || j.closed {
		return nil
	}
	j.closed = true
	if j.left != nil {
		return j.left.Close()
	}
	return nil
}
