package interpreter

func objectPairs(value Value) (map[string]Value, bool) {
	switch v := value.(type) {
	case *Object:
		return v.Pairs, true
	case *ModuleObject:
		if v.Env == nil {
			return nil, false
		}
		return v.Env.Snapshot(), true
	default:
		return nil, false
	}
}
