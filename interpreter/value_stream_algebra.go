package interpreter

import "sort"

func newStreamMapStage(e *Evaluator, fn Value) *StreamStageValue {
	return &StreamStageValue{
		name: "map",
		apply: func(upstream streamIterator) streamIterator {
			return &streamMapIterator{
				upstream: upstream,
				eval:     e,
				fn:       fn,
			}
		},
	}
}

func newStreamFilterStage(e *Evaluator, fn Value) *StreamStageValue {
	return &StreamStageValue{
		name: "filter",
		apply: func(upstream streamIterator) streamIterator {
			return &streamFilterIterator{
				upstream: upstream,
				eval:     e,
				fn:       fn,
			}
		},
	}
}

func newStreamFlatMapStage(e *Evaluator, fn Value) *StreamStageValue {
	return &StreamStageValue{
		name: "flatMap",
		apply: func(upstream streamIterator) streamIterator {
			return &streamFlatMapIterator{
				upstream: upstream,
				eval:     e,
				fn:       fn,
				pending:  nil,
			}
		},
	}
}

func newStreamTakeStage(n int64) *StreamStageValue {
	return &StreamStageValue{
		name: "take",
		apply: func(upstream streamIterator) streamIterator {
			return &streamTakeIterator{
				upstream:  upstream,
				remaining: n,
			}
		},
	}
}

func newStreamDropStage(n int64) *StreamStageValue {
	return &StreamStageValue{
		name: "drop",
		apply: func(upstream streamIterator) streamIterator {
			return &streamDropIterator{
				upstream:  upstream,
				remaining: n,
			}
		},
	}
}

func newStreamChunkStage(size int64) *StreamStageValue {
	return &StreamStageValue{
		name: "chunk",
		apply: func(upstream streamIterator) streamIterator {
			return &streamChunkIterator{
				upstream: upstream,
				size:     int(size),
			}
		},
	}
}

func newStreamWindowStage(size int64, step int64) *StreamStageValue {
	return &StreamStageValue{
		name: "window",
		apply: func(upstream streamIterator) streamIterator {
			return &streamWindowIterator{
				upstream: upstream,
				size:     int(size),
				step:     int(step),
				buffer:   nil,
				primed:   false,
				drained:  false,
			}
		},
	}
}

func newStreamThrottleStage(e *Evaluator, ms int64) *StreamStageValue {
	return &StreamStageValue{
		name: "throttle",
		apply: func(upstream streamIterator) streamIterator {
			return &streamThrottleIterator{
				upstream: upstream,
				eval:     e,
				ms:       ms,
			}
		},
	}
}

func newStreamFromJSONStage() *StreamStageValue {
	return &StreamStageValue{
		name: "fromJson",
		apply: func(upstream streamIterator) streamIterator {
			return &streamFromJSONIterator{upstream: upstream}
		},
	}
}

func newStreamToJSONStage() *StreamStageValue {
	return &StreamStageValue{
		name: "toJson",
		apply: func(upstream streamIterator) streamIterator {
			return &streamToJSONIterator{upstream: upstream}
		},
	}
}

func newStreamFromUTF8Stage() *StreamStageValue {
	return &StreamStageValue{
		name: "fromUtf8",
		apply: func(upstream streamIterator) streamIterator {
			return &streamFromUTF8Iterator{upstream: upstream}
		},
	}
}

func newStreamToUTF8Stage() *StreamStageValue {
	return &StreamStageValue{
		name: "toUtf8",
		apply: func(upstream streamIterator) streamIterator {
			return &streamToUTF8Iterator{upstream: upstream}
		},
	}
}

func newStreamDistinctStage() *StreamStageValue {
	return &StreamStageValue{
		name: "distinct",
		apply: func(upstream streamIterator) streamIterator {
			return &streamDistinctIterator{
				upstream: upstream,
				seen:     make(map[MapKey]struct{}),
			}
		},
	}
}

func newStreamSortStage(e *Evaluator, cmp Value) *StreamStageValue {
	return &StreamStageValue{
		name: "sort",
		apply: func(upstream streamIterator) streamIterator {
			return &streamSortIterator{
				upstream: upstream,
				eval:     e,
				cmp:      cmp,
			}
		},
	}
}

func newStreamReduceSink(e *Evaluator, initial Value, fn Value) *StreamSinkValue {
	return &StreamSinkValue{
		name: "reduce",
		run: func(_ *Evaluator, upstream streamIterator) (Value, error) {
			acc := initial
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return acc, nil
				}
				next, sig, err := e.applyFunction(fn, []Value{acc, item})
				if err != nil {
					return nil, err
				}
				if sig != nil {
					return nil, &RuntimeError{Message: "break/continue outside loop"}
				}
				acc = next
			}
		},
	}
}

