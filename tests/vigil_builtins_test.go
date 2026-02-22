package tests

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"karl/interpreter"
)

func TestVigilBuiltinsSQLCRUD(t *testing.T) {
	ensureFakeSQLDriverRegistered()
	dsn := fmt.Sprintf("crud-%d", time.Now().UnixNano())
	fakeDriverResetDSN(dsn)

	input := fmt.Sprintf(`
let db = sqlOpen("%s")
let inserted = sqlExec(db,
    "INSERT INTO users (id, email) VALUES ($1, $2) ON CONFLICT DO NOTHING",
    ["u1", "u1@acme.com"]
)
let rows = sqlQuery(db,
    "SELECT id, email, last_email_received FROM users WHERE id = $1",
    ["u1"]
)
sqlClose(db)
{
    rowsAffected: inserted.rowsAffected,
    count: rows.length,
    id: rows[0].id,
    email: rows[0].email,
    ts: rows[0].last_email_received,
}
`, dsn)

	val, err := evalWithConfiguredEvaluator(t, input, func(eval *interpreter.Evaluator) {
		eval.SetSQLDriver(fakeSQLDriverName)
	})
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	obj, ok := val.(*Object)
	if !ok {
		t.Fatalf("expected object, got %T", val)
	}
	assertInteger(t, obj.Pairs["rowsAffected"], 1)
	assertInteger(t, obj.Pairs["count"], 1)
	assertString(t, obj.Pairs["id"], "u1")
	assertString(t, obj.Pairs["email"], "u1@acme.com")
	assertInteger(t, obj.Pairs["ts"], 1700000000000)
}

func TestVigilBuiltinsSQLTxAndQueryOne(t *testing.T) {
	ensureFakeSQLDriverRegistered()
	dsn := fmt.Sprintf("tx-%d", time.Now().UnixNano())
	fakeDriverResetDSN(dsn)

	input := fmt.Sprintf(`
let db = sqlOpen("%s")
let missing = sqlQueryOne(db, "SELECT id, email FROM users WHERE id = $1", ["missing"])

let tx1 = sqlBegin(db)
sqlExec(tx1, "INSERT INTO users (id, email) VALUES ($1, $2) ON CONFLICT DO NOTHING", ["u2", "u2@acme.com"])
sqlRollback(tx1)
let afterRollback = sqlQueryOne(db, "SELECT id, email FROM users WHERE id = $1", ["u2"])

let tx2 = sqlBegin(db)
sqlExec(tx2, "INSERT INTO users (id, email) VALUES ($1, $2) ON CONFLICT DO NOTHING", ["u3", "u3@acme.com"])
sqlCommit(tx2)
let afterCommit = sqlQueryOne(db, "SELECT id, email FROM users WHERE id = $1", ["u3"])

sqlClose(db);
[missing, afterRollback, afterCommit.id]
`, dsn)

	val, err := evalWithConfiguredEvaluator(t, input, func(eval *interpreter.Evaluator) {
		eval.SetSQLDriver(fakeSQLDriverName)
	})
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
	if !Equivalent(arr.Elements[0], NullValue) {
		t.Fatalf("expected first element null, got %s", arr.Elements[0].Inspect())
	}
	if !Equivalent(arr.Elements[1], NullValue) {
		t.Fatalf("expected second element null, got %s", arr.Elements[1].Inspect())
	}
	assertString(t, arr.Elements[2], "u3")
}

func TestVigilBuiltinsSHAUUIDTime(t *testing.T) {
	val := mustEval(t, `
let hash = sha256("abc")
let id = uuidNew()
let normalized = uuidParse("550E8400-E29B-41D4-A716-446655440000")
let parsed = timeParseRFC3339("2026-02-22T00:00:00Z")
let formatted = timeFormatRFC3339(parsed)
let next = timeAdd(parsed, 30 * 1000)
let delta = timeDiff(next, parsed)
;
[hash, uuidValid(id), uuidValid("bad"), normalized, formatted, delta]
`)

	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertString(t, arr.Elements[0], "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	if b, ok := arr.Elements[1].(*Boolean); !ok || !b.Value {
		t.Fatalf("expected uuidValid(id) true, got %v", arr.Elements[1])
	}
	if b, ok := arr.Elements[2].(*Boolean); !ok || b.Value {
		t.Fatalf("expected uuidValid(\"bad\") false, got %v", arr.Elements[2])
	}
	assertString(t, arr.Elements[3], "550e8400-e29b-41d4-a716-446655440000")
	assertString(t, arr.Elements[4], "2026-02-22T00:00:00Z")
	assertInteger(t, arr.Elements[5], 30000)

	val2 := mustEval(t, `uuidParse("bad") ? { error.kind }`)
	assertString(t, val2, "uuidParse")
}

