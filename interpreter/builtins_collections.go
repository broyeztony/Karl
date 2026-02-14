package interpreter

func registerCollectionBuiltins() {
	builtins["map"] = &Builtin{Name: "map", Fn: builtinMap}
	builtins["get"] = &Builtin{Name: "get", Fn: builtinMapGet}
	builtins["set"] = &Builtin{Name: "set", Fn: builtinMapSet}
	builtins["add"] = &Builtin{Name: "add", Fn: builtinSetAdd}
	builtins["has"] = &Builtin{Name: "has", Fn: builtinMapHas}
	builtins["delete"] = &Builtin{Name: "delete", Fn: builtinMapDelete}
	builtins["keys"] = &Builtin{Name: "keys", Fn: builtinMapKeys}
	builtins["values"] = &Builtin{Name: "values", Fn: builtinMapValues}
}
