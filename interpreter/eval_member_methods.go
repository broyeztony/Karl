package interpreter

func (e *Evaluator) arrayMethod(arr *Array, name string) (Value, *Signal, error) {
	switch name {
	case "map", "filter", "reduce", "forEach", "sum", "find", "sort":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, arr)}, nil, nil
	case "push":
		return &Builtin{
			Name: "push",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "push expects 1 argument"}
				}
				arr.Elements = append(arr.Elements, args[0])
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown array member: " + name}
	}
}

func (e *Evaluator) channelMethod(ch *Channel, name string) (Value, *Signal, error) {
	switch name {
	case "send", "recv", "done":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, ch)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown channel member: " + name}
	}
}

func (e *Evaluator) stringMethod(str *String, name string) (Value, *Signal, error) {
	switch name {
	case "split", "chars", "trim", "toLower", "toUpper", "contains", "startsWith", "endsWith", "replace":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, str)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown string member: " + name}
	}
}

func (e *Evaluator) mapMethod(m *Map, name string) (Value, *Signal, error) {
	switch name {
	case "get", "set", "has", "delete", "keys", "values":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, m)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown map member: " + name}
	}
}

func (e *Evaluator) setMethod(s *Set, name string) (Value, *Signal, error) {
	switch name {
	case "add", "has", "delete", "values":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, s)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown set member: " + name}
	}
}

