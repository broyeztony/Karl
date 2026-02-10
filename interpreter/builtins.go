package interpreter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var builtins = map[string]*Builtin{}

func RegisterBuiltins() {
	builtins["exit"] = &Builtin{Name: "exit", Fn: builtinExit}
	builtins["fail"] = &Builtin{Name: "fail", Fn: builtinFail}
	builtins["rendezvous"] = &Builtin{Name: "rendezvous", Fn: builtinChannel}
	builtins["channel"] = &Builtin{Name: "channel", Fn: builtinChannel}
	builtins["buffered"] = &Builtin{Name: "buffered", Fn: builtinBufferedChannel}
	builtins["sleep"] = &Builtin{Name: "sleep", Fn: builtinSleep}
	builtins["log"] = &Builtin{Name: "log", Fn: builtinLog}
	builtins["str"] = &Builtin{Name: "str", Fn: builtinStr}
	builtins["rand"] = &Builtin{Name: "rand", Fn: builtinRand}
	builtins["randInt"] = &Builtin{Name: "randInt", Fn: builtinRandInt}
	builtins["randFloat"] = &Builtin{Name: "randFloat", Fn: builtinRandFloat}
	builtins["parseInt"] = &Builtin{Name: "parseInt", Fn: builtinParseInt}
	builtins["now"] = &Builtin{Name: "now", Fn: builtinNow}
	builtins["readFile"] = &Builtin{Name: "readFile", Fn: builtinReadFile}
	builtins["writeFile"] = &Builtin{Name: "writeFile", Fn: builtinWriteFile}
	builtins["appendFile"] = &Builtin{Name: "appendFile", Fn: builtinAppendFile}
	builtins["deleteFile"] = &Builtin{Name: "deleteFile", Fn: builtinDeleteFile}
	builtins["exists"] = &Builtin{Name: "exists", Fn: builtinExists}
	builtins["listDir"] = &Builtin{Name: "listDir", Fn: builtinListDir}
	builtins["http"] = &Builtin{Name: "http", Fn: builtinHTTP}
	builtins["encodeJson"] = &Builtin{Name: "encodeJson", Fn: builtinEncodeJSON}
	builtins["decodeJson"] = &Builtin{Name: "decodeJson", Fn: builtinDecodeJSON}
	builtins["then"] = &Builtin{Name: "then", Fn: builtinThen}
	builtins["send"] = &Builtin{Name: "send", Fn: builtinSend}
	builtins["recv"] = &Builtin{Name: "recv", Fn: builtinRecv}
	builtins["done"] = &Builtin{Name: "done", Fn: builtinDone}
	builtins["split"] = &Builtin{Name: "split", Fn: builtinSplit}
	builtins["chars"] = &Builtin{Name: "chars", Fn: builtinChars}
	builtins["trim"] = &Builtin{Name: "trim", Fn: builtinTrim}
	builtins["toLower"] = &Builtin{Name: "toLower", Fn: builtinToLower}
	builtins["toUpper"] = &Builtin{Name: "toUpper", Fn: builtinToUpper}
	builtins["contains"] = &Builtin{Name: "contains", Fn: builtinContains}
	builtins["startsWith"] = &Builtin{Name: "startsWith", Fn: builtinStartsWith}
	builtins["endsWith"] = &Builtin{Name: "endsWith", Fn: builtinEndsWith}
	builtins["replace"] = &Builtin{Name: "replace", Fn: builtinReplace}
	builtins["map"] = &Builtin{Name: "map", Fn: builtinMap}
	builtins["get"] = &Builtin{Name: "get", Fn: builtinMapGet}
	builtins["set"] = &Builtin{Name: "set", Fn: builtinMapSet}
	builtins["add"] = &Builtin{Name: "add", Fn: builtinSetAdd}
	builtins["has"] = &Builtin{Name: "has", Fn: builtinMapHas}
	builtins["delete"] = &Builtin{Name: "delete", Fn: builtinMapDelete}
	builtins["keys"] = &Builtin{Name: "keys", Fn: builtinMapKeys}
	builtins["values"] = &Builtin{Name: "values", Fn: builtinMapValues}
	builtins["sort"] = &Builtin{Name: "sort", Fn: builtinSort}
	builtins["filter"] = &Builtin{Name: "filter", Fn: builtinFilter}
	builtins["reduce"] = &Builtin{Name: "reduce", Fn: builtinReduce}
	builtins["sum"] = &Builtin{Name: "sum", Fn: builtinSum}
	builtins["find"] = &Builtin{Name: "find", Fn: builtinFind}
	builtins["abs"] = &Builtin{Name: "abs", Fn: builtinAbs}
	builtins["sqrt"] = &Builtin{Name: "sqrt", Fn: builtinSqrt}
	builtins["pow"] = &Builtin{Name: "pow", Fn: builtinPow}
	builtins["sin"] = &Builtin{Name: "sin", Fn: builtinSin}
	builtins["cos"] = &Builtin{Name: "cos", Fn: builtinCos}
	builtins["tan"] = &Builtin{Name: "tan", Fn: builtinTan}
	builtins["floor"] = &Builtin{Name: "floor", Fn: builtinFloor}
	builtins["ceil"] = &Builtin{Name: "ceil", Fn: builtinCeil}
	builtins["min"] = &Builtin{Name: "min", Fn: builtinMin}
	builtins["max"] = &Builtin{Name: "max", Fn: builtinMax}
	builtins["clamp"] = &Builtin{Name: "clamp", Fn: builtinClamp}
}

