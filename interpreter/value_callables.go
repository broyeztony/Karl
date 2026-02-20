package interpreter

import (
	"karl/ast"
	"strings"
)

type Function struct {
	Name   string
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
