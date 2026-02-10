package interpreter

import (
	"os"
	"sort"
)

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
