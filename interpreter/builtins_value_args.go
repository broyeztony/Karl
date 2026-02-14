package interpreter

func stringArg(val Value) (string, bool) {
	switch v := val.(type) {
	case *String:
		return v.Value, true
	case *Char:
		return v.Value, true
	default:
		return "", false
	}
}