func (e *Evaluator) treeMethod(t *Tree, name string) (Value, *Signal, error) {
	switch name {
	case "set":
		return &Builtin{
			Name: "set",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 2 {
					return nil, &RuntimeError{Message: "set expects key and value"}
				}
				if err := t.Set(args[0], args[1]); err != nil {
					return nil, err
				}
				return t, nil
			},
		}, nil, nil
	case "get":
		return &Builtin{
			Name: "get",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "get expects key"}
				}
				return t.Get(args[0])
			},
		}, nil, nil
	case "has":
		return &Builtin{
			Name: "has",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "has expects key"}
				}
				ok, err := t.Has(args[0])
				if err != nil {
					return nil, err
				}
				return &Boolean{Value: ok}, nil
			},
		}, nil, nil
	case "delete":
		return &Builtin{
			Name: "delete",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "delete expects key"}
				}
				ok, err := t.Delete(args[0])
				if err != nil {
					return nil, err
				}
				return &Boolean{Value: ok}, nil
			},
		}, nil, nil
	case "kind":
		return &Builtin{
			Name: "kind",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "kind expects no arguments"}
				}
				return &String{Value: t.kind}, nil
			},
		}, nil, nil
	case "min":
		return &Builtin{
			Name: "min",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "min expects no arguments"}
				}
				return treeItemValue(t.MinNode()), nil
			},
		}, nil, nil
	case "max":
		return &Builtin{
			Name: "max",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "max expects no arguments"}
				}
				return treeItemValue(t.MaxNode()), nil
			},
		}, nil, nil
	case "maxDepth":
		return &Builtin{
			Name: "maxDepth",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "maxDepth expects no arguments"}
				}
				return &Integer{Value: int64(t.MaxDepth())}, nil
			},
		}, nil, nil
	case "maxWidth":
		return &Builtin{
			Name: "maxWidth",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "maxWidth expects no arguments"}
				}
				return &Integer{Value: int64(t.MaxWidth())}, nil
			},
		}, nil, nil
	case "lowerBound":
		return &Builtin{
			Name: "lowerBound",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "lowerBound expects key"}
				}
				n, err := t.LowerBound(args[0])
				if err != nil {
					return nil, err
				}
				return treeItemValue(n), nil
			},
		}, nil, nil
	case "upperBound":
		return &Builtin{
			Name: "upperBound",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "upperBound expects key"}
				}
				n, err := t.UpperBound(args[0])
				if err != nil {
					return nil, err
				}
				return treeItemValue(n), nil
			},
		}, nil, nil
	case "floor":
		return &Builtin{
			Name: "floor",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "floor expects key"}
				}
				n, err := t.Floor(args[0])
				if err != nil {
					return nil, err
				}
				return treeItemValue(n), nil
			},
		}, nil, nil
	case "ceil":
		return &Builtin{
			Name: "ceil",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "ceil expects key"}
				}
				n, err := t.Ceil(args[0])
				if err != nil {
					return nil, err
				}
				return treeItemValue(n), nil
			},
		}, nil, nil
	case "predecessor":
		return &Builtin{
			Name: "predecessor",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "predecessor expects key"}
				}
				n, err := t.Predecessor(args[0])
				if err != nil {
					return nil, err
				}
				return treeItemValue(n), nil
			},
		}, nil, nil
	case "successor":
		return &Builtin{
			Name: "successor",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "successor expects key"}
				}
				n, err := t.Successor(args[0])
				if err != nil {
					return nil, err
				}
				return treeItemValue(n), nil
			},
		}, nil, nil
	case "closest":
		return &Builtin{
			Name: "closest",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) < 1 || len(args) > 2 {
					return nil, &RuntimeError{Message: "closest expects key and optional opts"}
				}
				tieUpper := false
				if len(args) == 2 {
					opts, ok := objectPairs(args[1])
					if !ok {
						return nil, &RuntimeError{Message: "closest opts must be object"}
					}
					if tie, ok := opts["tie"]; ok && !Equivalent(tie, NullValue) {
						tieText, ok := stringArg(tie)
						if !ok {
							return nil, &RuntimeError{Message: "closest tie must be string"}
						}
						if tieText == "upper" {
							tieUpper = true
						} else if tieText != "lower" {
							return nil, &RuntimeError{Message: "closest tie must be \"lower\" or \"upper\""}
						}
					}
				}
				n, exact, err := t.Closest(args[0], tieUpper)
				if err != nil {
					return nil, err
				}
				if n == nil {
					return NullValue, nil
				}
				return &Object{Pairs: map[string]Value{
					"key":   treeKeyToValue(n.key),
					"value": n.value,
					"exact": &Boolean{Value: exact},
				}}, nil
			},
		}, nil, nil
	case "range":
		return &Builtin{
			Name: "range",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) < 2 || len(args) > 3 {
					return nil, &RuntimeError{Message: "range expects from, to, and optional opts"}
				}
				includeFrom := true
				includeTo := true
				limit := 0
				if len(args) == 3 {
					opts, ok := objectPairs(args[2])
					if !ok {
						return nil, &RuntimeError{Message: "range opts must be object"}
					}
					if v, ok := opts["includeFrom"]; ok && !Equivalent(v, NullValue) {
						b, ok := v.(*Boolean)
						if !ok {
							return nil, &RuntimeError{Message: "range includeFrom must be bool"}
						}
						includeFrom = b.Value
					}
					if v, ok := opts["includeTo"]; ok && !Equivalent(v, NullValue) {
						b, ok := v.(*Boolean)
						if !ok {
							return nil, &RuntimeError{Message: "range includeTo must be bool"}
						}
						includeTo = b.Value
					}
					if v, ok := opts["limit"]; ok && !Equivalent(v, NullValue) {
						i, ok := v.(*Integer)
						if !ok {
							return nil, &RuntimeError{Message: "range limit must be integer"}
						}
						if i.Value <= 0 {
							return nil, &RuntimeError{Message: "range limit must be > 0"}
						}
						limit = int(i.Value)
					}
				}
				items, err := t.Range(args[0], args[1], includeFrom, includeTo, limit)
				if err != nil {
					return nil, err
				}
				return &Array{Elements: items}, nil
			},
		}, nil, nil
	case "path":
		return &Builtin{
			Name: "path",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "path expects key"}
				}
				path, err := t.Path(args[0])
				if err != nil {
					return nil, err
				}
				if path == nil {
					return NullValue, nil
				}
				return &Array{Elements: path}, nil
			},
		}, nil, nil
	case "keys":
		return &Builtin{
			Name: "keys",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "keys expects no arguments"}
				}
				return &Array{Elements: t.Keys()}, nil
			},
		}, nil, nil
	case "values":
		return &Builtin{
			Name: "values",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "values expects no arguments"}
				}
				return &Array{Elements: t.Values()}, nil
			},
		}, nil, nil
	case "items":
		return &Builtin{
			Name: "items",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "items expects no arguments"}
				}
				return &Array{Elements: t.Items()}, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown tree member: " + name}
	}
}

