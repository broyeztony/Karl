package interpreter

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var builtins = map[string]*Builtin{}

func RegisterBuiltins() {
	builtins["exit"] = &Builtin{Name: "exit", Fn: builtinExit}
	builtins["fail"] = &Builtin{Name: "fail", Fn: builtinFail}
	builtins["rendezvous"] = &Builtin{Name: "rendezvous", Fn: builtinChannel}
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

func builtinSleep(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sleep expects 1 argument"}
	}
	ms, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "sleep expects integer milliseconds"}
	}
	time.Sleep(time.Duration(ms.Value) * time.Millisecond)
	return UnitValue, nil
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

func builtinReadFile(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "readFile expects path"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "readFile expects string path"}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, recoverableError("readFile", "readFile error: "+err.Error())
	}
	return &String{Value: string(data)}, nil
}

func builtinWriteFile(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "writeFile expects path and data"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "writeFile expects string path"}
	}
	data, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "writeFile expects string data"}
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		return nil, recoverableError("writeFile", "writeFile error: "+err.Error())
	}
	return UnitValue, nil
}

func builtinAppendFile(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "appendFile expects path and data"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "appendFile expects string path"}
	}
	data, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "appendFile expects string data"}
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, recoverableError("appendFile", "appendFile error: "+err.Error())
	}
	defer file.Close()
	if _, err := file.WriteString(data); err != nil {
		return nil, recoverableError("appendFile", "appendFile error: "+err.Error())
	}
	return UnitValue, nil
}

func builtinDeleteFile(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "deleteFile expects path"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "deleteFile expects string path"}
	}
	if err := os.Remove(path); err != nil {
		return nil, recoverableError("deleteFile", "deleteFile error: "+err.Error())
	}
	return UnitValue, nil
}

func builtinExists(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "exists expects path"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "exists expects string path"}
	}
	_, err := os.Stat(path)
	if err == nil {
		return &Boolean{Value: true}, nil
	}
	if os.IsNotExist(err) {
		return &Boolean{Value: false}, nil
	}
	return nil, recoverableError("exists", "exists error: "+err.Error())
}

func builtinListDir(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "listDir expects path"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "listDir expects string path"}
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, recoverableError("listDir", "listDir error: "+err.Error())
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	out := make([]Value, 0, len(names))
	for _, name := range names {
		out = append(out, &String{Value: name})
	}
	return &Array{Elements: out}, nil
}

func builtinHTTP(_ *Evaluator, args []Value) (Value, error) {
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
	req, err := http.NewRequest(method, urlStr, body)
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
	fn := args[1]
	thenTask := newTask()
	go func() {
		val, sig, err := taskAwait(task)
		if err != nil {
			exitProcess(e.formatError(err))
			return
		}
		if sig != nil {
			exitProcess("break/continue outside loop")
			return
		}
		res, sig, err := e.applyFunction(fn, []Value{val})
		if err != nil {
			exitProcess(e.formatError(err))
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

func numberArg(val Value) (float64, bool, bool) {
	switch v := val.(type) {
	case *Integer:
		return float64(v.Value), true, true
	case *Float:
		return v.Value, false, true
	default:
		return 0, false, false
	}
}

func builtinAbs(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "abs expects 1 argument"}
	}
	switch v := args[0].(type) {
	case *Integer:
		if v.Value == math.MinInt64 {
			return nil, &RuntimeError{Message: "abs overflow"}
		}
		if v.Value < 0 {
			return &Integer{Value: -v.Value}, nil
		}
		return v, nil
	case *Float:
		return &Float{Value: math.Abs(v.Value)}, nil
	default:
		return nil, &RuntimeError{Message: "abs expects number"}
	}
}

func builtinSqrt(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sqrt expects 1 argument"}
	}
	val, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "sqrt expects number"}
	}
	if val < 0 {
		return nil, &RuntimeError{Message: "sqrt expects non-negative number"}
	}
	return &Float{Value: math.Sqrt(val)}, nil
}

