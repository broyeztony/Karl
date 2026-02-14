package interpreter

import (
	"net/http"
	"strings"
)

func extractHeaders(val Value) (map[string]string, error) {
	switch headers := val.(type) {
	case *Object:
		out := make(map[string]string, len(headers.Pairs))
		for k, v := range headers.Pairs {
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k] = str
		}
		return out, nil
	case *ModuleObject:
		if headers.Env == nil {
			return nil, &RuntimeError{Message: "http headers must be object or map"}
		}
		out := make(map[string]string)
		for k, v := range headers.Env.Snapshot() {
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k] = str
		}
		return out, nil
	case *Map:
		out := make(map[string]string, len(headers.Pairs))
		for k, v := range headers.Pairs {
			if k.Type != STRING && k.Type != CHAR {
				return nil, &RuntimeError{Message: "http headers keys must be strings"}
			}
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k.Value] = str
		}
		return out, nil
	default:
		return nil, &RuntimeError{Message: "http headers must be object or map"}
	}
}

func httpResponseObject(resp *http.Response, body []byte) Value {
	headerMap := &Map{Pairs: make(map[MapKey]Value)}
	for key, values := range resp.Header {
		headerMap.Pairs[MapKey{Type: STRING, Value: key}] = &String{Value: strings.Join(values, ", ")}
	}
	return &Object{Pairs: map[string]Value{
		"status":  &Integer{Value: int64(resp.StatusCode)},
		"headers": headerMap,
		"body":    &String{Value: string(body)},
	}}
}
