package ast

import (
	"bytes"
	"fmt"
)

// Format returns a multi-line, indented view of the AST.
func Format(node Node) string {
	p := &printer{}
	p.writeNode(node)
	return p.buf.String()
}

type printer struct {
	buf    bytes.Buffer
	indent int
}

func (p *printer) line(format string, args ...interface{}) {
	for i := 0; i < p.indent; i++ {
		p.buf.WriteString("  ")
	}
	fmt.Fprintf(&p.buf, format, args...)
	p.buf.WriteByte('\n')
}

func (p *printer) writeNode(node Node) {
	switch n := node.(type) {
	case *Program:
		p.line("Program")
		p.indent++
		for _, stmt := range n.Statements {
			p.writeNode(stmt)
		}
		p.indent--
	case *LetStatement:
		p.line("LetStatement")
		p.indent++
		p.line("Name:")
		p.indent++
		p.writeNode(n.Name)
		p.indent--
		p.line("Value:")
		p.indent++
		p.writeNode(n.Value)
		p.indent--
		p.indent--
	case *ExpressionStatement:
		p.line("ExpressionStatement")
		p.indent++
		p.writeNode(n.Expression)
		p.indent--
	case *Identifier:
		p.line("Identifier(%s)", n.Value)
	case *Placeholder:
		p.line("Placeholder")
	case *IntegerLiteral:
		p.line("Integer(%d)", n.Value)
	case *FloatLiteral:
		p.line("Float(%g)", n.Value)
	case *StringLiteral:
		p.line("String(%q)", n.Value)
	case *CharLiteral:
		p.line("Char(%q)", n.Value)
	case *BooleanLiteral:
		p.line("Boolean(%t)", n.Value)
	case *NullLiteral:
		p.line("Null")
	case *UnitLiteral:
		p.line("Unit")
	case *PrefixExpression:
		p.line("Prefix(%s)", n.Operator)
		p.indent++
		p.writeNode(n.Right)
		p.indent--
	case *InfixExpression:
		p.line("Infix(%s)", n.Operator)
		p.indent++
		p.line("Left:")
		p.indent++
		p.writeNode(n.Left)
		p.indent--
		p.line("Right:")
		p.indent++
		p.writeNode(n.Right)
		p.indent--
		p.indent--
	case *AssignExpression:
		p.line("Assign(%s)", n.Operator)
		p.indent++
		p.line("Left:")
		p.indent++
		p.writeNode(n.Left)
		p.indent--
		p.line("Right:")
		p.indent++
		p.writeNode(n.Right)
		p.indent--
		p.indent--
	case *PostfixExpression:
		p.line("Postfix(%s)", n.Operator)
		p.indent++
		p.writeNode(n.Left)
		p.indent--
	case *AwaitExpression:
		p.line("Wait")
		p.indent++
		p.writeNode(n.Value)
		p.indent--
	case *ImportExpression:
		p.line("Import")
		p.indent++
		p.line("Path:")
		p.indent++
		p.writeNode(n.Path)
		p.indent--
		p.indent--
	case *RecoverExpression:
		p.line("Recover")
		p.indent++
		p.line("Target:")
		p.indent++
		p.writeNode(n.Target)
		p.indent--
		p.line("Fallback:")
		p.indent++
		p.writeNode(n.Fallback)
		p.indent--
		p.indent--
	case *IfExpression:
		p.line("If")
		p.indent++
		p.line("Condition:")
		p.indent++
		p.writeNode(n.Condition)
		p.indent--
		p.line("Then:")
		p.indent++
		p.writeNode(n.Consequence)
		p.indent--
		if n.Alternative != nil {
			p.line("Else:")
			p.indent++
			p.writeNode(n.Alternative)
			p.indent--
		}
		p.indent--
	case *BlockExpression:
		p.line("Block")
		p.indent++
		for _, stmt := range n.Statements {
			p.writeNode(stmt)
		}
		p.indent--
	case *MatchExpression:
		p.line("Match")
		p.indent++
		p.line("Value:")
		p.indent++
		p.writeNode(n.Value)
		p.indent--
		p.line("Arms:")
		p.indent++
		for _, arm := range n.Arms {
			p.line("Arm")
			p.indent++
			p.line("Pattern:")
			p.indent++
			p.writeNode(arm.Pattern)
			p.indent--
			if arm.Guard != nil {
				p.line("Guard:")
				p.indent++
				p.writeNode(arm.Guard)
				p.indent--
			}
			p.line("Body:")
			p.indent++
			p.writeNode(arm.Body)
			p.indent--
			p.indent--
		}
		p.indent--
		p.indent--
	case *ForExpression:
		p.line("For")
		p.indent++
		p.line("Condition:")
		p.indent++
		p.writeNode(n.Condition)
		p.indent--
		if len(n.Bindings) > 0 {
			p.line("Bindings:")
			p.indent++
			for _, b := range n.Bindings {
				p.line("Binding")
				p.indent++
				p.line("Pattern:")
				p.indent++
				p.writeNode(b.Pattern)
				p.indent--
				p.line("Value:")
				p.indent++
				p.writeNode(b.Value)
				p.indent--
				p.indent--
			}
			p.indent--
		}
		p.line("Body:")
		p.indent++
		p.writeNode(n.Body)
		p.indent--
		if n.Then != nil {
			p.line("Then:")
			p.indent++
			p.writeNode(n.Then)
			p.indent--
		}
		p.indent--
	case *LambdaExpression:
		p.line("Lambda")
		p.indent++
		p.line("Params:")
		p.indent++
		for _, param := range n.Params {
			p.writeNode(param)
		}
		p.indent--
		p.line("Body:")
		p.indent++
		p.writeNode(n.Body)
		p.indent--
		p.indent--
	case *CallExpression:
		p.line("Call")
		p.indent++
		p.line("Function:")
		p.indent++
		p.writeNode(n.Function)
		p.indent--
		p.line("Args:")
		p.indent++
		for _, arg := range n.Arguments {
			p.writeNode(arg)
		}
		p.indent--
		p.indent--
	case *MemberExpression:
		p.line("Member")
		p.indent++
		p.line("Object:")
		p.indent++
		p.writeNode(n.Object)
		p.indent--
		p.line("Property:")
		p.indent++
		p.writeNode(n.Property)
		p.indent--
		p.indent--
	case *IndexExpression:
		p.line("Index")
		p.indent++
		p.line("Left:")
		p.indent++
		p.writeNode(n.Left)
		p.indent--
		p.line("Index:")
		p.indent++
		p.writeNode(n.Index)
		p.indent--
		p.indent--
	case *SliceExpression:
		p.line("Slice")
		p.indent++
		p.line("Left:")
		p.indent++
		p.writeNode(n.Left)
		p.indent--
		p.line("Start:")
		p.indent++
		if n.Start != nil {
			p.writeNode(n.Start)
		} else {
			p.line("nil")
		}
		p.indent--
		p.line("End:")
		p.indent++
		if n.End != nil {
			p.writeNode(n.End)
		} else {
			p.line("nil")
		}
		p.indent--
		p.indent--
	case *ArrayLiteral:
		p.line("Array")
		p.indent++
		for _, el := range n.Elements {
			p.writeNode(el)
		}
		p.indent--
	case *ObjectLiteral:
		p.line("Object")
		p.indent++
		for _, entry := range n.Entries {
			p.line("Entry")
			p.indent++
			if entry.Spread {
				p.line("Spread")
				p.indent++
				p.writeNode(entry.Value)
				p.indent--
			} else {
				p.line("Key: %s", entry.Key)
				if entry.Shorthand {
					p.line("Shorthand")
				}
				p.line("Value:")
				p.indent++
				p.writeNode(entry.Value)
				p.indent--
			}
			p.indent--
		}
		p.indent--
	case *StructInitExpression:
		p.line("StructInit")
		p.indent++
		p.line("Type: %s", n.TypeName.Value)
		p.line("Value:")
		p.indent++
		p.writeNode(n.Value)
		p.indent--
		p.indent--
	case *RangeExpression:
		p.line("Range")
		p.indent++
		p.line("Start:")
		p.indent++
		p.writeNode(n.Start)
		p.indent--
		p.line("End:")
		p.indent++
		p.writeNode(n.End)
		p.indent--
		if n.Step != nil {
			p.line("Step:")
			p.indent++
			p.writeNode(n.Step)
			p.indent--
		}
		p.indent--
	case *QueryExpression:
		p.line("Query")
		p.indent++
		p.line("Var: %s", n.Var.Value)
		p.line("Source:")
		p.indent++
		p.writeNode(n.Source)
		p.indent--
		for _, clause := range n.Where {
			p.line("Where:")
			p.indent++
			p.writeNode(clause)
			p.indent--
		}
		if n.OrderBy != nil {
			p.line("OrderBy:")
			p.indent++
			p.writeNode(n.OrderBy)
			p.indent--
		}
		p.line("Select:")
		p.indent++
		p.writeNode(n.Select)
		p.indent--
		p.indent--
	case *RaceExpression:
		p.line("Race")
		p.indent++
		for _, task := range n.Tasks {
			p.writeNode(task)
		}
		p.indent--
	case *SpawnExpression:
		p.line("Spawn")
		p.indent++
		if n.Task != nil {
			p.line("Task:")
			p.indent++
			p.writeNode(n.Task)
			p.indent--
		}
		if len(n.Group) > 0 {
			p.line("Group:")
			p.indent++
			for _, task := range n.Group {
				p.writeNode(task)
			}
			p.indent--
		}
		p.indent--
	case *BreakExpression:
		p.line("Break")
		p.indent++
		if n.Value != nil {
			p.writeNode(n.Value)
		} else {
			p.line("nil")
		}
		p.indent--
	case *ContinueExpression:
		p.line("Continue")
	case *WildcardPattern:
		p.line("Wildcard")
	case *RangePattern:
		p.line("RangePattern")
		p.indent++
		p.line("Start:")
		p.indent++
		p.writeNode(n.Start)
		p.indent--
		p.line("End:")
		p.indent++
		p.writeNode(n.End)
		p.indent--
		p.indent--
	case *ObjectPattern:
		p.line("ObjectPattern")
		p.indent++
		for _, entry := range n.Entries {
			p.line("Entry: %s", entry.Key)
			p.indent++
			p.writeNode(entry.Pattern)
			p.indent--
		}
		p.indent--
	case *ArrayPattern:
		p.line("ArrayPattern")
		p.indent++
		for _, el := range n.Elements {
			p.writeNode(el)
		}
		if n.Rest != nil {
			p.line("Rest:")
			p.indent++
			p.writeNode(n.Rest)
			p.indent--
		}
		p.indent--
	case *TuplePattern:
		p.line("TuplePattern")
		p.indent++
		for _, el := range n.Elements {
			p.writeNode(el)
		}
		p.indent--
	case *CallPattern:
		p.line("CallPattern")
		p.indent++
		p.line("Name: %s", n.Name.Value)
		p.line("Args:")
		p.indent++
		for _, arg := range n.Args {
			p.writeNode(arg)
		}
		p.indent--
		p.indent--
	default:
		p.line("Unknown(%T)", node)
	}
}
