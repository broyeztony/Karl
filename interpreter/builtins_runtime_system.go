package interpreter

import "fmt"

func registerRuntimeSystemBuiltins() {
	builtins["argv"] = &Builtin{Name: "argv", Fn: builtinArgv}
	builtins["programPath"] = &Builtin{Name: "programPath", Fn: builtinProgramPath}
	builtins["environ"] = &Builtin{Name: "environ", Fn: builtinEnviron}
	builtins["env"] = &Builtin{Name: "env", Fn: builtinEnv}
	builtins["readLine"] = &Builtin{Name: "readLine", Fn: builtinReadLine}
}

func builtinArgv(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "argv expects no arguments"}
	}
	argv := runtimeProgramArgs(e)
	out := make([]Value, 0, len(argv))
	for _, arg := range argv {
		out = append(out, &String{Value: arg})
	}
	return &Array{Elements: out}, nil
}

func builtinProgramPath(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "programPath expects no arguments"}
	}
	path, ok := runtimeProgramPath(e)
	if !ok {
		return NullValue, nil
	}
	return &String{Value: path}, nil
}

func builtinEnviron(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "environ expects no arguments"}
	}
	environ := runtimeEnviron(e)
	out := make([]Value, 0, len(environ))
	for _, entry := range environ {
		out = append(out, &String{Value: entry})
	}
	return &Array{Elements: out}, nil
}

func builtinEnv(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "env expects 1 argument"}
	}
	name, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "env expects string argument"}
	}
	value, found := runtimeLookupEnv(e, name.Value)
	if !found {
		return NullValue, nil
	}
	return &String{Value: value}, nil
}

func builtinReadLine(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "readLine expects no arguments"}
	}
	line, ok, err := runtimeReadLine(e)
	if err != nil {
		return nil, recoverableError("readLine", fmt.Sprintf("readLine error: %v", err))
	}
	if !ok {
		return NullValue, nil
	}
	return &String{Value: line}, nil
}

func runtimeProgramArgs(e *Evaluator) []string {
	if e == nil || e.runtime == nil {
		return nil
	}
	return e.runtime.snapshotProgramArgs()
}

func runtimeProgramPath(e *Evaluator) (string, bool) {
	if e == nil || e.runtime == nil {
		return "", false
	}
	return e.runtime.getProgramPath()
}

func runtimeEnviron(e *Evaluator) []string {
	if e == nil || e.runtime == nil {
		return nil
	}
	return e.runtime.snapshotEnviron()
}

func runtimeLookupEnv(e *Evaluator, name string) (string, bool) {
	if e == nil || e.runtime == nil {
		return "", false
	}
	return e.runtime.lookupEnv(name)
}

func runtimeReadLine(e *Evaluator) (string, bool, error) {
	if e == nil || e.runtime == nil {
		return "", false, nil
	}
	return e.runtime.readLine()
}