func builtinPow(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "pow expects base and exponent"}
	}
	base, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "pow expects numeric base"}
	}
	exp, _, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "pow expects numeric exponent"}
	}
	result := math.Pow(base, exp)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return nil, &RuntimeError{Message: "pow result not finite"}
	}
	return &Float{Value: result}, nil
}

func builtinSin(_ *Evaluator, args []Value) (Value, error) {
	return unaryMath(args, "sin", math.Sin)
}

func builtinCos(_ *Evaluator, args []Value) (Value, error) {
	return unaryMath(args, "cos", math.Cos)
}

func builtinTan(_ *Evaluator, args []Value) (Value, error) {
	return unaryMath(args, "tan", math.Tan)
}

func unaryMath(args []Value, name string, fn func(float64) float64) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: name + " expects 1 argument"}
	}
	val, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: name + " expects number"}
	}
	result := fn(val)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return nil, &RuntimeError{Message: name + " result not finite"}
	}
	return &Float{Value: result}, nil
}

func builtinFloor(_ *Evaluator, args []Value) (Value, error) {
	return integralMath(args, "floor", math.Floor)
}

func builtinCeil(_ *Evaluator, args []Value) (Value, error) {
	return integralMath(args, "ceil", math.Ceil)
}

func integralMath(args []Value, name string, fn func(float64) float64) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: name + " expects 1 argument"}
	}
	val, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: name + " expects number"}
	}
	result := fn(val)
	if result > float64(math.MaxInt64) || result < float64(math.MinInt64) {
		return nil, &RuntimeError{Message: name + " overflow"}
	}
	return &Integer{Value: int64(result)}, nil
}

func builtinMin(_ *Evaluator, args []Value) (Value, error) {
	return minMax(args, "min", false)
}

func builtinMax(_ *Evaluator, args []Value) (Value, error) {
	return minMax(args, "max", true)
}

func minMax(args []Value, name string, takeMax bool) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: name + " expects 2 arguments"}
	}
	left, leftInt, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: name + " expects number"}
	}
	right, rightInt, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: name + " expects number"}
	}
	var out float64
	if takeMax {
		out = math.Max(left, right)
	} else {
		out = math.Min(left, right)
	}
	if leftInt && rightInt {
		return &Integer{Value: int64(out)}, nil
	}
	return &Float{Value: out}, nil
}

func builtinClamp(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "clamp expects value, min, max"}
	}
	val, valInt, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "clamp expects numeric value"}
	}
	min, minInt, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "clamp expects numeric min"}
	}
	max, maxInt, ok := numberArg(args[2])
	if !ok {
		return nil, &RuntimeError{Message: "clamp expects numeric max"}
	}
	if min > max {
		return nil, &RuntimeError{Message: "clamp expects min <= max"}
	}
	out := val
	if out < min {
		out = min
	}
	if out > max {
		out = max
	}
	if valInt && minInt && maxInt {
		return &Integer{Value: int64(out)}, nil
	}
	return &Float{Value: out}, nil
}

func stringArg(val Value) (string, bool) {
	switch v := val.(type) {
	case *String:
		return v.Value, true
	case *Char:
		return v.Value, true
	default:
		return "", false
	}
}

func builtinSplit(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "split expects string and separator"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "split expects string as first argument"}
	}
	sep, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "split expects string separator"}
	}
	if sep == "" {
		runes := []rune(str.Value)
		out := make([]Value, 0, len(runes))
		for _, r := range runes {
			out = append(out, &String{Value: string(r)})
		}
		return &Array{Elements: out}, nil
	}
	parts := strings.Split(str.Value, sep)
	out := make([]Value, 0, len(parts))
	for _, part := range parts {
		out = append(out, &String{Value: part})
	}
	return &Array{Elements: out}, nil
}

func builtinChars(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "chars expects string"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "chars expects string as first argument"}
	}
	runes := []rune(str.Value)
	out := make([]Value, 0, len(runes))
	for _, r := range runes {
		out = append(out, &Char{Value: string(r)})
	}
	return &Array{Elements: out}, nil
}

