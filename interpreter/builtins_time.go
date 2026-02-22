package interpreter

import (
	"time"
)

func registerTimeBuiltins() {
	builtins["timeParseRFC3339"] = &Builtin{Name: "timeParseRFC3339", Fn: builtinTimeParseRFC3339}
	builtins["timeFormatRFC3339"] = &Builtin{Name: "timeFormatRFC3339", Fn: builtinTimeFormatRFC3339}
	builtins["timeAdd"] = &Builtin{Name: "timeAdd", Fn: builtinTimeAdd}
	builtins["timeDiff"] = &Builtin{Name: "timeDiff", Fn: builtinTimeDiff}
}

func builtinTimeParseRFC3339(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "timeParseRFC3339 expects 1 argument"}
	}
	text, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "timeParseRFC3339 expects string"}
	}
	ts, err := time.Parse(time.RFC3339, text)
	if err != nil {
		return nil, recoverableError("timeParseRFC3339", "timeParseRFC3339 error: "+err.Error())
	}
	return &Integer{Value: ts.UTC().UnixNano() / int64(time.Millisecond)}, nil
}

func builtinTimeFormatRFC3339(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "timeFormatRFC3339 expects 1 argument"}
	}
	ms, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "timeFormatRFC3339 expects integer milliseconds"}
	}
	ts := time.Unix(0, ms.Value*int64(time.Millisecond)).UTC()
	return &String{Value: ts.Format(time.RFC3339)}, nil
}

func builtinTimeAdd(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "timeAdd expects 2 arguments"}
	}
	base, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "timeAdd expects integer milliseconds"}
	}
	delta, ok := args[1].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "timeAdd expects integer delta milliseconds"}
	}
	return &Integer{Value: base.Value + delta.Value}, nil
}

func builtinTimeDiff(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "timeDiff expects 2 arguments"}
	}
	a, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "timeDiff expects integer first argument"}
	}
	b, ok := args[1].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "timeDiff expects integer second argument"}
	}
	return &Integer{Value: a.Value - b.Value}, nil
}
