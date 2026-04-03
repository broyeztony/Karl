package interpreter

import "unicode/utf8"

func registerEncodingBuiltins() {
	builtins["toUtf8"] = &Builtin{Name: "toUtf8", Fn: builtinToUtf8}
	builtins["fromUtf8"] = &Builtin{Name: "fromUtf8", Fn: builtinFromUtf8}
	builtins["toBase58"] = &Builtin{Name: "toBase58", Fn: builtinToBase58}
	builtins["fromBase58"] = &Builtin{Name: "fromBase58", Fn: builtinFromBase58}
	builtins["bytesJoin"] = &Builtin{Name: "bytesJoin", Fn: builtinBytesJoin}
}

func builtinToUtf8(_ *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return newStreamToUTF8Stage(), nil
	}
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "toUtf8 expects 1 argument or no arguments for stream stage"}
	}
	text, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "toUtf8 expects string"}
	}
	return &Bytes{Value: []byte(text)}, nil
}

func builtinFromUtf8(_ *Evaluator, args []Value) (Value, error) {
	if len(args) == 0 {
		return newStreamFromUTF8Stage(), nil
	}
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "fromUtf8 expects 1 argument or no arguments for stream stage"}
	}
	payload, ok := bytesArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "fromUtf8 expects bytes"}
	}
	if !utf8.Valid(payload) {
		return nil, recoverableError("utf8_decode", "fromUtf8 error: invalid UTF-8 bytes")
	}
	return &String{Value: string(payload)}, nil
}

func builtinBytesJoin(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "bytesJoin expects 1 argument"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "bytesJoin expects array of bytes"}
	}
	total := 0
	for _, el := range arr.Elements {
		chunk, ok := el.(*Bytes)
		if !ok {
			return nil, &RuntimeError{Message: "bytesJoin expects array of bytes"}
		}
		total += len(chunk.Value)
	}
	out := make([]byte, 0, total)
	for _, el := range arr.Elements {
		chunk := el.(*Bytes)
		out = append(out, chunk.Value...)
	}
	return &Bytes{Value: out}, nil
}
