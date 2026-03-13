package interpreter

type debounceEvent struct {
	item Value
	eof  bool
	err  error
}

type streamDebounceIterator struct {
	upstream streamIterator
	eval     *Evaluator
	delayMs  int64

	events    chan debounceEvent
	stopCh    chan struct{}
	started   bool
	pending   Value
	drained   bool
	closed    bool
	pumpEnded bool
}

func newStreamDebounceIterator(upstream streamIterator, eval *Evaluator, ms int64) *streamDebounceIterator {
	return &streamDebounceIterator{
		upstream: upstream,
		eval:     eval,
		delayMs:  ms,
		events:   make(chan debounceEvent, 64),
		stopCh:   make(chan struct{}),
	}
}

func (s *streamDebounceIterator) Next() (Value, bool, error) {
	if s == nil || s.closed || s.upstream == nil || s.drained {
		return nil, true, nil
	}
	s.startPumpIfNeeded()

	if s.pending == nil {
		ev, ok := s.nextEvent()
		if !ok {
			s.drained = true
			return nil, true, nil
		}
		if ev.err != nil {
			return nil, false, ev.err
		}
		if ev.eof {
			s.drained = true
			return nil, true, nil
		}
		s.pending = ev.item
	}

	if s.delayMs <= 0 {
		out := cloneStreamItem(s.pending)
		s.pending = nil
		return out, false, nil
	}

	if _, err := builtinSleep(s.eval, []Value{&Integer{Value: s.delayMs}}); err != nil {
		return nil, false, err
	}

	for {
		select {
		case ev, ok := <-s.events:
			if !ok || ev.eof {
				out := cloneStreamItem(s.pending)
				s.pending = nil
				s.drained = true
				return out, false, nil
			}
			if ev.err != nil {
				return nil, false, ev.err
			}
			s.pending = ev.item
		default:
			out := cloneStreamItem(s.pending)
			s.pending = nil
			return out, false, nil
		}
	}
}

func (s *streamDebounceIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.stopCh != nil {
		close(s.stopCh)
	}
	if s.upstream != nil {
		return s.upstream.Close()
	}
	return nil
}

func (s *streamDebounceIterator) startPumpIfNeeded() {
	if s == nil || s.started {
		return
	}
	s.started = true
	go func() {
		defer func() {
			close(s.events)
			s.pumpEnded = true
		}()
		for {
			item, eof, err := s.upstream.Next()
			ev := debounceEvent{item: item, eof: eof, err: err}
			select {
			case <-s.stopCh:
				return
			case s.events <- ev:
			}
			if eof || err != nil {
				return
			}
		}
	}()
}

func (s *streamDebounceIterator) nextEvent() (debounceEvent, bool) {
	if s == nil {
		return debounceEvent{eof: true}, false
	}
	select {
	case <-s.stopCh:
		return debounceEvent{eof: true}, false
	case ev, ok := <-s.events:
		return ev, ok
	}
}
