package interpreter

import (
	"os"
	"sort"
)

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
