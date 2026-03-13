package interpreter

import (
	"encoding/json"
	"io"
	"strings"
)

func registerJSONBuiltins() {
	builtins["toJson"] = &Builtin{Name: "toJson", Fn: builtinToJSON}
	builtins["fromJson"] = &Builtin{Name: "fromJson", Fn: builtinFromJSON}
}

func builtinToJSON(_ *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return newStreamToJSONStage(), nil
	}
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "toJson expects 1 argument or no arguments for stream stage"}
	}
	value, err := encodeJSONValue(args[0])
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, &RuntimeError{Message: "toJson error: " + err.Error()}
	}
	return &String{Value: string(data)}, nil
}

func builtinFromJSON(_ *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return newStreamFromJSONStage(), nil
	}
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "fromJson expects 1 argument or no arguments for stream stage"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "fromJson expects string"}
	}
	decoder := json.NewDecoder(strings.NewReader(str.Value))
	decoder.UseNumber()
	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return nil, recoverableError("fromJson", "fromJson error: "+err.Error())
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, recoverableError("fromJson", "fromJson expects a single JSON value")
	}
	return decodeJSONValue(data)
}
