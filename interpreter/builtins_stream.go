//go:build !js

package interpreter

import (
	"os"
)

func registerStreamBuiltins() {
	builtins["reader"] = &Builtin{Name: "reader", Fn: builtinReader}
	builtins["writer"] = &Builtin{Name: "writer", Fn: builtinWriter}
	registerStreamPipelineBuiltins()
}

func builtinReader(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "reader expects (path, opts?)"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "reader path must be string"}
	}

	mode := streamTypeBytes
	if len(args) == 2 && !Equivalent(args[1], NullValue) {
		pairs, ok := objectPairs(args[1])
		if !ok {
			return nil, &RuntimeError{Message: "reader opts must be object"}
		}
		if typeVal, ok := pairs["type"]; ok && !Equivalent(typeVal, NullValue) {
			parsed, err := parseStreamType(typeVal, "type")
			if err != nil {
				return nil, err
			}
			mode = parsed
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, recoverableError("stream_open", "reader open error: "+err.Error())
	}
	return &StreamReader{
		reader: file,
		closer: file,
		mode:   mode,
	}, nil
}

func builtinWriter(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "writer expects (path, opts?)"}
	}
	path, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "writer path must be string"}
	}

	mode := streamTypeBytes
	appendMode := false
	if len(args) == 2 && !Equivalent(args[1], NullValue) {
		pairs, ok := objectPairs(args[1])
		if !ok {
			return nil, &RuntimeError{Message: "writer opts must be object"}
		}
		if typeVal, ok := pairs["type"]; ok && !Equivalent(typeVal, NullValue) {
			parsed, err := parseStreamType(typeVal, "type")
			if err != nil {
				return nil, err
			}
			mode = parsed
		}
		if appendVal, ok := pairs["append"]; ok && !Equivalent(appendVal, NullValue) {
			parsed, ok := appendVal.(*Boolean)
			if !ok {
				return nil, &RuntimeError{Message: "writer append must be bool"}
			}
			appendMode = parsed.Value
		}
	}

	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return nil, recoverableError("stream_open", "writer open error: "+err.Error())
	}
	return &StreamWriter{
		writer: file,
		closer: file,
		mode:   mode,
	}, nil
}