func getBuiltin(name string) *Builtin {
	if b, ok := builtins[name]; ok {
		return b
	}
	return nil
}

func NewBaseEnvironment() *Environment {
	RegisterBuiltins()
	env := NewEnvironment()
	for name, builtin := range builtins {
		env.Define(name, builtin)
	}
	return env
}

func bindReceiver(fn BuiltinFunction, receiver Value) BuiltinFunction {
	return func(e *Evaluator, args []Value) (Value, error) {
		return fn(e, append([]Value{receiver}, args...))
	}
}

func recoverableError(kind string, msg string) *RecoverableError {
	return &RecoverableError{Kind: kind, Message: msg}
}

func runtimeFatalSignal(e *Evaluator) <-chan struct{} {
	if e == nil || e.runtime == nil {
		return nil
	}
	return e.runtime.fatalSignal()
}

func runtimeFatalError(e *Evaluator) error {
	if e != nil && e.runtime != nil {
		if err := e.runtime.getFatalTaskFailure(); err != nil {
			return err
		}
	}
	return &RuntimeError{Message: "runtime terminated"}
}

func builtinExit(_ *Evaluator, args []Value) (Value, error) {
	msg := ""
	if len(args) > 0 {
		msg = args[0].Inspect()
	}
	exitProcess(msg)
	return nil, &ExitError{Message: msg}
}

func builtinFail(_ *Evaluator, args []Value) (Value, error) {
	if len(args) > 1 {
		return nil, &RuntimeError{Message: "fail expects 0 or 1 argument"}
	}
	msg := ""
	if len(args) == 1 {
		s, ok := args[0].(*String)
		if !ok {
			return nil, &RuntimeError{Message: "fail expects string message"}
		}
		msg = s.Value
	}
	return nil, recoverableError("fail", msg)
}

func builtinChannel(_ *Evaluator, _ []Value) (Value, error) {
	return &Channel{Ch: make(chan Value)}, nil
}

func builtinBufferedChannel(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "buffered expects 1 argument (buffer size)"}
	}
	size, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "buffered expects integer buffer size"}
	}
	if size.Value < 0 {
		return nil, &RuntimeError{Message: "buffered expects non-negative buffer size"}
	}
	if size.Value > 1000000 {
		return nil, &RuntimeError{Message: "buffered buffer size too large (max 1000000)"}
	}
	return &Channel{Ch: make(chan Value, size.Value)}, nil
}

