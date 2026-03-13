package interpreter

import (
	"sort"
	"strings"
)

func isCallable(val Value) bool {
	switch val.(type) {
	case *Function, *Builtin, *Partial:
		return true
	default:
		return false
	}
}

func parseStreamType(val Value, field string) (string, error) {
	mode, ok := stringArg(val)
	if !ok {
		return "", &RuntimeError{Message: "process " + field + " must be string"}
	}
	mode = strings.TrimSpace(strings.ToLower(mode))
	switch mode {
	case streamTypeText, streamTypeBytes:
		return mode, nil
	default:
		return "", &RuntimeError{Message: "process " + field + " must be \"text\" or \"bytes\""}
	}
}

func rejectUnknownObjectKeys(pairs map[string]Value, label string, allowed []string) error {
	if len(pairs) == 0 {
		return nil
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}

	unknown := make([]string, 0, len(pairs))
	for key := range pairs {
		if _, ok := allowedSet[key]; !ok {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) == 0 {
		return nil
	}

	sort.Strings(unknown)
	return &RuntimeError{Message: label + " has unknown field(s): " + strings.Join(unknown, ", ")}
}
