//go:build !js

package interpreter

import (
	"io"
	"os"
)

func registerStreamBuiltins() {
	builtins["reader"] = &Builtin{Name: "reader", Fn: builtinReader}
	builtins["writer"] = &Builtin{Name: "writer", Fn: builtinWriter}
	builtins["copy"] = &Builtin{Name: "copy", Fn: builtinCopy}
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

func builtinCopy(_ *Evaluator, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, &RuntimeError{Message: "copy expects (srcReader, dstWriter, opts?)"}
	}
	src, ok := args[0].(*StreamReader)
	if !ok {
		return nil, &RuntimeError{Message: "copy src must be stream reader"}
	}
	dst, ok := args[1].(*StreamWriter)
	if !ok {
		return nil, &RuntimeError{Message: "copy dst must be stream writer"}
	}

	bufferSize := int64(defaultStreamCopyBuffer)
	if len(args) == 3 && !Equivalent(args[2], NullValue) {
		pairs, ok := objectPairs(args[2])
		if !ok {
			return nil, &RuntimeError{Message: "copy opts must be object"}
		}
		if sizeVal, ok := pairs["bufferSize"]; ok && !Equivalent(sizeVal, NullValue) {
			size, ok := sizeVal.(*Integer)
			if !ok {
				return nil, &RuntimeError{Message: "copy bufferSize must be integer"}
			}
			if size.Value <= 0 {
				return nil, &RuntimeError{Message: "copy bufferSize must be > 0"}
			}
			bufferSize = size.Value
		}
	}

	src.mu.Lock()
	srcClosed := src.closed
	srcReader := src.reader
	src.mu.Unlock()
	dst.mu.Lock()
	dstClosed := dst.closed
	dstWriter := dst.writer
	dst.mu.Unlock()

	if dstClosed || dstWriter == nil {
		return nil, recoverableError("stream_write", "stream write error: stream writer unavailable")
	}
	if srcClosed || srcReader == nil {
		return &Object{Pairs: map[string]Value{
			"bytes":  &Integer{Value: 0},
			"chunks": &Integer{Value: 0},
		}}, nil
	}

	if total, chunks, used, err := streamCopyTryFastPath(srcReader, dstWriter, int(bufferSize)); used {
		if err != nil {
			return nil, recoverableError("stream_write", "stream write error: "+err.Error())
		}
		if total > 0 && chunks == 0 {
			chunks = 1
		}
		return &Object{Pairs: map[string]Value{
			"bytes":  &Integer{Value: total},
			"chunks": &Integer{Value: chunks},
		}}, nil
	}

	buf := make([]byte, int(bufferSize))
	var total int64
	var chunks int64
	for {
		n, readErr := readIntoBuffer(src, buf)
		if n > 0 {
			written, writeErr := dst.WriteChunk(buf[:n])
			total += int64(written)
			chunks++
			if writeErr != nil {
				return nil, recoverableError("stream_write", "stream write error: "+writeErr.Error())
			}
			if written != n {
				return nil, recoverableError("stream_write", "stream write error: short write")
			}
		}
		if readErr == nil {
			continue
		}
		if readErr == io.EOF {
			break
		}
		return nil, recoverableError("stream_read", "stream read error: "+readErr.Error())
	}

	return &Object{Pairs: map[string]Value{
		"bytes":  &Integer{Value: total},
		"chunks": &Integer{Value: chunks},
	}}, nil
}

func readIntoBuffer(src *StreamReader, buf []byte) (int, error) {
	if src == nil {
		return 0, io.EOF
	}
	src.mu.Lock()
	closed := src.closed
	reader := src.reader
	src.mu.Unlock()
	if closed || reader == nil {
		return 0, io.EOF
	}
	n, err := reader.Read(buf)
	if n > 0 {
		return n, nil
	}
	if err == nil {
		return 0, nil
	}
	if streamReadEnded(err) {
		return 0, io.EOF
	}
	return 0, err
}
