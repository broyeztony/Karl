package interpreter

var builtins = map[string]*Builtin{}

func RegisterBuiltins() {
	builtins = map[string]*Builtin{}
	registerRuntimeBuiltins()
	registerFSBuiltins()
	registerHTTPBuiltins()
	registerJSONBuiltins()
	registerSQLBuiltins()
	registerCryptoBuiltins()
	registerUUIDBuiltins()
	registerTimeBuiltins()
	registerSignalBuiltins()
	registerAsyncBuiltins()
	registerStringBuiltins()
	registerCollectionBuiltins()
	registerListBuiltins()
	registerMathBuiltins()
}

func getBuiltin(name string) *Builtin {
	if b, ok := builtins[name]; ok {
		return b
	}
	return nil
}

func NewBaseEnvironment() *Environment {
	RegisterBuiltins()
	env := NewEnvironment()
	for name, builtin := range builtins {
		env.Define(name, builtin)
	}
	return env
}

func bindReceiver(fn BuiltinFunction, receiver Value) BuiltinFunction {
	return func(e *Evaluator, args []Value) (Value, error) {
		return fn(e, append([]Value{receiver}, args...))
	}
}

func recoverableError(kind string, msg string) *RecoverableError {
	return &RecoverableError{Kind: kind, Message: msg}
}
