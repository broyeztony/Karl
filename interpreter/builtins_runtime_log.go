package interpreter

import (
	"fmt"
	"strings"
	"time"
)

func builtinLog(_ *Evaluator, args []Value) (Value, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatLogValue(arg)
	}
	fmt.Println(strings.Join(parts, " "))
	return UnitValue, nil
}

func builtinLogt(_ *Evaluator, args []Value) (Value, error) {
	prefix := "[" + time.Now().UTC().Format(time.RFC3339) + "]"
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, prefix)
	for _, arg := range args {
		parts = append(parts, formatLogValue(arg))
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