func builtinTrim(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "trim expects string"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "trim expects string as first argument"}
	}
	return &String{Value: strings.TrimSpace(str.Value)}, nil
}

func builtinToLower(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "toLower expects string"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "toLower expects string as first argument"}
	}
	return &String{Value: strings.ToLower(str.Value)}, nil
}

func builtinToUpper(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "toUpper expects string"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "toUpper expects string as first argument"}
	}
	return &String{Value: strings.ToUpper(str.Value)}, nil
}

func builtinContains(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "contains expects string and substring"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "contains expects string as first argument"}
	}
	sub, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "contains expects string substring"}
	}
	return &Boolean{Value: strings.Contains(str.Value, sub)}, nil
}

func builtinStartsWith(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "startsWith expects string and prefix"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "startsWith expects string as first argument"}
	}
	prefix, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "startsWith expects string prefix"}
	}
	return &Boolean{Value: strings.HasPrefix(str.Value, prefix)}, nil
}

func builtinEndsWith(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "endsWith expects string and suffix"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "endsWith expects string as first argument"}
	}
	suffix, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "endsWith expects string suffix"}
	}
	return &Boolean{Value: strings.HasSuffix(str.Value, suffix)}, nil
}

func builtinReplace(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "replace expects string, old, new"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "replace expects string as first argument"}
	}
	oldVal, ok := stringArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "replace expects string old value"}
	}
	newVal, ok := stringArg(args[2])
	if !ok {
		return nil, &RuntimeError{Message: "replace expects string new value"}
	}
	return &String{Value: strings.ReplaceAll(str.Value, oldVal, newVal)}, nil
}

func builtinMap(e *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return &Map{Pairs: make(map[MapKey]Value)}, nil
	}
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "map expects no arguments or array and function"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "map expects array as first argument"}
	}
	fn := args[1]
	out := make([]Value, 0, len(arr.Elements))
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{el})
		if err != nil {
			return nil, err
		}
		out = append(out, val)
	}
	return &Array{Elements: out}, nil
}

func builtinMapGet(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "get expects map and key"}
	}
	m, ok := args[0].(*Map)
	if !ok {
		return nil, &RuntimeError{Message: "get expects map as first argument"}
	}
	key, err := mapKeyForValue(args[1])
	if err != nil {
		return nil, err
	}
	val, ok := m.Pairs[key]
	if !ok {
		return NullValue, nil
	}
	return val, nil
}

func builtinMapSet(_ *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return &Set{Elements: make(map[MapKey]struct{})}, nil
	}
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "set expects no arguments or map, key, value"}
	}
	m, ok := args[0].(*Map)
	if !ok {
		return nil, &RuntimeError{Message: "set expects map as first argument"}
	}
	key, err := mapKeyForValue(args[1])
	if err != nil {
		return nil, err
	}
	m.Pairs[key] = args[2]
	return m, nil
}

func builtinSetAdd(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "add expects set and value"}
	}
	s, ok := args[0].(*Set)
	if !ok {
		return nil, &RuntimeError{Message: "add expects set as first argument"}
	}
	key, err := setKeyForValue(args[1])
	if err != nil {
		return nil, err
	}
	s.Elements[key] = struct{}{}
	return s, nil
}

func builtinMapHas(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "has expects map or set and key"}
	}
	switch target := args[0].(type) {
	case *Map:
		key, err := mapKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		_, ok := target.Pairs[key]
		return &Boolean{Value: ok}, nil
	case *Set:
		key, err := setKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		_, ok := target.Elements[key]
		return &Boolean{Value: ok}, nil
	default:
		return nil, &RuntimeError{Message: "has expects map or set as first argument"}
	}
}

func builtinMapDelete(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "delete expects map or set and key"}
	}
	switch target := args[0].(type) {
	case *Map:
		key, err := mapKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		_, ok := target.Pairs[key]
		if ok {
			delete(target.Pairs, key)
		}
		return &Boolean{Value: ok}, nil
	case *Set:
		key, err := setKeyForValue(args[1])
		if err != nil {
			return nil, err
		}
		_, ok := target.Elements[key]
		if ok {
			delete(target.Elements, key)
		}
		return &Boolean{Value: ok}, nil
	default:
		return nil, &RuntimeError{Message: "delete expects map or set as first argument"}
	}
}

