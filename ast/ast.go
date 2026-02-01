package ast

import "karl/token"

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Pattern interface {
	Node
	patternNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// Statements

type LetStatement struct {
	Token token.Token
	Name  Pattern
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// Expressions

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) patternNode()         {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

type Placeholder struct {
	Token token.Token
}

func (p *Placeholder) expressionNode()      {}
func (p *Placeholder) patternNode()         {}
func (p *Placeholder) TokenLiteral() string { return p.Token.Literal }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) patternNode()         {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) patternNode()         {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) patternNode()         {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

type CharLiteral struct {
	Token token.Token
	Value string
}

func (cl *CharLiteral) expressionNode()      {}
func (cl *CharLiteral) patternNode()         {}
func (cl *CharLiteral) TokenLiteral() string { return cl.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) patternNode()         {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }

type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) patternNode()         {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }

type UnitLiteral struct {
	Token token.Token
}

func (ul *UnitLiteral) expressionNode()      {}
func (ul *UnitLiteral) TokenLiteral() string { return ul.Token.Literal }

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

type AssignExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }

type PostfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
}

func (pe *PostfixExpression) expressionNode()      {}
func (pe *PostfixExpression) TokenLiteral() string { return pe.Token.Literal }

type AwaitExpression struct {
	Token token.Token
	Value Expression
}

func (ae *AwaitExpression) expressionNode()      {}
func (ae *AwaitExpression) TokenLiteral() string { return ae.Token.Literal }

type ImportExpression struct {
	Token token.Token
	Path  *StringLiteral
}

func (ie *ImportExpression) expressionNode()      {}
func (ie *ImportExpression) TokenLiteral() string { return ie.Token.Literal }

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockExpression
	Alternative Expression
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

type BlockExpression struct {
	Token      token.Token
	Statements []Statement
}

func (be *BlockExpression) expressionNode()      {}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Literal }

type MatchExpression struct {
	Token token.Token
	Value Expression
	Arms  []MatchArm
}

func (me *MatchExpression) expressionNode()      {}
func (me *MatchExpression) TokenLiteral() string { return me.Token.Literal }

type MatchArm struct {
	Token   token.Token
	Pattern Pattern
	Guard   Expression
	Body    Expression
}

type ForExpression struct {
	Token     token.Token
	Condition Expression
	Bindings  []Binding
	Body      *BlockExpression
	Then      Expression
}

func (fe *ForExpression) expressionNode()      {}
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }

type Binding struct {
	Pattern Pattern
	Value   Expression
}

type LambdaExpression struct {
	Token  token.Token
	Params []Pattern
	Body   Expression
}

func (le *LambdaExpression) expressionNode()      {}
func (le *LambdaExpression) TokenLiteral() string { return le.Token.Literal }

type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

type RecoverExpression struct {
	Token    token.Token
	Target   Expression
	Fallback Expression
}

func (re *RecoverExpression) expressionNode()      {}
func (re *RecoverExpression) TokenLiteral() string { return re.Token.Literal }

type MemberExpression struct {
	Token    token.Token
	Object   Expression
	Property *Identifier
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

type SliceExpression struct {
	Token token.Token
	Left  Expression
	Start Expression
	End   Expression
}

func (se *SliceExpression) expressionNode()      {}
func (se *SliceExpression) TokenLiteral() string { return se.Token.Literal }

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

type ObjectLiteral struct {
	Token   token.Token
	Entries []ObjectEntry
}

func (ol *ObjectLiteral) expressionNode()      {}
func (ol *ObjectLiteral) TokenLiteral() string { return ol.Token.Literal }

type ObjectEntry struct {
	Token     token.Token
	Key       string
	Value     Expression
	Shorthand bool
	Spread    bool
}

type StructInitExpression struct {
	Token    token.Token
	TypeName *Identifier
	Value    *ObjectLiteral
}

func (se *StructInitExpression) expressionNode()      {}
func (se *StructInitExpression) TokenLiteral() string { return se.Token.Literal }

type RangeExpression struct {
	Token token.Token
	Start Expression
	End   Expression
	Step  Expression
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Literal }

type QueryExpression struct {
	Token   token.Token
	Var     *Identifier
	Source  Expression
	Where   []Expression
	OrderBy Expression
	Select  Expression
}

func (qe *QueryExpression) expressionNode()      {}
func (qe *QueryExpression) TokenLiteral() string { return qe.Token.Literal }

type RaceExpression struct {
	Token token.Token
	Tasks []Expression
}

func (re *RaceExpression) expressionNode()      {}
func (re *RaceExpression) TokenLiteral() string { return re.Token.Literal }

type SpawnExpression struct {
	Token token.Token
	Task  Expression
	Group []Expression
}

func (se *SpawnExpression) expressionNode()      {}
func (se *SpawnExpression) TokenLiteral() string { return se.Token.Literal }

type BreakExpression struct {
	Token token.Token
	Value Expression
}

func (be *BreakExpression) expressionNode()      {}
func (be *BreakExpression) TokenLiteral() string { return be.Token.Literal }

type ContinueExpression struct {
	Token token.Token
}

func (ce *ContinueExpression) expressionNode()      {}
func (ce *ContinueExpression) TokenLiteral() string { return ce.Token.Literal }

// Patterns

type WildcardPattern struct {
	Token token.Token
}

func (wp *WildcardPattern) patternNode()         {}
func (wp *WildcardPattern) TokenLiteral() string { return wp.Token.Literal }

type RangePattern struct {
	Token token.Token
	Start Pattern
	End   Pattern
}

func (rp *RangePattern) patternNode()         {}
func (rp *RangePattern) TokenLiteral() string { return rp.Token.Literal }

type ObjectPattern struct {
	Token   token.Token
	Entries []PatternEntry
}

func (op *ObjectPattern) patternNode()         {}
func (op *ObjectPattern) TokenLiteral() string { return op.Token.Literal }

type PatternEntry struct {
	Token   token.Token
	Key     string
	Pattern Pattern
}

type ArrayPattern struct {
	Token    token.Token
	Elements []Pattern
	Rest     Pattern
}

func (ap *ArrayPattern) patternNode()         {}
func (ap *ArrayPattern) TokenLiteral() string { return ap.Token.Literal }

type TuplePattern struct {
	Token    token.Token
	Elements []Pattern
}

func (tp *TuplePattern) patternNode()         {}
func (tp *TuplePattern) TokenLiteral() string { return tp.Token.Literal }

type CallPattern struct {
	Token token.Token
	Name  *Identifier
	Args  []Pattern
}

func (cp *CallPattern) patternNode()         {}
func (cp *CallPattern) TokenLiteral() string { return cp.Token.Literal }
