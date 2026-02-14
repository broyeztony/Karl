package interpreter

import (
	"bytes"
)

type Array struct {
	Elements []Value
}

func (a *Array) Type() ValueType { return ARRAY }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	out.WriteString("[")
	for i, el := range a.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(el.Inspect())
	}
	out.WriteString("]")
	return out.String()
}

type Object struct {
	Pairs map[string]Value
}

func (o *Object) Type() ValueType { return OBJECT }
func (o *Object) Inspect() string {
	return inspectObjectPairs(o.Pairs)
}

type ModuleObject struct {
	Env *Environment
}

func (o *ModuleObject) Type() ValueType { return OBJECT }
func (o *ModuleObject) Inspect() string {
	if o.Env == nil {
		return "{}"
	}
	return inspectObjectPairs(o.Env.Snapshot())
}

type MapKey struct {
	Type  ValueType
	Value string
}

type Map struct {
	Pairs map[MapKey]Value
}

func (m *Map) Type() ValueType { return MAP }
func (m *Map) Inspect() string {
	var out bytes.Buffer
	out.WriteString("map{")
	first := true
	for k, v := range m.Pairs {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(formatMapKey(k))
		out.WriteString(": ")
		out.WriteString(v.Inspect())
	}
	out.WriteString("}")
	return out.String()
}

type Set struct {
	Elements map[MapKey]struct{}
}

func (s *Set) Type() ValueType { return SET }
func (s *Set) Inspect() string {
	var out bytes.Buffer
	out.WriteString("set{")
	first := true
	for k := range s.Elements {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(formatMapKey(k))
	}
	out.WriteString("}")
	return out.String()
}
