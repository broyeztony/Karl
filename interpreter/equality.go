package interpreter

func StrictEqual(left, right Value) bool {
	if left == nil || right == nil {
		return left == right
	}
	if left.Type() != right.Type() {
		return false
	}

	switch l := left.(type) {
	case *Integer:
		return l.Value == right.(*Integer).Value
	case *Float:
		return l.Value == right.(*Float).Value
	case *Boolean:
		return l.Value == right.(*Boolean).Value
	case *String:
		return l.Value == right.(*String).Value
	case *Char:
		return l.Value == right.(*Char).Value
	case *Null, *Unit:
		return true
	case *Map:
		return left == right
	default:
		// Arrays, objects, functions, tasks, channels: identity equality.
		return left == right
	}
}
