package interpreter

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
)

type SQLDB struct {
	DB *sql.DB

	mu     sync.Mutex
	closed bool
}

func (d *SQLDB) Type() ValueType { return DB }
func (d *SQLDB) Inspect() string { return "<db>" }

func (d *SQLDB) Close() error {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	db := d.DB
	d.mu.Unlock()
	if db == nil {
		return nil
	}
	return db.Close()
}

func (d *SQLDB) isClosed() bool {
	if d == nil {
		return true
	}
	d.mu.Lock()
	closed := d.closed
	d.mu.Unlock()
	return closed
}

type SQLTx struct {
	Tx *sql.Tx

	mu         sync.Mutex
	done       bool
	cancel     context.CancelFunc
	cancelOnce sync.Once
}

func (t *SQLTx) Type() ValueType { return TX }
func (t *SQLTx) Inspect() string { return "<tx>" }

func (t *SQLTx) isDone() bool {
	if t == nil {
		return true
	}
	t.mu.Lock()
	done := t.done
	t.mu.Unlock()
	return done
}

func (t *SQLTx) markDone() {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.done = true
	t.mu.Unlock()
	t.cancelOnce.Do(func() {
		if t.cancel != nil {
			t.cancel()
		}
	})
}

type HTTPServer struct {
	Server *http.Server

	stopMu  sync.Mutex
	stopped bool
	evalMu  sync.Mutex
}

func (s *HTTPServer) Type() ValueType { return SERVER }
func (s *HTTPServer) Inspect() string { return "<server>" }

func (s *HTTPServer) Stop(ctx context.Context) error {
	if s == nil || s.Server == nil {
		return nil
	}
	s.stopMu.Lock()
	if s.stopped {
		s.stopMu.Unlock()
		return nil
	}
	s.stopped = true
	server := s.Server
	s.stopMu.Unlock()
	return server.Shutdown(ctx)
}
