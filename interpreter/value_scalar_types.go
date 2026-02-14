package interpreter

import "fmt"

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
