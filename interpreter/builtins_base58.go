package interpreter

import (
	"fmt"
	"math/big"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var base58Index = func() [256]int {
	var idx [256]int
	for i := range idx {
		idx[i] = -1
	}
	for i, c := range []byte(base58Alphabet) {
		idx[c] = i
	}
	return idx
}()

func builtinToBase58(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "toBase58 expects 1 argument"}
	}
	payload, ok := bytesArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "toBase58 expects bytes"}
	}
	return &String{Value: encodeBase58(payload)}, nil
}

func builtinFromBase58(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "fromBase58 expects 1 argument"}
	}
	text, ok := stringArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "fromBase58 expects string"}
	}
	payload, err := decodeBase58(text)
	if err != nil {
		return nil, recoverableError("base58_decode", "fromBase58 error: "+err.Error())
	}
	return &Bytes{Value: payload}, nil
}

func encodeBase58(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}

	leadingZeros := 0
	for leadingZeros < len(payload) && payload[leadingZeros] == 0 {
		leadingZeros++
	}

	n := new(big.Int).SetBytes(payload)
	base := big.NewInt(58)
	mod := new(big.Int)
	out := make([]byte, 0, len(payload)*2)

	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		out = append(out, base58Alphabet[mod.Int64()])
	}
	for i := 0; i < leadingZeros; i++ {
		out = append(out, base58Alphabet[0])
	}
	reverseBytes(out)
	return string(out)
}

func decodeBase58(text string) ([]byte, error) {
	if text == "" {
		return []byte{}, nil
	}

	leadingOnes := 0
	for leadingOnes < len(text) && text[leadingOnes] == base58Alphabet[0] {
		leadingOnes++
	}

	n := big.NewInt(0)
	base := big.NewInt(58)
	for i := 0; i < len(text); i++ {
		ch := text[i]
		val := base58Index[ch]
		if val < 0 {
			return nil, fmt.Errorf("invalid base58 character %q at index %d", ch, i)
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(val)))
	}

	decoded := n.Bytes()
	out := make([]byte, leadingOnes+len(decoded))
	copy(out[leadingOnes:], decoded)
	return out, nil
}

func reverseBytes(buf []byte) {
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
}
