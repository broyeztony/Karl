package interpreter

import (
	"fmt"
	"strings"
	"time"
)

func builtinLog(e *Evaluator, args []Value) (Value, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatLogValue(arg)
	}
	if err := writeLogLine(e, strings.Join(parts, " ")); err != nil {
		return nil, &RuntimeError{Message: "log write failed: " + err.Error()}
	}
	return UnitValue, nil
}

func builtinLogt(e *Evaluator, args []Value) (Value, error) {
	prefix := "[" + time.Now().UTC().Format(time.RFC3339) + "]"
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, prefix)
	for _, arg := range args {
		parts = append(parts, formatLogValue(arg))
	}
	if err := writeLogLine(e, strings.Join(parts, " ")); err != nil {
		return nil, &RuntimeError{Message: "log write failed: " + err.Error()}
	}
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

func writeLogLine(e *Evaluator, line string) error {
	if e == nil || e.runtime == nil {
		fmt.Println(line)
		return nil
	}
	return e.runtime.writeOutputLine(line)
}