func TestVigilBuiltinsSignalWatchType(t *testing.T) {
	val := mustEval(t, `let ch = signalWatch(["SIGINT", "SIGTERM"]); done(ch); "ok"`)
	assertString(t, val, "ok")

	_, err := evalInput(t, `signalWatch(["NOPE"])`)
	if err == nil {
		t.Fatalf("expected signalWatch unknown signal error")
	}
	if !strings.Contains(err.Error(), "unknown signal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVigilBuiltinsHTTPServeAndStop(t *testing.T) {
	addr := reserveLocalAddr(t)
	input := fmt.Sprintf(`
let srv = httpServe({
    addr: "%s",
    routes: [
        {
            method: "GET",
            path: "/health",
            handler: req -> {
                status: 200,
                body: jsonEncode({ status: "ok", method: req.method, q: req.query.q, }),
            }
        }
    ]
})

sleep(80)
let resp = http({ method: "GET", url: "http://%s/health?q=test" })
let payload = jsonDecode(resp.body)
httpServerStop(srv)
;
[resp.status, payload.status, payload.method, payload.q]
`, addr, addr)

	val := mustEval(t, input)
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertInteger(t, arr.Elements[0], 200)
	assertString(t, arr.Elements[1], "ok")
	assertString(t, arr.Elements[2], "GET")
	assertString(t, arr.Elements[3], "test")
}

func reserveLocalAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve listener: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

const fakeSQLDriverName = "karltestsql"

var (
	fakeSQLDriverOnce sync.Once
	fakeDriver        = &fakeSQLDriver{stores: map[string]*fakeSQLStore{}}
)

func ensureFakeSQLDriverRegistered() {
	fakeSQLDriverOnce.Do(func() {
		sql.Register(fakeSQLDriverName, fakeDriver)
	})
}

func fakeDriverResetDSN(dsn string) {
	fakeDriver.mu.Lock()
	defer fakeDriver.mu.Unlock()
	fakeDriver.stores[dsn] = &fakeSQLStore{users: map[string]fakeSQLUser{}}
}

type fakeSQLDriver struct {
	mu     sync.Mutex
	stores map[string]*fakeSQLStore
}

type fakeSQLStore struct {
	users map[string]fakeSQLUser
}

type fakeSQLUser struct {
	id                string
	email             string
	lastEmailReceived *time.Time
	lastEmailCheckMs  int64
}

func (d *fakeSQLDriver) Open(name string) (driver.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.stores[name]; !ok {
		d.stores[name] = &fakeSQLStore{users: map[string]fakeSQLUser{}}
	}
	return &fakeSQLConn{driver: d, dsn: name}, nil
}

type fakeSQLConn struct {
	driver *fakeSQLDriver
	dsn    string

	closed bool
	inTx   bool
	txData *fakeSQLStore
}

func (c *fakeSQLConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}

func (c *fakeSQLConn) Close() error {
	c.closed = true
	return nil
}

func (c *fakeSQLConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *fakeSQLConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	if c.closed {
		return nil, errors.New("connection closed")
	}
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.inTx {
		return nil, errors.New("transaction already started")
	}
	base := c.driver.stores[c.dsn]
	c.txData = base.clone()
	c.inTx = true
	return &fakeSQLTx{conn: c}, nil
}

func (c *fakeSQLConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	norm := normalizeSQL(query)
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	store := c.currentStoreLocked()

	switch {
	case strings.HasPrefix(norm, "insert into users"):
		if len(args) < 2 {
			return nil, errors.New("insert expects 2 args")
		}
		id, err := argString(args[0])
		if err != nil {
			return nil, err
		}
		email, err := argString(args[1])
		if err != nil {
			return nil, err
		}
		if _, exists := store.users[id]; exists {
			return driver.RowsAffected(0), nil
		}
		fixed := time.Unix(1700000000, 0).UTC()
		store.users[id] = fakeSQLUser{
			id:                id,
			email:             email,
			lastEmailReceived: &fixed,
		}
		return driver.RowsAffected(1), nil
	case strings.HasPrefix(norm, "update users set last_email_check"):
		if len(args) < 2 {
			return nil, errors.New("update expects 2 args")
		}
		ms, err := argInt64(args[0])
		if err != nil {
			return nil, err
		}
		id, err := argString(args[1])
		if err != nil {
			return nil, err
		}
		user, exists := store.users[id]
		if !exists {
			return driver.RowsAffected(0), nil
		}
		user.lastEmailCheckMs = ms
		store.users[id] = user
		return driver.RowsAffected(1), nil
	default:
		return nil, fmt.Errorf("unsupported exec query: %s", query)
	}
}

func (c *fakeSQLConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	norm := normalizeSQL(query)
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	store := c.currentStoreLocked()

	switch {
	case strings.Contains(norm, "select id, email, last_email_received from users where id = $1"):
		if len(args) < 1 {
			return nil, errors.New("select expects id")
		}
		id, err := argString(args[0])
		if err != nil {
			return nil, err
		}
		user, exists := store.users[id]
		if !exists {
			return &fakeSQLRows{cols: []string{"id", "email", "last_email_received"}}, nil
		}
		row := []driver.Value{user.id, user.email, nil}
		if user.lastEmailReceived != nil {
			row[2] = *user.lastEmailReceived
		}
		return &fakeSQLRows{
			cols: []string{"id", "email", "last_email_received"},
			rows: [][]driver.Value{row},
		}, nil
	case strings.Contains(norm, "select id, email from users where id = $1"):
		if len(args) < 1 {
			return nil, errors.New("select expects id")
		}
		id, err := argString(args[0])
		if err != nil {
			return nil, err
		}
		user, exists := store.users[id]
		if !exists {
			return &fakeSQLRows{cols: []string{"id", "email"}}, nil
		}
		return &fakeSQLRows{
			cols: []string{"id", "email"},
			rows: [][]driver.Value{{user.id, user.email}},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported query: %s", query)
	}
}

func (c *fakeSQLConn) currentStoreLocked() *fakeSQLStore {
	if c.inTx && c.txData != nil {
		return c.txData
	}
	return c.driver.stores[c.dsn]
}

type fakeSQLTx struct {
	conn *fakeSQLConn
}

func (tx *fakeSQLTx) Commit() error {
	c := tx.conn
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if !c.inTx || c.txData == nil {
		return errors.New("no active transaction")
	}
	c.driver.stores[c.dsn] = c.txData
	c.txData = nil
	c.inTx = false
	return nil
}

func (tx *fakeSQLTx) Rollback() error {
	c := tx.conn
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if !c.inTx {
		return errors.New("no active transaction")
	}
	c.txData = nil
	c.inTx = false
	return nil
}

type fakeSQLRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *fakeSQLRows) Columns() []string {
	return r.cols
}

func (r *fakeSQLRows) Close() error {
	return nil
}

func (r *fakeSQLRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.idx]
	r.idx++
	copy(dest, row)
	return nil
}

func (s *fakeSQLStore) clone() *fakeSQLStore {
	out := &fakeSQLStore{users: make(map[string]fakeSQLUser, len(s.users))}
	for k, v := range s.users {
		copyUser := v
		if v.lastEmailReceived != nil {
			ts := *v.lastEmailReceived
			copyUser.lastEmailReceived = &ts
		}
		out.users[k] = copyUser
	}
	return out
}

func normalizeSQL(query string) string {
	parts := strings.Fields(strings.ToLower(query))
	return strings.Join(parts, " ")
}

func argString(v driver.NamedValue) (string, error) {
	s, ok := v.Value.(string)
	if !ok {
		return "", fmt.Errorf("expected string arg, got %T", v.Value)
	}
	return s, nil
}

func argInt64(v driver.NamedValue) (int64, error) {
	switch n := v.Value.(type) {
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("expected int64 arg, got %T", v.Value)
	}
}