func (e *Evaluator) nTreeMethod(t *NTree, name string) (Value, *Signal, error) {
	nodeArray := func(nodes []*nTreeNode) *Array {
		out := make([]Value, 0, len(nodes))
		for _, n := range nodes {
			if n == nil {
				continue
			}
			out = append(out, n.nodeValue())
		}
		return &Array{Elements: out}
	}

	stringID := func(v Value, label string) (string, error) {
		id, ok := stringArg(v)
		if !ok {
			return "", &RuntimeError{Message: label + " must be string"}
		}
		return id, nil
	}

	switch name {
	case "get":
		return &Builtin{
			Name: "get",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "get expects id"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				return t.nodeValueByID(id), nil
			},
		}, nil, nil
	case "set":
		return &Builtin{
			Name: "set",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 2 {
					return nil, &RuntimeError{Message: "set expects id and value"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				return &Boolean{Value: t.SetValue(id, args[1])}, nil
			},
		}, nil, nil
	case "append":
		return &Builtin{
			Name: "append",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 3 {
					return nil, &RuntimeError{Message: "append expects parentId, childId, value"}
				}
				parentID, err := stringID(args[0], "parentId")
				if err != nil {
					return nil, err
				}
				childID, err := stringID(args[1], "childId")
				if err != nil {
					return nil, err
				}
				if err := t.Append(parentID, childID, args[2]); err != nil {
					return nil, err
				}
				return t, nil
			},
		}, nil, nil
	case "prepend":
		return &Builtin{
			Name: "prepend",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 3 {
					return nil, &RuntimeError{Message: "prepend expects parentId, childId, value"}
				}
				parentID, err := stringID(args[0], "parentId")
				if err != nil {
					return nil, err
				}
				childID, err := stringID(args[1], "childId")
				if err != nil {
					return nil, err
				}
				if err := t.Prepend(parentID, childID, args[2]); err != nil {
					return nil, err
				}
				return t, nil
			},
		}, nil, nil
	case "insertAt":
		return &Builtin{
			Name: "insertAt",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 4 {
					return nil, &RuntimeError{Message: "insertAt expects parentId, index, childId, value"}
				}
				parentID, err := stringID(args[0], "parentId")
				if err != nil {
					return nil, err
				}
				index, ok := args[1].(*Integer)
				if !ok {
					return nil, &RuntimeError{Message: "index must be integer"}
				}
				if index.Value < 0 {
					return nil, &RuntimeError{Message: "index must be >= 0"}
				}
				childID, err := stringID(args[2], "childId")
				if err != nil {
					return nil, err
				}
				if err := t.InsertAt(parentID, int(index.Value), childID, args[3]); err != nil {
					return nil, err
				}
				return t, nil
			},
		}, nil, nil
	case "remove":
		return &Builtin{
			Name: "remove",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) < 1 || len(args) > 2 {
					return nil, &RuntimeError{Message: "remove expects id and optional opts"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				subtree := true
				if len(args) == 2 {
					opts, ok := objectPairs(args[1])
					if !ok {
						return nil, &RuntimeError{Message: "remove opts must be object"}
					}
					if v, ok := opts["subtree"]; ok && !Equivalent(v, NullValue) {
						flag, ok := v.(*Boolean)
						if !ok {
							return nil, &RuntimeError{Message: "remove subtree must be bool"}
						}
						subtree = flag.Value
					}
				}
				removed, err := t.Remove(id, subtree)
				if err != nil {
					return nil, err
				}
				return &Boolean{Value: removed}, nil
			},
		}, nil, nil
	case "move":
		return &Builtin{
			Name: "move",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) < 2 || len(args) > 3 {
					return nil, &RuntimeError{Message: "move expects id, newParentId, and optional opts"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				parentID, err := stringID(args[1], "newParentId")
				if err != nil {
					return nil, err
				}
				index := 0
				hasIndex := false
				if len(args) == 3 {
					opts, ok := objectPairs(args[2])
					if !ok {
						return nil, &RuntimeError{Message: "move opts must be object"}
					}
					if v, ok := opts["index"]; ok && !Equivalent(v, NullValue) {
						i, ok := v.(*Integer)
						if !ok {
							return nil, &RuntimeError{Message: "move index must be integer"}
						}
						if i.Value < 0 {
							return nil, &RuntimeError{Message: "move index must be >= 0"}
						}
						index = int(i.Value)
						hasIndex = true
					}
				}
				if err := t.Move(id, parentID, index, hasIndex); err != nil {
					return nil, err
				}
				return t, nil
			},
		}, nil, nil
	case "parent":
		return &Builtin{
			Name: "parent",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "parent expects id"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				parent, ok := t.Parent(id)
				if !ok || parent == nil {
					return NullValue, nil
				}
				return parent.nodeValue(), nil
			},
		}, nil, nil
	case "children":
		return &Builtin{
			Name: "children",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "children expects id"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				children, err := t.Children(id)
				if err != nil {
					return nil, err
				}
				return nodeArray(children), nil
			},
		}, nil, nil
	case "siblings":
		return &Builtin{
			Name: "siblings",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) < 1 || len(args) > 2 {
					return nil, &RuntimeError{Message: "siblings expects id and optional opts"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				includeSelf := false
				if len(args) == 2 {
					opts, ok := objectPairs(args[1])
					if !ok {
						return nil, &RuntimeError{Message: "siblings opts must be object"}
					}
					if v, ok := opts["includeSelf"]; ok && !Equivalent(v, NullValue) {
						b, ok := v.(*Boolean)
						if !ok {
							return nil, &RuntimeError{Message: "siblings includeSelf must be bool"}
						}
						includeSelf = b.Value
					}
				}
				siblings, err := t.Siblings(id, includeSelf)
				if err != nil {
					return nil, err
				}
				return nodeArray(siblings), nil
			},
		}, nil, nil
	case "ancestors":
		return &Builtin{
			Name: "ancestors",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "ancestors expects id"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				ancestors, err := t.Ancestors(id)
				if err != nil {
					return nil, err
				}
				return nodeArray(ancestors), nil
			},
		}, nil, nil
	case "path":
		return &Builtin{
			Name: "path",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "path expects id"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				path, ok := t.Path(id)
				if !ok {
					return NullValue, nil
				}
				return nodeArray(path), nil
			},
		}, nil, nil
	case "descendants":
		return &Builtin{
			Name: "descendants",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) < 1 || len(args) > 2 {
					return nil, &RuntimeError{Message: "descendants expects id and optional opts"}
				}
				id, err := stringID(args[0], "id")
				if err != nil {
					return nil, err
				}
				traversal := "dfs"
				if len(args) == 2 {
					opts, ok := objectPairs(args[1])
					if !ok {
						return nil, &RuntimeError{Message: "descendants opts must be object"}
					}
					if v, ok := opts["traversal"]; ok && !Equivalent(v, NullValue) {
						mode, ok := stringArg(v)
						if !ok {
							return nil, &RuntimeError{Message: "descendants traversal must be string"}
						}
						traversal = mode
					}
				}
				desc, err := t.Descendants(id, traversal)
				if err != nil {
					return nil, err
				}
				return nodeArray(desc), nil
			},
		}, nil, nil
	case "find":
		return &Builtin{
			Name: "find",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "find expects function"}
				}
				if !isCallable(args[0]) {
					return nil, &RuntimeError{Message: "find expects function"}
				}
				if t.rootID == "" {
					return NullValue, nil
				}
				stack := []string{t.rootID}
				for len(stack) > 0 {
					last := len(stack) - 1
					id := stack[last]
					stack = stack[:last]
					n, ok := t.nodeByID(id)
					if !ok {
						continue
					}
					out, sig, err := e.applyFunction(args[0], []Value{n.nodeValue()})
					if err != nil {
						return nil, err
					}
					if sig != nil {
						return nil, &RuntimeError{Message: "break/continue outside loop"}
					}
					if isTruthy(out) {
						return n.nodeValue(), nil
					}
					for i := len(n.children) - 1; i >= 0; i-- {
						stack = append(stack, n.children[i])
					}
				}
				return NullValue, nil
			},
		}, nil, nil
	case "findAll":
		return &Builtin{
			Name: "findAll",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "findAll expects function"}
				}
				if !isCallable(args[0]) {
					return nil, &RuntimeError{Message: "findAll expects function"}
				}
				out := []Value{}
				if t.rootID == "" {
					return &Array{Elements: out}, nil
				}
				stack := []string{t.rootID}
				for len(stack) > 0 {
					last := len(stack) - 1
					id := stack[last]
					stack = stack[:last]
					n, ok := t.nodeByID(id)
					if !ok {
						continue
					}
					keep, sig, err := e.applyFunction(args[0], []Value{n.nodeValue()})
					if err != nil {
						return nil, err
					}
					if sig != nil {
						return nil, &RuntimeError{Message: "break/continue outside loop"}
					}
					if isTruthy(keep) {
						out = append(out, n.nodeValue())
					}
					for i := len(n.children) - 1; i >= 0; i-- {
						stack = append(stack, n.children[i])
					}
				}
				return &Array{Elements: out}, nil
			},
		}, nil, nil
	case "root":
		return &Builtin{
			Name: "root",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "root expects no arguments"}
				}
				if t.rootID == "" {
					return NullValue, nil
				}
				return t.nodeValueByID(t.rootID), nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown ntree member: " + name}
	}
}

