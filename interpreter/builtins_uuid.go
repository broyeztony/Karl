package interpreter

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

func registerUUIDBuiltins() {
	builtins["uuidNew"] = &Builtin{Name: "uuidNew", Fn: builtinUUIDNew}
	builtins["uuidValid"] = &Builtin{Name: "uuidValid", Fn: builtinUUIDValid}
	builtins["uuidParse"] = &Builtin{Name: "uuidParse", Fn: builtinUUIDParse}
}

func builtinUUIDNew(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "uuidNew expects no arguments"}
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return nil, recoverableError("uuidNew", "uuidNew error: "+err.Error())
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	return &String{Value: formatUUID(raw)}, nil
}

func builtinUUIDValid(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "uuidValid expects 1 argument"}
	}
	input, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "uuidValid expects string"}
	}
	_, err := parseUUID(input)
	return &Boolean{Value: err == nil}, nil
}

func builtinUUIDParse(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "uuidParse expects 1 argument"}
	}
	input, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "uuidParse expects string"}
	}
	raw, err := parseUUID(input)
	if err != nil {
		return nil, recoverableError("uuidParse", "uuidParse error: "+err.Error())
	}
	return &String{Value: formatUUID(raw)}, nil
}

func parseUUID(input string) ([16]byte, error) {
	var out [16]byte
	if len(input) != 36 {
		return out, fmt.Errorf("invalid UUID length")
	}
	for _, idx := range []int{8, 13, 18, 23} {
		if input[idx] != '-' {
			return out, fmt.Errorf("invalid UUID format")
		}
	}
	clean := strings.ToLower(strings.ReplaceAll(input, "-", ""))
	if len(clean) != 32 {
		return out, fmt.Errorf("invalid UUID format")
	}
	decoded, err := hex.DecodeString(clean)
	if err != nil {
		return out, fmt.Errorf("invalid UUID encoding")
	}
	copy(out[:], decoded)
	return out, nil
}

func formatUUID(raw [16]byte) string {
	hexText := hex.EncodeToString(raw[:])
	return hexText[0:8] + "-" + hexText[8:12] + "-" + hexText[12:16] + "-" + hexText[16:20] + "-" + hexText[20:32]
}
