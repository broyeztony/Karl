package interpreter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

func registerSQLBuiltins() {
	builtins["sqlOpen"] = &Builtin{Name: "sqlOpen", Fn: builtinSQLOpen}
	builtins["sqlClose"] = &Builtin{Name: "sqlClose", Fn: builtinSQLClose}
	builtins["sqlExec"] = &Builtin{Name: "sqlExec", Fn: builtinSQLExec}
	builtins["sqlQuery"] = &Builtin{Name: "sqlQuery", Fn: builtinSQLQuery}
	builtins["sqlQueryOne"] = &Builtin{Name: "sqlQueryOne", Fn: builtinSQLQueryOne}
	builtins["sqlBegin"] = &Builtin{Name: "sqlBegin", Fn: builtinSQLBegin}
	builtins["sqlCommit"] = &Builtin{Name: "sqlCommit", Fn: builtinSQLCommit}
	builtins["sqlRollback"] = &Builtin{Name: "sqlRollback", Fn: builtinSQLRollback}
}

func builtinSQLOpen(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sqlOpen expects dsn"}
	}
	dsn, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "sqlOpen expects string dsn"}
	}
	driver := runtimeSQLDriver(e)
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, recoverableError("sqlOpen", "sqlOpen error: "+err.Error())
	}
	ctx, done := runtimeCancelableContext(e)
	defer done()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if mapped := mapRuntimeContextError(e, err); mapped != err {
			return nil, mapped
		}
		return nil, recoverableError("sqlOpen", "sqlOpen error: "+err.Error())
	}
	return &SQLDB{DB: db}, nil
}

func builtinSQLClose(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sqlClose expects db"}
	}
	db, ok := args[0].(*SQLDB)
	if !ok {
		return nil, &RuntimeError{Message: "sqlClose expects db"}
	}
	if err := db.Close(); err != nil {
		return nil, recoverableError("sqlClose", "sqlClose error: "+err.Error())
	}
	return UnitValue, nil
}

func builtinSQLExec(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "sqlExec expects connOrTx, query, params"}
	}
	execConn, err := sqlExecQuerierFromValue(args[0])
	if err != nil {
		return nil, err
	}
	query, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "sqlExec expects string query"}
	}
	params, err := sqlParamsArray(args[2])
	if err != nil {
		return nil, err
	}

	ctx, done := runtimeCancelableContext(e)
	defer done()
	result, err := execConn.ExecContext(ctx, query, params...)
	if err != nil {
		if mapped := mapRuntimeContextError(e, err); mapped != err {
			return nil, mapped
		}
		return nil, recoverableError("sqlExec", "sqlExec error: "+err.Error())
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, recoverableError("sqlExec", "sqlExec error: "+err.Error())
	}
	return &Object{Pairs: map[string]Value{
		"rowsAffected": &Integer{Value: rowsAffected},
	}}, nil
}

func builtinSQLQuery(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "sqlQuery expects connOrTx, query, params"}
	}
	execConn, err := sqlExecQuerierFromValue(args[0])
	if err != nil {
		return nil, err
	}
	query, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "sqlQuery expects string query"}
	}
	params, err := sqlParamsArray(args[2])
	if err != nil {
		return nil, err
	}

	ctx, done := runtimeCancelableContext(e)
	defer done()
	rows, err := execConn.QueryContext(ctx, query, params...)
	if err != nil {
		if mapped := mapRuntimeContextError(e, err); mapped != err {
			return nil, mapped
		}
		return nil, recoverableError("sqlQuery", "sqlQuery error: "+err.Error())
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, recoverableError("sqlQuery", "sqlQuery error: "+err.Error())
	}

	out := make([]Value, 0)
	for rows.Next() {
		raw := make([]interface{}, len(cols))
		dest := make([]interface{}, len(cols))
		for i := range raw {
			dest[i] = &raw[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, recoverableError("sqlQuery", "sqlQuery error: "+err.Error())
		}

		pairs := make(map[string]Value, len(cols))
		for i, col := range cols {
			val, err := sqlResultValue(raw[i])
			if err != nil {
				return nil, recoverableError("sqlQuery", "sqlQuery error: "+err.Error())
			}
			pairs[col] = val
		}
		out = append(out, &Object{Pairs: pairs})
	}

	if err := rows.Err(); err != nil {
		if mapped := mapRuntimeContextError(e, err); mapped != err {
			return nil, mapped
		}
		return nil, recoverableError("sqlQuery", "sqlQuery error: "+err.Error())
	}

	return &Array{Elements: out}, nil
}

func builtinSQLQueryOne(e *Evaluator, args []Value) (Value, error) {
	rows, err := builtinSQLQuery(e, args)
	if err != nil {
		return nil, err
	}
	arr := rows.(*Array)
	if len(arr.Elements) == 0 {
		return NullValue, nil
	}
	return arr.Elements[0], nil
}

func builtinSQLBegin(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sqlBegin expects db"}
	}
	db, ok := args[0].(*SQLDB)
	if !ok {
		return nil, &RuntimeError{Message: "sqlBegin expects db"}
	}
	if db.isClosed() {
		return nil, &RuntimeError{Message: "sqlBegin on closed db"}
	}

	ctx, done := runtimeCancelableContext(e)
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		done()
		if mapped := mapRuntimeContextError(e, err); mapped != err {
			return nil, mapped
		}
		return nil, recoverableError("sqlBegin", "sqlBegin error: "+err.Error())
	}
	return &SQLTx{Tx: tx, cancel: done}, nil
}