func (e *Evaluator) taskMethod(t *Task, name string) (Value, *Signal, error) {
	switch name {
	case "then":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, t)}, nil, nil
	case "cancel":
		return &Builtin{
			Name: "cancel",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "cancel expects no arguments"}
				}
				t.Cancel()
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown task member: " + name}
	}
}

func (e *Evaluator) processMethod(p *Process, name string) (Value, *Signal, error) {
	switch name {
	case "pid":
		return &Integer{Value: p.PID()}, nil, nil
	case "running":
		return &Boolean{Value: p.Running()}, nil, nil
	case "stdin":
		stream, ok := p.inputStream()
		if !ok {
			return nil, nil, recoverableError("process_state", "stdin is only available when stdin mode is \"pipe\"")
		}
		return stream, nil, nil
	case "stdout":
		stream, ok := p.outputStream()
		if !ok {
			return nil, nil, recoverableError("process_state", "stdout is only available when stdout mode is \"pipe\"")
		}
		return stream, nil, nil
	case "stderr":
		stream, ok := p.errorStream()
		if !ok {
			return nil, nil, recoverableError("process_state", "stderr is only available when stderr mode is \"pipe\"")
		}
		return stream, nil, nil
	case "abort":
		return &Builtin{
			Name: "abort",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "abort expects no arguments"}
				}
				if err := p.Abort(); err != nil {
					return nil, recoverableError("process_state", "process abort error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	case "kill":
		return &Builtin{
			Name: "kill",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "kill expects no arguments"}
				}
				if err := p.Kill(); err != nil {
					return nil, recoverableError("process_state", "process kill error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	case "signal":
		return &Builtin{
			Name: "signal",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "signal expects signal name"}
				}
				name, ok := stringArg(args[0])
				if !ok {
					return nil, &RuntimeError{Message: "signal expects string signal name"}
				}
				if err := p.Signal(name); err != nil {
					return nil, recoverableError("process_state", "process signal error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown process member: " + name}
	}
}

func (e *Evaluator) streamReaderMethod(s *StreamReader, name string) (Value, *Signal, error) {
	switch name {
	case "read":
		return &Builtin{
			Name: "read",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) > 1 {
					return nil, &RuntimeError{Message: "read expects optional size"}
				}
				size := int64(defaultStreamReadSize)
				if len(args) == 1 {
					i, ok := args[0].(*Integer)
					if !ok {
						return nil, &RuntimeError{Message: "read size must be integer"}
					}
					if i.Value <= 0 {
						return nil, &RuntimeError{Message: "read size must be > 0"}
					}
					size = i.Value
				}

				chunk, eof, err := s.ReadChunk(int(size))
				if err != nil {
					return nil, recoverableError("stream_read", "stream read error: "+err.Error())
				}
				if eof {
					return &Array{Elements: []Value{NullValue, &Boolean{Value: true}}}, nil
				}
				if s.Mode() == streamTypeBytes {
					copied := append([]byte{}, chunk...)
					return &Array{Elements: []Value{&Bytes{Value: copied}, &Boolean{Value: false}}}, nil
				}
				return &Array{Elements: []Value{&String{Value: string(chunk)}, &Boolean{Value: false}}}, nil
			},
		}, nil, nil
	case "close":
		return &Builtin{
			Name: "close",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "close expects no arguments"}
				}
				if err := s.Close(); err != nil {
					return nil, recoverableError("stream_close", "stream close error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown stream reader member: " + name}
	}
}