func builtinSleep(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sleep expects 1 argument"}
	}
	ms, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "sleep expects integer milliseconds"}
	}

	d := time.Duration(ms.Value) * time.Millisecond
	if d <= 0 {
		return UnitValue, nil
	}
	fatalCh := runtimeFatalSignal(e)
	cancelCh := (<-chan struct{})(nil)
	if e != nil && e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}
	if cancelCh == nil && fatalCh == nil {
		time.Sleep(d)
		return UnitValue, nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return UnitValue, nil
	case <-cancelCh:
		return nil, canceledError()
	case <-fatalCh:
		return nil, runtimeFatalError(e)
	}
}

func builtinLog(_ *Evaluator, args []Value) (Value, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatLogValue(arg)
	}
	fmt.Println(strings.Join(parts, " "))
	return UnitValue, nil
}

func builtinStr(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "str expects 1 argument"}
	}
	return &String{Value: formatLogValue(args[0])}, nil
}

func formatLogValue(val Value) string {
	switch v := val.(type) {
	case *String:
		return v.Value
	case *Char:
		return v.Value
	case *Null:
		return "null"
	case *Unit:
		return "()"
	default:
		return val.Inspect()
	}
}

func builtinRand(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "rand expects no arguments"}
	}
	return &Integer{Value: rand.Int63()}, nil
}

func builtinRandInt(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "randInt expects min and max"}
	}
	min, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "randInt expects integer min"}
	}
	max, ok := args[1].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "randInt expects integer max"}
	}
	if max.Value < min.Value {
		return nil, &RuntimeError{Message: "randInt expects min <= max"}
	}
	if max.Value == min.Value {
		return &Integer{Value: min.Value}, nil
	}
	diff := max.Value - min.Value
	if diff < 0 || diff == math.MaxInt64 {
		return nil, &RuntimeError{Message: "randInt range too large"}
	}
	n := rand.Int63n(diff+1) + min.Value
	return &Integer{Value: n}, nil
}

func builtinRandFloat(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "randFloat expects min and max"}
	}
	min, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "randFloat expects numeric min"}
	}
	max, _, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "randFloat expects numeric max"}
	}
	if max < min {
		return nil, &RuntimeError{Message: "randFloat expects min <= max"}
	}
	return &Float{Value: min + rand.Float64()*(max-min)}, nil
}

func builtinParseInt(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "parseInt expects 1 argument"}
	}
	s, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "parseInt expects string"}
	}
	n, err := strconv.ParseInt(s.Value, 10, 64)
	if err != nil {
		return nil, &RuntimeError{Message: "invalid integer: " + s.Value}
	}
	return &Integer{Value: n}, nil
}

func builtinNow(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "now expects no arguments"}
	}
	return &Integer{Value: time.Now().UnixNano() / int64(time.Millisecond)}, nil
}

func builtinHTTP(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "http expects request object"}
	}
	reqObj, ok := objectPairs(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "http expects object request"}
	}
	methodVal, ok := reqObj["method"]
	if !ok {
		methodVal = &String{Value: "GET"}
	}
	method, ok := stringArg(methodVal)
	if !ok {
		return nil, &RuntimeError{Message: "http method must be string"}
	}
	urlVal, ok := reqObj["url"]
	if !ok {
		return nil, &RuntimeError{Message: "http expects url"}
	}
	urlStr, ok := stringArg(urlVal)
	if !ok {
		return nil, &RuntimeError{Message: "http url must be string"}
	}
	var body io.Reader
	if bodyVal, ok := reqObj["body"]; ok && bodyVal != NullValue {
		bodyStr, ok := stringArg(bodyVal)
		if !ok {
			return nil, &RuntimeError{Message: "http body must be string"}
		}
		body = strings.NewReader(bodyStr)
	}

	reqDone := make(chan struct{})
	defer close(reqDone)

	ctx, cancel := context.WithCancel(context.Background())
	cancelCh := (<-chan struct{})(nil)
	if e != nil && e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}
	fatalCh := runtimeFatalSignal(e)
	go func() {
		select {
		case <-cancelCh:
			cancel()
		case <-fatalCh:
			cancel()
		case <-reqDone:
			cancel()
		}
	}()

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, recoverableError("http", "http request error: "+err.Error())
	}
	if headersVal, ok := reqObj["headers"]; ok && headersVal != NullValue {
		headers, err := extractHeaders(headersVal)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			if cancelCh != nil {
				select {
				case <-cancelCh:
					return nil, canceledError()
				default:
				}
			}
			if fatalCh != nil {
				select {
				case <-fatalCh:
					return nil, runtimeFatalError(e)
				default:
				}
			}
			return nil, canceledError()
		}
		return nil, recoverableError("http", "http error: "+err.Error())
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, recoverableError("http", "http read error: "+err.Error())
	}
	headerMap := &Map{Pairs: make(map[MapKey]Value)}
	for key, values := range resp.Header {
		headerMap.Pairs[MapKey{Type: STRING, Value: key}] = &String{Value: strings.Join(values, ", ")}
	}
	return &Object{Pairs: map[string]Value{
		"status":  &Integer{Value: int64(resp.StatusCode)},
		"headers": headerMap,
		"body":    &String{Value: string(data)},
	}}, nil
}

