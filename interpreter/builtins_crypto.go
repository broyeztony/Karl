package interpreter

import (
	"crypto/sha256"
	"encoding/hex"
)

func registerCryptoBuiltins() {
	builtins["sha256"] = &Builtin{Name: "sha256", Fn: builtinSHA256}
}

func builtinSHA256(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sha256 expects 1 argument"}
	}
	input, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "sha256 expects string"}
	}
	hash := sha256.Sum256([]byte(input))
	return &String{Value: hex.EncodeToString(hash[:])}, nil
}