func (e *Evaluator) streamWriterMethod(s *StreamWriter, name string) (Value, *Signal, error) {
	switch name {
	case "write":
		return &Builtin{
			Name: "write",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, &RuntimeError{Message: "write expects 1 argument"}
				}
				mode := s.Mode()
				var payload []byte
				if mode == streamTypeBytes {
					bytesPayload, ok := bytesArg(args[0])
					if !ok {
						return nil, &RuntimeError{Message: "write expects bytes payload in BYTES mode"}
					}
					payload = bytesPayload
				} else {
					textPayload, ok := stringArg(args[0])
					if !ok {
						return nil, &RuntimeError{Message: "write expects string payload in TEXT mode"}
					}
					payload = []byte(textPayload)
				}
				n, err := s.WriteChunk(payload)
				if err != nil {
					return nil, recoverableError("stream_write", "stream write error: "+err.Error())
				}
				return &Integer{Value: int64(n)}, nil
			},
		}, nil, nil
	case "close":
		return &Builtin{
			Name: "close",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "close expects no arguments"}
				}
				if err := s.Close(); err != nil {
					return nil, recoverableError("stream_close", "stream close error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown stream writer member: " + name}
	}
}

func (e *Evaluator) streamValueMethod(value Value, name string) (Value, *Signal, error) {
	switch name {
	case "read":
		return &Builtin{
			Name: "read",
			Fn: func(eval *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "read expects no arguments"}
				}
				return streamReadValue(eval, value)
			},
		}, nil, nil
	case "close":
		return &Builtin{
			Name: "close",
			Fn: func(_ *Evaluator, args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, &RuntimeError{Message: "close expects no arguments"}
				}
				if err := streamCloseValue(value); err != nil {
					if _, ok := err.(*RecoverableError); ok {
						return nil, err
					}
					return nil, recoverableError("stream_close", "stream close error: "+err.Error())
				}
				return UnitValue, nil
			},
		}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown stream member: " + name}
	}
}
