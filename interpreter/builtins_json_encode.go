package interpreter

import "math"

func encodeJSONValue(value Value) (interface{}, error) {
	switch v := value.(type) {
	case *Null:
		return nil, nil
	case *Boolean:
		return v.Value, nil
	case *Integer:
		return v.Value, nil
	case *Float:
		if math.IsNaN(v.Value) || math.IsInf(v.Value, 0) {
			return nil, &RuntimeError{Message: "jsonEncode cannot encode NaN or Inf"}
		}
		return v.Value, nil
	case *String:
		return v.Value, nil
	case *Char:
		return v.Value, nil
	case *Array:
		out := make([]interface{}, 0, len(v.Elements))
		for _, el := range v.Elements {
			enc, err := encodeJSONValue(el)
			if err != nil {
				return nil, err
			}
			out = append(out, enc)
		}
		return out, nil
	case *Object:
		out := make(map[string]interface{}, len(v.Pairs))
		for k, val := range v.Pairs {
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = enc
		}
		return out, nil
	case *ModuleObject:
		if v.Env == nil {
			return map[string]interface{}{}, nil
		}
		out := make(map[string]interface{})
		for k, val := range v.Env.Snapshot() {
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = enc
		}
		return out, nil
	case *Map:
		out := make(map[string]interface{}, len(v.Pairs))
		for k, val := range v.Pairs {
			if k.Type != STRING && k.Type != CHAR {
				return nil, &RuntimeError{Message: "jsonEncode only supports map with string keys"}
			}
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k.Value] = enc
		}
		return out, nil
	default:
		return nil, &RuntimeError{Message: "jsonEncode does not support " + string(value.Type())}
	}
}
