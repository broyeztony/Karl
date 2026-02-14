package interpreter

// isTruthy determines if a value is truthy according to Karl's truthy/falsy rules.
func isTruthy(val Value) bool {
	switch v := val.(type) {
	case *Null:
		return false
	case *Boolean:
		return v.Value
	case *Integer:
		return v.Value != 0
	case *Float:
		return v.Value != 0.0
	case *String:
		return v.Value != ""
	case *Array:
		return len(v.Elements) > 0
	case *Object:
		return len(v.Pairs) > 0
	case *Map:
		return len(v.Pairs) > 0
	case *Set:
		return len(v.Elements) > 0
	default:
		// All other types are truthy (functions, tasks, channels, etc.)
		return true
	}
}