func extractHeaders(val Value) (map[string]string, error) {
	switch headers := val.(type) {
	case *Object:
		out := make(map[string]string, len(headers.Pairs))
		for k, v := range headers.Pairs {
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k] = str
		}
		return out, nil
	case *ModuleObject:
		if headers.Env == nil {
			return nil, &RuntimeError{Message: "http headers must be object or map"}
		}
		out := make(map[string]string)
		for k, v := range headers.Env.Snapshot() {
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k] = str
		}
		return out, nil
	case *Map:
		out := make(map[string]string, len(headers.Pairs))
		for k, v := range headers.Pairs {
			if k.Type != STRING && k.Type != CHAR {
				return nil, &RuntimeError{Message: "http headers keys must be strings"}
			}
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k.Value] = str
		}
		return out, nil
	default:
		return nil, &RuntimeError{Message: "http headers must be object or map"}
	}
}

func builtinEncodeJSON(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "encodeJson expects 1 argument"}
	}
	value, err := encodeJSONValue(args[0])
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, &RuntimeError{Message: "encodeJson error: " + err.Error()}
	}
	return &String{Value: string(data)}, nil
}

func encodeJSONValue(value Value) (interface{}, error) {
	switch v := value.(type) {
	case *Null:
		return nil, nil
	case *Boolean:
		return v.Value, nil
	case *Integer:
		return v.Value, nil
	case *Float:
		if math.IsNaN(v.Value) || math.IsInf(v.Value, 0) {
			return nil, &RuntimeError{Message: "encodeJson cannot encode NaN or Inf"}
		}
		return v.Value, nil
	case *String:
		return v.Value, nil
	case *Char:
		return v.Value, nil
	case *Array:
		out := make([]interface{}, 0, len(v.Elements))
		for _, el := range v.Elements {
			enc, err := encodeJSONValue(el)
			if err != nil {
				return nil, err
			}
			out = append(out, enc)
		}
		return out, nil
	case *Object:
		out := make(map[string]interface{}, len(v.Pairs))
		for k, val := range v.Pairs {
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = enc
		}
		return out, nil
	case *ModuleObject:
		if v.Env == nil {
			return map[string]interface{}{}, nil
		}
		out := make(map[string]interface{})
		for k, val := range v.Env.Snapshot() {
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = enc
		}
		return out, nil
	case *Map:
		out := make(map[string]interface{}, len(v.Pairs))
		for k, val := range v.Pairs {
			if k.Type != STRING && k.Type != CHAR {
				return nil, &RuntimeError{Message: "encodeJson only supports map with string keys"}
			}
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k.Value] = enc
		}
		return out, nil
	default:
		return nil, &RuntimeError{Message: "encodeJson does not support " + string(value.Type())}
	}
}

