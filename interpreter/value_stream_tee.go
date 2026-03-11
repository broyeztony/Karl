package interpreter

type teeSideMessage struct {
	item Value
	eof  bool
}

type teeSideIterator struct {
	ch <-chan teeSideMessage
}

func (t *teeSideIterator) Next() (Value, bool, error) {
	if t == nil || t.ch == nil {
		return nil, true, nil
	}
	msg, ok := <-t.ch
	if !ok || msg.eof {
		return nil, true, nil
	}
	return msg.item, false, nil
}

func (t *teeSideIterator) Close() error { return nil }

type teeStreamIterator struct {
	eval     *Evaluator
	upstream streamIterator
	side     *StreamSinkValue

	started bool
	closed  bool

	sideCh     chan teeSideMessage
	sideDone   chan error
	sideClosed bool
	sideErr    error
	sideReady  bool
}

func newTeeStreamIterator(eval *Evaluator, upstream streamIterator, side *StreamSinkValue) streamIterator {
	return &teeStreamIterator{
		eval:     eval,
		upstream: upstream,
		side:     side,
	}
}

func (t *teeStreamIterator) Next() (Value, bool, error) {
	if t == nil || t.closed || t.upstream == nil {
		return nil, true, nil
	}
	if t.sideErr != nil {
		return nil, false, t.sideErr
	}
	t.startSideIfNeeded()

	item, eof, err := t.upstream.Next()
	if err != nil {
		t.closeSide()
		_ = t.awaitSideResult()
		return nil, false, err
	}
	if eof {
		t.closeSide()
		if err := t.awaitSideResult(); err != nil {
			return nil, false, err
		}
		return nil, true, nil
	}

	if err := t.sendToSide(cloneStreamItem(item)); err != nil {
		return nil, false, err
	}
	return item, false, nil
}

func (t *teeStreamIterator) Close() error {
	if t == nil || t.closed {
		return nil
	}
	t.closed = true

	var firstErr error
	if t.upstream != nil {
		if err := t.upstream.Close(); err != nil {
			firstErr = err
		}
	}
	t.closeSide()
	if err := t.awaitSideResult(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func (t *teeStreamIterator) startSideIfNeeded() {
	if t == nil || t.started || t.side == nil || t.side.run == nil {
		return
	}
	t.started = true
	t.sideCh = make(chan teeSideMessage, 1)
	t.sideDone = make(chan error, 1)

	go func() {
		_, err := t.side.run(t.eval, &teeSideIterator{ch: t.sideCh})
		t.sideDone <- err
	}()
}

func (t *teeStreamIterator) sendToSide(item Value) error {
	if t == nil || !t.started || t.sideClosed {
		return nil
	}
	if t.sideErr != nil {
		return t.sideErr
	}
	select {
	case t.sideCh <- teeSideMessage{item: item}:
		return nil
	case err := <-t.sideDone:
		t.sideReady = true
		t.sideErr = err
		if err != nil {
			return err
		}
		return &RuntimeError{Message: "tee sink finished before upstream EOF"}
	}
}

func (t *teeStreamIterator) closeSide() {
	if t == nil || !t.started || t.sideClosed {
		return
	}
	t.sideClosed = true
	close(t.sideCh)
}

func (t *teeStreamIterator) awaitSideResult() error {
	if t == nil || !t.started || t.sideReady {
		return t.sideErr
	}
	err := <-t.sideDone
	t.sideReady = true
	t.sideErr = err
	return err
}