func builtinMapKeys(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "keys expects map"}
	}
	m, ok := args[0].(*Map)
	if !ok {
		return nil, &RuntimeError{Message: "keys expects map as first argument"}
	}
	out := make([]Value, 0, len(m.Pairs))
	for key := range m.Pairs {
		out = append(out, mapKeyToValue(key))
	}
	return &Array{Elements: out}, nil
}

func builtinMapValues(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "values expects map or set"}
	}
	switch target := args[0].(type) {
	case *Map:
		out := make([]Value, 0, len(target.Pairs))
		for _, val := range target.Pairs {
			out = append(out, val)
		}
		return &Array{Elements: out}, nil
	case *Set:
		out := make([]Value, 0, len(target.Elements))
		for key := range target.Elements {
			out = append(out, mapKeyToValue(key))
		}
		return &Array{Elements: out}, nil
	default:
		return nil, &RuntimeError{Message: "values expects map or set as first argument"}
	}
}

func builtinSort(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "sort expects array and comparator"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "sort expects array as first argument"}
	}
	cmp := args[1]
	out := append([]Value{}, arr.Elements...)
	var cmpErr error
	sort.Slice(out, func(i, j int) bool {
		if cmpErr != nil {
			return false
		}
		val, _, err := e.applyFunction(cmp, []Value{out[i], out[j]})
		if err != nil {
			cmpErr = err
			return false
		}
		num, _, ok := numberArg(val)
		if !ok {
			cmpErr = &RuntimeError{Message: "sort comparator must return number"}
			return false
		}
		return num < 0
	})
	if cmpErr != nil {
		return nil, cmpErr
	}
	return &Array{Elements: out}, nil
}

func builtinFilter(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "filter expects array and function"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "filter expects array"}
	}
	fn := args[1]
	out := []Value{}
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{el})
		if err != nil {
			return nil, err
		}
		b, ok := val.(*Boolean)
		if !ok {
			return nil, &RuntimeError{Message: "filter predicate must return bool"}
		}
		if b.Value {
			out = append(out, el)
		}
	}
	return &Array{Elements: out}, nil
}

func builtinReduce(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 3 {
		return nil, &RuntimeError{Message: "reduce expects array, function, initial"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "reduce expects array"}
	}
	fn := args[1]
	acc := args[2]
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{acc, el})
		if err != nil {
			return nil, err
		}
		acc = val
	}
	return acc, nil
}

func builtinSum(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sum expects array"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "sum expects array"}
	}
	var total float64
	allInts := true
	for _, el := range arr.Elements {
		switch v := el.(type) {
		case *Integer:
			total += float64(v.Value)
		case *Float:
			allInts = false
			total += v.Value
		default:
			return nil, &RuntimeError{Message: "sum expects numeric array"}
		}
	}
	if allInts {
		return &Integer{Value: int64(total)}, nil
	}
	return &Float{Value: total}, nil
}

func builtinFind(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "find expects array and function"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "find expects array"}
	}
	fn := args[1]
	for _, el := range arr.Elements {
		val, _, err := e.applyFunction(fn, []Value{el})
		if err != nil {
			return nil, err
		}
		b, ok := val.(*Boolean)
		if !ok {
			return nil, &RuntimeError{Message: "find predicate must return bool"}
		}
		if b.Value {
			return el, nil
		}
	}
	return NullValue, nil
}

func builtinSend(_ *Evaluator, args []Value) (Value, error) {
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
	ch.Ch <- args[1]
	return UnitValue, nil
}

func builtinRecv(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "recv expects channel"}
	}
	ch, ok := args[0].(*Channel)
	if !ok {
		return nil, &RuntimeError{Message: "recv expects channel"}
	}
	val, ok := <-ch.Ch
	if !ok {
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