func builtinDecodeJSON(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "decodeJson expects 1 argument"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "decodeJson expects string"}
	}
	decoder := json.NewDecoder(strings.NewReader(str.Value))
	decoder.UseNumber()
	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return nil, recoverableError("decodeJson", "decodeJson error: "+err.Error())
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, recoverableError("decodeJson", "decodeJson expects a single JSON value")
	}
	return decodeJSONValue(data)
}

func builtinThen(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "then expects task and function"}
	}
	task, ok := args[0].(*Task)
	if !ok {
		return nil, &RuntimeError{Message: "then expects task as receiver"}
	}
	// Register observation at chaining time so fail-fast does not treat this task
	// as detached/unobserved while the continuation goroutine starts.
	task.markObserved()
	fn := args[1]
	thenTask := e.newTask(e.currentTask, false)
	thenEval := e.cloneForTask(thenTask)
	go func() {
		val, sig, err := taskAwaitWithCancel(task, thenTask.cancelCh, e.runtime)
		if err != nil {
			thenEval.handleAsyncError(thenTask, err)
			return
		}
		if sig != nil {
			exitProcess("break/continue outside loop")
			return
		}
		res, sig, err := thenEval.applyFunction(fn, []Value{val})
		if err != nil {
			thenEval.handleAsyncError(thenTask, err)
			return
		}
		if sig != nil {
			exitProcess("break/continue outside loop")
			return
		}
		thenTask.complete(res, nil)
	}()
	return thenTask, nil
}

func decodeJSONValue(value interface{}) (Value, error) {
	switch v := value.(type) {
	case nil:
		return NullValue, nil
	case bool:
		return &Boolean{Value: v}, nil
	case string:
		return &String{Value: v}, nil
	case json.Number:
		raw := v.String()
		if strings.ContainsAny(raw, ".eE") {
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
				return nil, recoverableError("decodeJson", "decodeJson invalid float: "+raw)
			}
			return &Float{Value: f}, nil
		}
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, recoverableError("decodeJson", "decodeJson integer overflow: "+raw)
		}
		return &Integer{Value: i}, nil
	case []interface{}:
		out := make([]Value, 0, len(v))
		for _, el := range v {
			val, err := decodeJSONValue(el)
			if err != nil {
				return nil, err
			}
			out = append(out, val)
		}
		return &Array{Elements: out}, nil
	case map[string]interface{}:
		out := make(map[string]Value, len(v))
		for k, val := range v {
			decoded, err := decodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = decoded
		}
		return &Object{Pairs: out}, nil
	default:
		return nil, recoverableError("decodeJson", "decodeJson unsupported value")
	}
}

func builtinSend(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "send expects channel and value"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "send expects channel"}
	}
	if ch.Closed {
		return nil, &RuntimeError{Message: "send on closed channel"}
	}

	fatalCh := runtimeFatalSignal(e)
	cancelCh := (<-chan struct{})(nil)
	if e != nil && e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}

	if cancelCh == nil && fatalCh == nil {
		ch.Ch <- args[1]
		return UnitValue, nil
	}
	select {
	case ch.Ch <- args[1]:
		return UnitValue, nil
	case <-cancelCh:
		return nil, canceledError()
	case <-fatalCh:
		return nil, runtimeFatalError(e)
	}
}

func builtinRecv(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "recv expects channel"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "recv expects channel"}
	}
	var val Value
	var okRecv bool
	fatalCh := runtimeFatalSignal(e)
	cancelCh := (<-chan struct{})(nil)
	if e != nil && e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}
	if cancelCh == nil && fatalCh == nil {
		val, okRecv = <-ch.Ch
	} else {
		select {
		case val, okRecv = <-ch.Ch:
		case <-cancelCh:
			return nil, canceledError()
		case <-fatalCh:
			return nil, runtimeFatalError(e)
		}
	}
	if !okRecv {
		return &Array{Elements: []Value{NullValue, &Boolean{Value: true}}}, nil
	}
	return &Array{Elements: []Value{val, &Boolean{Value: false}}}, nil
}

func builtinDone(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "done expects channel"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "done expects channel"}
	}
	ch.Close()
	return UnitValue, nil
}
