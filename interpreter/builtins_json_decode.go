package interpreter

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

func decodeJSONValue(value interface{}) (Value, error) {
	switch v := value.(type) {
	case nil:
		return NullValue, nil
	case bool:
		return &Boolean{Value: v}, nil
	case string:
		return &String{Value: v}, nil
	case json.Number:
		raw := v.String()
		if strings.ContainsAny(raw, ".eE") {
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
				return nil, recoverableError("decodeJson", "decodeJson invalid float: "+raw)
			}
			return &Float{Value: f}, nil
		}
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, recoverableError("decodeJson", "decodeJson integer overflow: "+raw)
		}
		return &Integer{Value: i}, nil
	case []interface{}:
		out := make([]Value, 0, len(v))
		for _, el := range v {
			val, err := decodeJSONValue(el)
			if err != nil {
				return nil, err
			}
			out = append(out, val)
		}
		return &Array{Elements: out}, nil
	case map[string]interface{}:
		out := make(map[string]Value, len(v))
		for k, val := range v {
			decoded, err := decodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = decoded
		}
		return &Object{Pairs: out}, nil
	default:
		return nil, recoverableError("decodeJson", "decodeJson unsupported value")
	}
}
