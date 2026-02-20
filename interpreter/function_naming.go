package interpreter

func assignFunctionName(value Value, name string) Value {
	if name == "" {
		return value
	}
	fn, ok := value.(*Function)
	if !ok {
		return value
	}
	if fn.Name == "" {
		fn.Name = name
	}
	return value
}

func functionDebugName(fn *Function) string {
	if fn == nil || fn.Name == "" {
		return "<lambda>"
	}
	return fn.Name
}