func newStreamCountSink() *StreamSinkValue {
	return &StreamSinkValue{
		name: "count",
		run: func(_ *Evaluator, upstream streamIterator) (Value, error) {
			var count int64
			for {
				_, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return &Integer{Value: count}, nil
				}
				count++
			}
		},
	}
}

func newStreamForEachSink(e *Evaluator, fn Value) *StreamSinkValue {
	return &StreamSinkValue{
		name: "forEach",
		run: func(_ *Evaluator, upstream streamIterator) (Value, error) {
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return UnitValue, nil
				}
				_, sig, err := e.applyFunction(fn, []Value{item})
				if err != nil {
					return nil, err
				}
				if sig != nil {
					return nil, &RuntimeError{Message: "break/continue outside loop"}
				}
			}
		},
	}
}

func newStreamGroupCountSink(e *Evaluator, keyFn Value) *StreamSinkValue {
	return &StreamSinkValue{
		name: "group_count",
		run: func(_ *Evaluator, upstream streamIterator) (Value, error) {
			counts := make(map[MapKey]int64)
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					out := &Map{Pairs: make(map[MapKey]Value, len(counts))}
					for k, c := range counts {
						out.Pairs[k] = &Integer{Value: c}
					}
					return out, nil
				}
				keyVal := item
				if keyFn != nil {
					next, sig, err := e.applyFunction(keyFn, []Value{item})
					if err != nil {
						return nil, err
					}
					if sig != nil {
						return nil, &RuntimeError{Message: "break/continue outside loop"}
					}
					keyVal = next
				}
				key, err := mapKeyForValue(keyVal)
				if err != nil {
					return nil, &RuntimeError{Message: "group_count key must be hashable"}
				}
				counts[key]++
			}
		},
	}
}

func newStreamReduceByKeySink(e *Evaluator, keyFn Value, initial Value, reducer Value) *StreamSinkValue {
	return &StreamSinkValue{
		name: "reduce_by_key",
		run: func(_ *Evaluator, upstream streamIterator) (Value, error) {
			state := make(map[MapKey]Value)
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return &Map{Pairs: state}, nil
				}

				keyVal, sig, err := e.applyFunction(keyFn, []Value{item})
				if err != nil {
					return nil, err
				}
				if sig != nil {
					return nil, &RuntimeError{Message: "break/continue outside loop"}
				}
				key, err := mapKeyForValue(keyVal)
				if err != nil {
					return nil, &RuntimeError{Message: "reduce_by_key key must be hashable"}
				}

				acc, ok := state[key]
				if !ok {
					acc = cloneAggregateValue(initial)
				}

				next, sig, err := e.applyFunction(reducer, []Value{acc, item})
				if err != nil {
					return nil, err
				}
				if sig != nil {
					return nil, &RuntimeError{Message: "break/continue outside loop"}
				}
				state[key] = next
			}
		},
	}
}

func newStreamTopSink(e *Evaluator, n int64, scoreFn Value) *StreamSinkValue {
	return &StreamSinkValue{
		name: "top",
		run: func(_ *Evaluator, upstream streamIterator) (Value, error) {
			if n <= 0 {
				return &Array{Elements: []Value{}}, nil
			}
			entries := make([]streamTopEntry, 0, 64)
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					break
				}
				score, err := streamTopScore(e, item, scoreFn)
				if err != nil {
					return nil, err
				}
				entries = append(entries, streamTopEntry{item: cloneStreamItem(item), score: score})
			}

			sort.SliceStable(entries, func(i, j int) bool {
				return entries[i].score > entries[j].score
			})

			limit := int(n)
			if limit > len(entries) {
				limit = len(entries)
			}
			out := make([]Value, 0, limit)
			for i := 0; i < limit; i++ {
				out = append(out, entries[i].item)
			}
			return &Array{Elements: out}, nil
		},
	}
}

func newStreamSplitSink(e *Evaluator, fn Value) *StreamSinkValue {
	return &StreamSinkValue{
		name: "split",
		run: func(_ *Evaluator, upstream streamIterator) (Value, error) {
			left := make([]Value, 0, 32)
			right := make([]Value, 0, 32)
			for {
				item, eof, err := upstream.Next()
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return &Object{Pairs: map[string]Value{
						"left":  &Array{Elements: left},
						"right": &Array{Elements: right},
					}}, nil
				}
				matches, err := applyListPredicate(e, fn, item, "split")
				if err != nil {
					return nil, err
				}
				if matches {
					left = append(left, cloneStreamItem(item))
				} else {
					right = append(right, cloneStreamItem(item))
				}
			}
		},
	}
}