func builtinSQLCommit(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sqlCommit expects tx"}
	}
	tx, ok := args[0].(*SQLTx)
	if !ok {
		return nil, &RuntimeError{Message: "sqlCommit expects tx"}
	}
	if tx.isDone() {
		return nil, &RuntimeError{Message: "sqlCommit on closed tx"}
	}

	err := tx.Tx.Commit()
	if err != nil {
		if mapped := mapRuntimeContextError(e, err); mapped != err {
			return nil, mapped
		}
		return nil, recoverableError("sqlCommit", "sqlCommit error: "+err.Error())
	}
	tx.markDone()
	return UnitValue, nil
}

func builtinSQLRollback(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sqlRollback expects tx"}
	}
	tx, ok := args[0].(*SQLTx)
	if !ok {
		return nil, &RuntimeError{Message: "sqlRollback expects tx"}
	}
	if tx.isDone() {
		return nil, &RuntimeError{Message: "sqlRollback on closed tx"}
	}

	err := tx.Tx.Rollback()
	if err != nil {
		if mapped := mapRuntimeContextError(e, err); mapped != err {
			return nil, mapped
		}
		return nil, recoverableError("sqlRollback", "sqlRollback error: "+err.Error())
	}
	tx.markDone()
	return UnitValue, nil
}

type sqlExecQuerier interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}

func sqlExecQuerierFromValue(val Value) (sqlExecQuerier, error) {
	switch c := val.(type) {
	case *SQLDB:
		if c.isClosed() {
			return nil, &RuntimeError{Message: "sql connection is closed"}
		}
		if c.DB == nil {
			return nil, &RuntimeError{Message: "sql connection unavailable"}
		}
		return c.DB, nil
	case *SQLTx:
		if c.isDone() {
			return nil, &RuntimeError{Message: "sql transaction is closed"}
		}
		if c.Tx == nil {
			return nil, &RuntimeError{Message: "sql transaction unavailable"}
		}
		return c.Tx, nil
	default:
		return nil, &RuntimeError{Message: "sql expects db or tx"}
	}
}

func sqlParamsArray(val Value) ([]interface{}, error) {
	arr, ok := val.(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "sql params must be array"}
	}
	out := make([]interface{}, 0, len(arr.Elements))
	for _, el := range arr.Elements {
		converted, err := sqlParamValue(el)
		if err != nil {
			return nil, err
		}
		out = append(out, converted)
	}
	return out, nil
}

func sqlParamValue(val Value) (interface{}, error) {
	switch v := val.(type) {
	case *Null:
		return nil, nil
	case *Boolean:
		return v.Value, nil
	case *Integer:
		return v.Value, nil
	case *Float:
		return v.Value, nil
	case *String:
		return v.Value, nil
	case *Char:
		return v.Value, nil
	default:
		return nil, &RuntimeError{Message: "sql params support null, bool, number, string"}
	}
}

func sqlResultValue(raw interface{}) (Value, error) {
	switch v := raw.(type) {
	case nil:
		return NullValue, nil
	case bool:
		return &Boolean{Value: v}, nil
	case int:
		return &Integer{Value: int64(v)}, nil
	case int8:
		return &Integer{Value: int64(v)}, nil
	case int16:
		return &Integer{Value: int64(v)}, nil
	case int32:
		return &Integer{Value: int64(v)}, nil
	case int64:
		return &Integer{Value: v}, nil
	case uint:
		if uint64(v) > math.MaxInt64 {
			return nil, fmt.Errorf("integer overflow")
		}
		return &Integer{Value: int64(v)}, nil
	case uint8:
		return &Integer{Value: int64(v)}, nil
	case uint16:
		return &Integer{Value: int64(v)}, nil
	case uint32:
		return &Integer{Value: int64(v)}, nil
	case uint64:
		if v > math.MaxInt64 {
			return nil, fmt.Errorf("integer overflow")
		}
		return &Integer{Value: int64(v)}, nil
	case float32:
		return &Float{Value: float64(v)}, nil
	case float64:
		return &Float{Value: v}, nil
	case string:
		return &String{Value: v}, nil
	case []byte:
		return &String{Value: string(v)}, nil
	case time.Time:
		return &Integer{Value: v.UTC().UnixNano() / int64(time.Millisecond)}, nil
	case []interface{}:
		items := make([]Value, 0, len(v))
		for _, el := range v {
			item, err := sqlResultValue(el)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return &Array{Elements: items}, nil
	case map[string]interface{}:
		pairs := make(map[string]Value, len(v))
		for k, el := range v {
			item, err := sqlResultValue(el)
			if err != nil {
				return nil, err
			}
			pairs[k] = item
		}
		return &Object{Pairs: pairs}, nil
	case fmt.Stringer:
		return &String{Value: v.String()}, nil
	default:
		return nil, fmt.Errorf("unsupported SQL result type %T", raw)
	}
}

func runtimeSQLDriver(e *Evaluator) string {
	if e == nil || e.runtime == nil {
		return "pgx"
	}
	return e.runtime.getSQLDriver()
}

func runtimeCancelableContext(e *Evaluator) (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	cancelCh := runtimeCancelSignal(e)
	fatalCh := runtimeFatalSignal(e)
	if cancelCh == nil && fatalCh == nil {
		return ctx, cancel
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-cancelCh:
			cancel()
		case <-fatalCh:
			cancel()
		case <-done:
			cancel()
		}
	}()
	return ctx, func() {
		close(done)
		cancel()
	}
}

func mapRuntimeContextError(e *Evaluator, err error) error {
	if !errors.Is(err, context.Canceled) {
		return err
	}
	cancelCh := runtimeCancelSignal(e)
	if cancelCh != nil {
		select {
		case <-cancelCh:
			return canceledError()
		default:
		}
	}
	fatalCh := runtimeFatalSignal(e)
	if fatalCh != nil {
		select {
		case <-fatalCh:
			return runtimeFatalError(e)
		default:
		}
	}
	return canceledError()
}
