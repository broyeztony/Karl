package interpreter

import "unicode/utf8"

func registerEncodingBuiltins() {
	builtins["encodeUtf8"] = &Builtin{Name: "encodeUtf8", Fn: builtinEncodeUtf8}
	builtins["decodeUtf8"] = &Builtin{Name: "decodeUtf8", Fn: builtinDecodeUtf8}
}

func builtinEncodeUtf8(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "encodeUtf8 expects 1 argument"}
	}
	text, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "encodeUtf8 expects string"}
	}
	return &Bytes{Value: []byte(text)}, nil
}

func builtinDecodeUtf8(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "decodeUtf8 expects 1 argument"}
	}
	payload, ok := bytesArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "decodeUtf8 expects bytes"}
	}
	if !utf8.Valid(payload) {
		return nil, recoverableError("utf8_decode", "decodeUtf8 error: invalid UTF-8 bytes")
	}
	return &String{Value: string(payload)}, nil
}