type streamMapIterator struct {
	upstream streamIterator
	eval     *Evaluator
	fn       Value
	closed   bool
}

func (s *streamMapIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	item, eof, err := s.upstream.Next()
	if err != nil || eof {
		return nil, eof, err
	}
	out, sig, err := s.eval.applyFunction(s.fn, []Value{item})
	if err != nil {
		return nil, false, err
	}
	if sig != nil {
		return nil, false, &RuntimeError{Message: "break/continue outside loop"}
	}
	return out, false, nil
}

func (s *streamMapIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamFilterIterator struct {
	upstream streamIterator
	eval     *Evaluator
	fn       Value
	closed   bool
}

func (s *streamFilterIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	for {
		item, eof, err := s.upstream.Next()
		if err != nil || eof {
			return nil, eof, err
		}
		matches, err := applyListPredicate(s.eval, s.fn, item, "filter")
		if err != nil {
			return nil, false, err
		}
		if matches {
			return item, false, nil
		}
	}
}

func (s *streamFilterIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamFlatMapIterator struct {
	upstream streamIterator
	eval     *Evaluator
	fn       Value
	pending  []Value
	closed   bool
}

func (s *streamFlatMapIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	for {
		if len(s.pending) > 0 {
			next := s.pending[0]
			s.pending = s.pending[1:]
			return next, false, nil
		}

		item, eof, err := s.upstream.Next()
		if err != nil || eof {
			return nil, eof, err
		}
		mapped, sig, err := s.eval.applyFunction(s.fn, []Value{item})
		if err != nil {
			return nil, false, err
		}
		if sig != nil {
			return nil, false, &RuntimeError{Message: "break/continue outside loop"}
		}
		arr, ok := mapped.(*Array)
		if !ok {
			return nil, false, &RuntimeError{Message: "flatMap mapper must return array"}
		}
		if len(arr.Elements) == 0 {
			continue
		}
		s.pending = append([]Value{}, arr.Elements...)
	}
}

func (s *streamFlatMapIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamTakeIterator struct {
	upstream  streamIterator
	remaining int64
	done      bool
	closed    bool
}

func (s *streamTakeIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil || s.done || s.remaining <= 0 {
		s.done = true
		return nil, true, nil
	}
	item, eof, err := s.upstream.Next()
	if err != nil {
		return nil, false, err
	}
	if eof {
		s.done = true
		return nil, true, nil
	}
	s.remaining--
	if s.remaining <= 0 {
		s.done = true
	}
	return item, false, nil
}

func (s *streamTakeIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamDropIterator struct {
	upstream  streamIterator
	remaining int64
	closed    bool
}

func (s *streamDropIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	for s.remaining > 0 {
		_, eof, err := s.upstream.Next()
		if err != nil {
			return nil, false, err
		}
		if eof {
			return nil, true, nil
		}
		s.remaining--
	}
	return s.upstream.Next()
}

func (s *streamDropIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamChunkIterator struct {
	upstream streamIterator
	size     int
	drained  bool
	closed   bool
}

func (s *streamChunkIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil || s.drained {
		return nil, true, nil
	}
	chunk := make([]Value, 0, s.size)
	for len(chunk) < s.size {
		item, eof, err := s.upstream.Next()
		if err != nil {
			return nil, false, err
		}
		if eof {
			s.drained = true
			if len(chunk) == 0 {
				return nil, true, nil
			}
			return &Array{Elements: chunk}, false, nil
		}
		chunk = append(chunk, cloneStreamItem(item))
	}
	return &Array{Elements: chunk}, false, nil
}

func (s *streamChunkIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamWindowIterator struct {
	upstream streamIterator
	size     int
	step     int
	buffer   []Value
	primed   bool
	drained  bool
	closed   bool
}

func (s *streamWindowIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil || s.drained {
		return nil, true, nil
	}
	if !s.primed {
		for len(s.buffer) < s.size {
			item, eof, err := s.upstream.Next()
			if err != nil {
				return nil, false, err
			}
			if eof {
				s.drained = true
				return nil, true, nil
			}
			s.buffer = append(s.buffer, cloneStreamItem(item))
		}
		s.primed = true
		out := make([]Value, len(s.buffer))
		copy(out, s.buffer)
		return &Array{Elements: out}, false, nil
	}

	dropN := s.step
	if dropN > len(s.buffer) {
		dropN = len(s.buffer)
	}
	s.buffer = s.buffer[dropN:]

	for len(s.buffer) < s.size {
		item, eof, err := s.upstream.Next()
		if err != nil {
			return nil, false, err
		}
		if eof {
			s.drained = true
			return nil, true, nil
		}
		s.buffer = append(s.buffer, cloneStreamItem(item))
	}
	out := make([]Value, len(s.buffer))
	copy(out, s.buffer)
	return &Array{Elements: out}, false, nil
}

func (s *streamWindowIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

func cloneStreamItem(v Value) Value {
	switch it := v.(type) {
	case *Bytes:
		return &Bytes{Value: append([]byte{}, it.Value...)}
	case *String:
		return &String{Value: it.Value}
	case *Char:
		return &Char{Value: it.Value}
	default:
		return v
	}
}

type streamThrottleIterator struct {
	upstream streamIterator
	eval     *Evaluator
	ms       int64
	started  bool
	closed   bool
}

func (s *streamThrottleIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	if s.started && s.ms > 0 {
		if _, err := builtinSleep(s.eval, []Value{&Integer{Value: s.ms}}); err != nil {
			return nil, false, err
		}
	}
	item, eof, err := s.upstream.Next()
	if err != nil || eof {
		return nil, eof, err
	}
	s.started = true
	return item, false, nil
}

func (s *streamThrottleIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamFromJSONIterator struct {
	upstream streamIterator
	closed   bool
}

func (s *streamFromJSONIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	item, eof, err := s.upstream.Next()
	if err != nil || eof {
		return nil, eof, err
	}
	text, err := streamValueToTextChunk(item)
	if err != nil {
		return nil, false, err
	}
	decoded, err := builtinFromJSON(nil, []Value{&String{Value: text}})
	if err != nil {
		return nil, false, err
	}
	return decoded, false, nil
}

func (s *streamFromJSONIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamToJSONIterator struct {
	upstream streamIterator
	closed   bool
}

func (s *streamToJSONIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	item, eof, err := s.upstream.Next()
	if err != nil || eof {
		return nil, eof, err
	}
	encoded, err := builtinToJSON(nil, []Value{item})
	if err != nil {
		return nil, false, err
	}
	text, ok := encoded.(*String)
	if !ok {
		return nil, false, &RuntimeError{Message: "toJson stage produced non-string value"}
	}
	return &String{Value: text.Value}, false, nil
}

func (s *streamToJSONIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamFromUTF8Iterator struct {
	upstream streamIterator
	closed   bool
}

func (s *streamFromUTF8Iterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	item, eof, err := s.upstream.Next()
	if err != nil || eof {
		return nil, eof, err
	}
	bytesVal, ok := item.(*Bytes)
	if !ok {
		return nil, false, recoverableError("stream_state", "fromUtf8() expects bytes input")
	}
	decoded, err := builtinFromUtf8(nil, []Value{bytesVal})
	if err != nil {
		return nil, false, err
	}
	return decoded, false, nil
}

func (s *streamFromUTF8Iterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamToUTF8Iterator struct {
	upstream streamIterator
	closed   bool
}

func (s *streamToUTF8Iterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	item, eof, err := s.upstream.Next()
	if err != nil || eof {
		return nil, eof, err
	}
	textVal, ok := item.(*String)
	if !ok {
		return nil, false, recoverableError("stream_state", "toUtf8() expects text input")
	}
	encoded, err := builtinToUtf8(nil, []Value{textVal})
	if err != nil {
		return nil, false, err
	}
	return encoded, false, nil
}

func (s *streamToUTF8Iterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamDistinctIterator struct {
	upstream streamIterator
	seen     map[MapKey]struct{}
	closed   bool
}

func (s *streamDistinctIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	for {
		item, eof, err := s.upstream.Next()
		if err != nil || eof {
			return nil, eof, err
		}
		key, err := mapKeyForValue(item)
		if err != nil {
			return nil, false, &RuntimeError{Message: "distinct values must be hashable"}
		}
		if _, ok := s.seen[key]; ok {
			continue
		}
		s.seen[key] = struct{}{}
		return cloneStreamItem(item), false, nil
	}
}

func (s *streamDistinctIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamSortIterator struct {
	upstream streamIterator
	eval     *Evaluator
	cmp      Value
	items    []Value
	index    int
	loaded   bool
	closed   bool
}

func (s *streamSortIterator) Next() (Value, bool, error) {
	if s == nil || s.upstream == nil {
		return nil, true, nil
	}
	if !s.loaded {
		if err := s.loadSorted(); err != nil {
			return nil, false, err
		}
	}
	if s.index >= len(s.items) {
		return nil, true, nil
	}
	item := s.items[s.index]
	s.index++
	return item, false, nil
}

func (s *streamSortIterator) loadSorted() error {
	s.loaded = true
	s.items = make([]Value, 0, 64)
	for {
		item, eof, err := s.upstream.Next()
		if err != nil {
			return err
		}
		if eof {
			break
		}
		s.items = append(s.items, cloneStreamItem(item))
	}
	if len(s.items) <= 1 {
		return nil
	}

	if s.cmp != nil {
		var cmpErr error
		sort.SliceStable(s.items, func(i, j int) bool {
			if cmpErr != nil {
				return false
			}
			val, sig, err := s.eval.applyFunction(s.cmp, []Value{s.items[i], s.items[j]})
			if err != nil {
				cmpErr = err
				return false
			}
			if sig != nil {
				cmpErr = &RuntimeError{Message: "break/continue outside loop"}
				return false
			}
			num, _, ok := numberArg(val)
			if !ok {
				cmpErr = &RuntimeError{Message: "sort comparator must return number"}
				return false
			}
			return num < 0
		})
		return cmpErr
	}

	var defaultErr error
	sort.SliceStable(s.items, func(i, j int) bool {
		if defaultErr != nil {
			return false
		}
		less, err := streamDefaultSortLess(s.items[i], s.items[j])
		if err != nil {
			defaultErr = err
			return false
		}
		return less
	})
	return defaultErr
}

func (s *streamSortIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

type streamTopEntry struct {
	item  Value
	score float64
}

func streamTopScore(e *Evaluator, item Value, scoreFn Value) (float64, error) {
	if scoreFn == nil {
		score, _, ok := numberArg(item)
		if !ok {
			return 0, &RuntimeError{Message: "top expects numeric stream values when score function is omitted"}
		}
		return score, nil
	}
	val, sig, err := e.applyFunction(scoreFn, []Value{item})
	if err != nil {
		return 0, err
	}
	if sig != nil {
		return 0, &RuntimeError{Message: "break/continue outside loop"}
	}
	score, _, ok := numberArg(val)
	if !ok {
		return 0, &RuntimeError{Message: "top score function must return number"}
	}
	return score, nil
}

func streamDefaultSortLess(left Value, right Value) (bool, error) {
	switch l := left.(type) {
	case *Integer:
		r, ok := right.(*Integer)
		if !ok {
			return false, &RuntimeError{Message: "sort default comparator requires homogeneous int/float/string/char values"}
		}
		return l.Value < r.Value, nil
	case *Float:
		r, ok := right.(*Float)
		if !ok {
			return false, &RuntimeError{Message: "sort default comparator requires homogeneous int/float/string/char values"}
		}
		return l.Value < r.Value, nil
	case *String:
		r, ok := right.(*String)
		if !ok {
			return false, &RuntimeError{Message: "sort default comparator requires homogeneous int/float/string/char values"}
		}
		return l.Value < r.Value, nil
	case *Char:
		r, ok := right.(*Char)
		if !ok {
			return false, &RuntimeError{Message: "sort default comparator requires homogeneous int/float/string/char values"}
		}
		return l.Value < r.Value, nil
	default:
		return false, &RuntimeError{Message: "sort default comparator requires int/float/string/char values"}
	}
}

func cloneAggregateValue(v Value) Value {
	switch val := v.(type) {
	case *Integer:
		return &Integer{Value: val.Value}
	case *Float:
		return &Float{Value: val.Value}
	case *Boolean:
		return &Boolean{Value: val.Value}
	case *String:
		return &String{Value: val.Value}
	case *Bytes:
		return &Bytes{Value: append([]byte{}, val.Value...)}
	case *Char:
		return &Char{Value: val.Value}
	case *Array:
		out := make([]Value, 0, len(val.Elements))
		for _, el := range val.Elements {
			out = append(out, cloneAggregateValue(el))
		}
		return &Array{Elements: out}
	case *Object:
		out := make(map[string]Value, len(val.Pairs))
		for k, vv := range val.Pairs {
			out[k] = cloneAggregateValue(vv)
		}
		return &Object{Pairs: out}
	case *Map:
		out := make(map[MapKey]Value, len(val.Pairs))
		for k, vv := range val.Pairs {
			out[k] = cloneAggregateValue(vv)
		}
		return &Map{Pairs: out}
	case *Set:
		out := make(map[MapKey]struct{}, len(val.Elements))
		for k := range val.Elements {
			out[k] = struct{}{}
		}
		return &Set{Elements: out}
	default:
		return v
	}
}
