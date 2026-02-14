package interpreter

func Equivalent(left, right Value) bool {
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
	case *Array:
		r := right.(*Array)
		if len(l.Elements) != len(r.Elements) {
			return false
		}
		for i := range l.Elements {
			if !Equivalent(l.Elements[i], r.Elements[i]) {
				return false
			}
		}
		return true
	case *Object:
		return equivalentObjectPairs(l.Pairs, right)
	case *ModuleObject:
		if l.Env == nil {
			return false
		}
		return equivalentObjectPairs(l.Env.Snapshot(), right)
	case *Map:
		r := right.(*Map)
		if len(l.Pairs) != len(r.Pairs) {
			return false
		}
		for k, v := range l.Pairs {
			ov, ok := r.Pairs[k]
			if !ok {
				return false
			}
			if !Equivalent(v, ov) {
				return false
			}
		}
		return true
	default:
		return left == right
	}
}

func equivalentObjectPairs(left map[string]Value, right Value) bool {
	rightPairs, ok := objectPairs(right)
	if !ok {
		return false
	}
	if len(left) != len(rightPairs) {
		return false
	}
	for k, v := range left {
		ov, ok := rightPairs[k]
		if !ok {
			return false
		}
		if !Equivalent(v, ov) {
			return false
		}
	}
	return true
}
