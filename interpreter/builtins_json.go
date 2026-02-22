package interpreter

import (
	"encoding/json"
	"io"
	"strings"
)

func registerJSONBuiltins() {
	builtins["jsonEncode"] = &Builtin{Name: "jsonEncode", Fn: builtinEncodeJSON}
	builtins["jsonDecode"] = &Builtin{Name: "jsonDecode", Fn: builtinDecodeJSON}
}

func builtinEncodeJSON(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "jsonEncode expects 1 argument"}
	}
	value, err := encodeJSONValue(args[0])
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, &RuntimeError{Message: "jsonEncode error: " + err.Error()}
	}
	return &String{Value: string(data)}, nil
}

func builtinDecodeJSON(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "jsonDecode expects 1 argument"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "jsonDecode expects string"}
	}
	decoder := json.NewDecoder(strings.NewReader(str.Value))
	decoder.UseNumber()
	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return nil, recoverableError("jsonDecode", "jsonDecode error: "+err.Error())
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, recoverableError("jsonDecode", "jsonDecode expects a single JSON value")
	}
	return decodeJSONValue(data)
}
