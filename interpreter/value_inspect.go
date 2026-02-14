package interpreter

import (
	"bytes"
	"fmt"
)

func inspectObjectPairs(pairs map[string]Value) string {
	var out bytes.Buffer
	out.WriteString("{")
	first := true
	for k, v := range pairs {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(inspectObjectKey(k))
		out.WriteString(": ")
		out.WriteString(v.Inspect())
	}
	out.WriteString("}")
	return out.String()
}

func inspectObjectKey(key string) string {
	if isIdentifierKey(key) {
		return key
	}
	return fmt.Sprintf("%q", key)
}

func isIdentifierKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	for i, r := range key {
		isAlpha := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isDigit := r >= '0' && r <= '9'
		if i == 0 {
			if !(isAlpha || r == '_') {
				return false
			}
			continue
		}
		if !(isAlpha || isDigit || r == '_') {
			return false
		}
	}
	return true
}
