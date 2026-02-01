package interpreter

import (
	"bytes"
	"fmt"
	"karl/ast"
	"strings"
	"sync"
)

type ValueType string

const (
	INTEGER ValueType = "INTEGER"
	FLOAT   ValueType = "FLOAT"
	BOOLEAN ValueType = "BOOLEAN"
	STRING  ValueType = "STRING"
	CHAR    ValueType = "CHAR"
	NULL    ValueType = "NULL"
	UNIT    ValueType = "UNIT"
	ARRAY   ValueType = "ARRAY"
	OBJECT  ValueType = "OBJECT"
	MAP     ValueType = "MAP"
	SET     ValueType = "SET"
	FUNC    ValueType = "FUNCTION"
	BUILTIN ValueType = "BUILTIN"
	TASK    ValueType = "TASK"
	CHANNEL ValueType = "CHANNEL"
	PARTIAL ValueType = "PARTIAL"
)

type Value interface {
	Type() ValueType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ValueType { return INTEGER }
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
}

func (f *Float) Type() ValueType { return FLOAT }
func (f *Float) Inspect() string { return fmt.Sprintf("%g", f.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ValueType { return BOOLEAN }
func (b *Boolean) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}

type String struct {
	Value string
}

func (s *String) Type() ValueType { return STRING }
func (s *String) Inspect() string { return fmt.Sprintf("%q", s.Value) }

type Char struct {
	Value string
}

func (c *Char) Type() ValueType { return CHAR }
func (c *Char) Inspect() string { return fmt.Sprintf("'%s'", c.Value) }

type Null struct{}

func (n *Null) Type() ValueType { return NULL }
func (n *Null) Inspect() string { return "null" }

type Unit struct{}

func (u *Unit) Type() ValueType { return UNIT }
func (u *Unit) Inspect() string { return "()" }

var (
	NullValue = &Null{}
	UnitValue = &Unit{}
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

func inspectObjectPairs(pairs map[string]Value) string {
	var out bytes.Buffer
	out.WriteString("{")
	first := true
	for k, v := range pairs {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(inspectObjectKey(k))
		out.WriteString(": ")
		out.WriteString(v.Inspect())
	}
	out.WriteString("}")
	return out.String()
}

func inspectObjectKey(key string) string {
	if isIdentifierKey(key) {
		return key
	}
	return fmt.Sprintf("%q", key)
}

func isIdentifierKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	for i, r := range key {
		isAlpha := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isDigit := r >= '0' && r <= '9'
		if i == 0 {
			if !(isAlpha || r == '_') {
				return false
			}
			continue
		}
		if !(isAlpha || isDigit || r == '_') {
			return false
		}
	}
	return true
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

type Function struct {
	Params []ast.Pattern
	Body   ast.Expression
	Env    *Environment
}

func (f *Function) Type() ValueType { return FUNC }
func (f *Function) Inspect() string { return "<function>" }

type BuiltinFunction func(e *Evaluator, args []Value) (Value, error)

type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

func (b *Builtin) Type() ValueType { return BUILTIN }
func (b *Builtin) Inspect() string { return "<builtin " + b.Name + ">" }

type Partial struct {
	Target Value
	Args   []Value
}

func (p *Partial) Type() ValueType { return PARTIAL }
func (p *Partial) Inspect() string {
	parts := []string{}
	for _, arg := range p.Args {
		if arg == nil {
			parts = append(parts, "_")
		} else {
			parts = append(parts, arg.Inspect())
		}
	}
	return "<partial (" + strings.Join(parts, ", ") + ")>"
}

type Task struct {
	ResultCh chan taskResult
	done     bool
	result   Value
	err      error
}

func (t *Task) Type() ValueType { return TASK }
func (t *Task) Inspect() string { return "<task>" }

type Channel struct {
	Ch        chan Value
	Closed    bool
	closeOnce sync.Once
}

func (c *Channel) Type() ValueType { return CHANNEL }
func (c *Channel) Inspect() string { return "<channel>" }
func (c *Channel) Close() {
	c.closeOnce.Do(func() {
		c.Closed = true
		close(c.Ch)
	})
}

type taskResult struct {
	value Value
	err   error
}
