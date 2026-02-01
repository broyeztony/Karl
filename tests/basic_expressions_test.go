package tests

import (
	"testing"

	"karl/ast"
)

func TestExpressionKinds(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		assert func(t *testing.T, expr ast.Expression)
	}{
		{
			name:  "identifier",
			input: "foo",
			assert: func(t *testing.T, expr ast.Expression) {
				ident, ok := expr.(*ast.Identifier)
				if !ok {
					t.Fatalf("expected Identifier, got %T", expr)
				}
				if ident.Value != "foo" {
					t.Fatalf("expected foo, got %q", ident.Value)
				}
			},
		},
		{
			name:  "placeholder",
			input: "_",
			assert: func(t *testing.T, expr ast.Expression) {
				if _, ok := expr.(*ast.Placeholder); !ok {
					t.Fatalf("expected Placeholder, got %T", expr)
				}
			},
		},
		{
			name:  "integer_literal",
			input: "42",
			assert: func(t *testing.T, expr ast.Expression) {
				lit, ok := expr.(*ast.IntegerLiteral)
				if !ok {
					t.Fatalf("expected IntegerLiteral, got %T", expr)
				}
				if lit.Value != 42 {
					t.Fatalf("expected 42, got %d", lit.Value)
				}
			},
		},
		{
			name:  "float_literal",
			input: "3.14",
			assert: func(t *testing.T, expr ast.Expression) {
				lit, ok := expr.(*ast.FloatLiteral)
				if !ok {
					t.Fatalf("expected FloatLiteral, got %T", expr)
				}
				if lit.Value != 3.14 {
					t.Fatalf("expected 3.14, got %f", lit.Value)
				}
			},
		},
		{
			name:  "string_literal",
			input: `"hi"`,
			assert: func(t *testing.T, expr ast.Expression) {
				lit, ok := expr.(*ast.StringLiteral)
				if !ok {
					t.Fatalf("expected StringLiteral, got %T", expr)
				}
				if lit.Value != "hi" {
					t.Fatalf("expected hi, got %q", lit.Value)
				}
			},
		},
		{
			name:  "char_literal",
			input: `'a'`,
			assert: func(t *testing.T, expr ast.Expression) {
				lit, ok := expr.(*ast.CharLiteral)
				if !ok {
					t.Fatalf("expected CharLiteral, got %T", expr)
				}
				if lit.Value != "a" {
					t.Fatalf("expected a, got %q", lit.Value)
				}
			},
		},
		{
			name:  "boolean_literal",
			input: "true",
			assert: func(t *testing.T, expr ast.Expression) {
				lit, ok := expr.(*ast.BooleanLiteral)
				if !ok {
					t.Fatalf("expected BooleanLiteral, got %T", expr)
				}
				if !lit.Value {
					t.Fatalf("expected true, got false")
				}
			},
		},
		{
			name:  "null_literal",
			input: "null",
			assert: func(t *testing.T, expr ast.Expression) {
				if _, ok := expr.(*ast.NullLiteral); !ok {
					t.Fatalf("expected NullLiteral, got %T", expr)
				}
			},
		},
		{
			name:  "unit_literal",
			input: "()",
			assert: func(t *testing.T, expr ast.Expression) {
				if _, ok := expr.(*ast.UnitLiteral); !ok {
					t.Fatalf("expected UnitLiteral, got %T", expr)
				}
			},
		},
		{
			name:  "prefix_expression",
			input: "!true",
			assert: func(t *testing.T, expr ast.Expression) {
				pe, ok := expr.(*ast.PrefixExpression)
				if !ok {
					t.Fatalf("expected PrefixExpression, got %T", expr)
				}
				if pe.Operator != "!" {
					t.Fatalf("expected !, got %q", pe.Operator)
				}
				if _, ok := pe.Right.(*ast.BooleanLiteral); !ok {
					t.Fatalf("expected BooleanLiteral, got %T", pe.Right)
				}
			},
		},
		{
			name:  "infix_expression",
			input: "1 + 2",
			assert: func(t *testing.T, expr ast.Expression) {
				ie, ok := expr.(*ast.InfixExpression)
				if !ok {
					t.Fatalf("expected InfixExpression, got %T", expr)
				}
				if ie.Operator != "+" {
					t.Fatalf("expected +, got %q", ie.Operator)
				}
			},
		},
		{
			name:  "equivalence_expression",
			input: "a eqv b",
			assert: func(t *testing.T, expr ast.Expression) {
				ie, ok := expr.(*ast.InfixExpression)
				if !ok {
					t.Fatalf("expected InfixExpression, got %T", expr)
				}
				if ie.Operator != "eqv" {
					t.Fatalf("expected eqv, got %q", ie.Operator)
				}
			},
		},
		{
			name:  "assign_expression",
			input: "x = 1",
			assert: func(t *testing.T, expr ast.Expression) {
				ae, ok := expr.(*ast.AssignExpression)
				if !ok {
					t.Fatalf("expected AssignExpression, got %T", expr)
				}
				if ae.Operator != "=" {
					t.Fatalf("expected =, got %q", ae.Operator)
				}
			},
		},
		{
			name:  "compound_assign_expression",
			input: "x += 1",
			assert: func(t *testing.T, expr ast.Expression) {
				ae, ok := expr.(*ast.AssignExpression)
				if !ok {
					t.Fatalf("expected AssignExpression, got %T", expr)
				}
				if ae.Operator != "+=" {
					t.Fatalf("expected +=, got %q", ae.Operator)
				}
			},
		},
		{
			name:  "postfix_increment",
			input: "i++",
			assert: func(t *testing.T, expr ast.Expression) {
				pe, ok := expr.(*ast.PostfixExpression)
				if !ok {
					t.Fatalf("expected PostfixExpression, got %T", expr)
				}
				if pe.Operator != "++" {
					t.Fatalf("expected ++, got %q", pe.Operator)
				}
			},
		},
		{
			name:  "postfix_decrement",
			input: "i--",
			assert: func(t *testing.T, expr ast.Expression) {
				pe, ok := expr.(*ast.PostfixExpression)
				if !ok {
					t.Fatalf("expected PostfixExpression, got %T", expr)
				}
				if pe.Operator != "--" {
					t.Fatalf("expected --, got %q", pe.Operator)
				}
			},
		},
		{
			name:  "wait_expression",
			input: "wait task()",
			assert: func(t *testing.T, expr ast.Expression) {
				aw, ok := expr.(*ast.AwaitExpression)
				if !ok {
					t.Fatalf("expected AwaitExpression, got %T", expr)
				}
				if _, ok := aw.Value.(*ast.CallExpression); !ok {
					t.Fatalf("expected CallExpression, got %T", aw.Value)
				}
			},
		},
		{
			name:  "import_expression",
			input: `import "examples/features/import_module.k"`,
			assert: func(t *testing.T, expr ast.Expression) {
				ie, ok := expr.(*ast.ImportExpression)
				if !ok {
					t.Fatalf("expected ImportExpression, got %T", expr)
				}
				if ie.Path == nil || ie.Path.Value != "examples/features/import_module.k" {
					t.Fatalf("expected import path, got %v", ie.Path)
				}
			},
		},
		{
			name:  "if_expression",
			input: "if ok { 1 } else { 0 }",
			assert: func(t *testing.T, expr ast.Expression) {
				ie, ok := expr.(*ast.IfExpression)
				if !ok {
					t.Fatalf("expected IfExpression, got %T", expr)
				}
				if ie.Consequence == nil || ie.Alternative == nil {
					t.Fatalf("expected both consequence and alternative")
				}
			},
		},
		{
			name:  "match_expression",
			input: "match x { case 1 -> 1 case _ -> 0 }",
			assert: func(t *testing.T, expr ast.Expression) {
				me, ok := expr.(*ast.MatchExpression)
				if !ok {
					t.Fatalf("expected MatchExpression, got %T", expr)
				}
				if len(me.Arms) != 2 {
					t.Fatalf("expected 2 arms, got %d", len(me.Arms))
				}
			},
		},
		{
			name:  "for_expression",
			input: "for i < 10 with i = 0 { i++ } then i",
			assert: func(t *testing.T, expr ast.Expression) {
				fe, ok := expr.(*ast.ForExpression)
				if !ok {
					t.Fatalf("expected ForExpression, got %T", expr)
				}
				if len(fe.Bindings) != 1 {
					t.Fatalf("expected 1 binding, got %d", len(fe.Bindings))
				}
				if fe.Then == nil {
					t.Fatalf("expected then expression")
				}
			},
		},
		{
			name:  "for_then_block",
			input: "for i < 3 with i = 0 { i++ } then { let x = i x }",
			assert: func(t *testing.T, expr ast.Expression) {
				fe, ok := expr.(*ast.ForExpression)
				if !ok {
					t.Fatalf("expected ForExpression, got %T", expr)
				}
				if _, ok := fe.Then.(*ast.BlockExpression); !ok {
					t.Fatalf("expected BlockExpression, got %T", fe.Then)
				}
			},
		},
		{
			name:  "lambda_expression",
			input: "let add = (a, b) -> a + b",
			assert: func(t *testing.T, expr ast.Expression) {
				le, ok := expr.(*ast.LambdaExpression)
				if !ok {
					t.Fatalf("expected LambdaExpression, got %T", expr)
				}
				if len(le.Params) != 2 {
					t.Fatalf("expected 2 params, got %d", len(le.Params))
				}
			},
		},
		{
			name:  "call_expression",
			input: "add(1, 2)",
			assert: func(t *testing.T, expr ast.Expression) {
				ce, ok := expr.(*ast.CallExpression)
				if !ok {
					t.Fatalf("expected CallExpression, got %T", expr)
				}
				if len(ce.Arguments) != 2 {
					t.Fatalf("expected 2 args, got %d", len(ce.Arguments))
				}
			},
		},
		{
			name:  "call_expression_trailing_comma",
			input: "add(1, 2,)",
			assert: func(t *testing.T, expr ast.Expression) {
				ce, ok := expr.(*ast.CallExpression)
				if !ok {
					t.Fatalf("expected CallExpression, got %T", expr)
				}
				if len(ce.Arguments) != 2 {
					t.Fatalf("expected 2 args, got %d", len(ce.Arguments))
				}
			},
		},
		{
			name:  "recover_expression",
			input: `decodeJson("{}") ? { foo: "bar", }`,
			assert: func(t *testing.T, expr ast.Expression) {
				re, ok := expr.(*ast.RecoverExpression)
				if !ok {
					t.Fatalf("expected RecoverExpression, got %T", expr)
				}
				if _, ok := re.Target.(*ast.CallExpression); !ok {
					t.Fatalf("expected CallExpression target, got %T", re.Target)
				}
			},
		},
		{
			name:  "member_expression",
			input: "user.profile",
			assert: func(t *testing.T, expr ast.Expression) {
				me, ok := expr.(*ast.MemberExpression)
				if !ok {
					t.Fatalf("expected MemberExpression, got %T", expr)
				}
				if me.Property.Value != "profile" {
					t.Fatalf("expected profile, got %q", me.Property.Value)
				}
			},
		},
		{
			name:  "member_expression_keyword",
			input: "task.then",
			assert: func(t *testing.T, expr ast.Expression) {
				me, ok := expr.(*ast.MemberExpression)
				if !ok {
					t.Fatalf("expected MemberExpression, got %T", expr)
				}
				if me.Property.Value != "then" {
					t.Fatalf("expected then, got %q", me.Property.Value)
				}
			},
		},
		{
			name:  "index_expression",
			input: "arr[0]",
			assert: func(t *testing.T, expr ast.Expression) {
				ie, ok := expr.(*ast.IndexExpression)
				if !ok {
					t.Fatalf("expected IndexExpression, got %T", expr)
				}
				if _, ok := ie.Index.(*ast.IntegerLiteral); !ok {
					t.Fatalf("expected IntegerLiteral, got %T", ie.Index)
				}
			},
		},
		{
			name:  "object_index_expression",
			input: `obj["a-field"]`,
			assert: func(t *testing.T, expr ast.Expression) {
				ie, ok := expr.(*ast.IndexExpression)
				if !ok {
					t.Fatalf("expected IndexExpression, got %T", expr)
				}
				if _, ok := ie.Index.(*ast.StringLiteral); !ok {
					t.Fatalf("expected StringLiteral, got %T", ie.Index)
				}
			},
		},
		{
			name:  "object_index_assignment_expression",
			input: `obj["a-field"] = 1`,
			assert: func(t *testing.T, expr ast.Expression) {
				ae, ok := expr.(*ast.AssignExpression)
				if !ok {
					t.Fatalf("expected AssignExpression, got %T", expr)
				}
				left, ok := ae.Left.(*ast.IndexExpression)
				if !ok {
					t.Fatalf("expected IndexExpression lhs, got %T", ae.Left)
				}
				if _, ok := left.Index.(*ast.StringLiteral); !ok {
					t.Fatalf("expected StringLiteral index, got %T", left.Index)
				}
			},
		},
		{
			name:  "slice_expression",
			input: "arr[..3]",
			assert: func(t *testing.T, expr ast.Expression) {
				se, ok := expr.(*ast.SliceExpression)
				if !ok {
					t.Fatalf("expected SliceExpression, got %T", expr)
				}
				if se.Start != nil {
					t.Fatalf("expected nil start, got %T", se.Start)
				}
				if se.End == nil {
					t.Fatalf("expected end expression")
				}
			},
		},
		{
			name:  "array_literal",
			input: "[1, 2]",
			assert: func(t *testing.T, expr ast.Expression) {
				al, ok := expr.(*ast.ArrayLiteral)
				if !ok {
					t.Fatalf("expected ArrayLiteral, got %T", expr)
				}
				if len(al.Elements) != 2 {
					t.Fatalf("expected 2 elements, got %d", len(al.Elements))
				}
			},
		},
		{
			name:  "array_literal_trailing_comma",
			input: "[1, 2,]",
			assert: func(t *testing.T, expr ast.Expression) {
				al, ok := expr.(*ast.ArrayLiteral)
				if !ok {
					t.Fatalf("expected ArrayLiteral, got %T", expr)
				}
				if len(al.Elements) != 2 {
					t.Fatalf("expected 2 elements, got %d", len(al.Elements))
				}
			},
		},
		{
			name:  "object_literal",
			input: "{ x: 1, y }",
			assert: func(t *testing.T, expr ast.Expression) {
				ol, ok := expr.(*ast.ObjectLiteral)
				if !ok {
					t.Fatalf("expected ObjectLiteral, got %T", expr)
				}
				if len(ol.Entries) != 2 {
					t.Fatalf("expected 2 entries, got %d", len(ol.Entries))
				}
				if ol.Entries[1].Shorthand != true {
					t.Fatalf("expected shorthand entry")
				}
			},
		},
		{
			name:  "object_spread",
			input: "{ ...other, x: 1 }",
			assert: func(t *testing.T, expr ast.Expression) {
				ol, ok := expr.(*ast.ObjectLiteral)
				if !ok {
					t.Fatalf("expected ObjectLiteral, got %T", expr)
				}
				if len(ol.Entries) != 2 {
					t.Fatalf("expected 2 entries, got %d", len(ol.Entries))
				}
				if !ol.Entries[0].Spread {
					t.Fatalf("expected spread entry")
				}
			},
		},
		{
			name:  "block_expression",
			input: "let block = { let x = 1 x }",
			assert: func(t *testing.T, expr ast.Expression) {
				if _, ok := expr.(*ast.BlockExpression); !ok {
					t.Fatalf("expected BlockExpression, got %T", expr)
				}
			},
		},
		{
			name:  "struct_init_expression",
			input: "User { name: \"tony\" }",
			assert: func(t *testing.T, expr ast.Expression) {
				se, ok := expr.(*ast.StructInitExpression)
				if !ok {
					t.Fatalf("expected StructInitExpression, got %T", expr)
				}
				if se.TypeName.Value != "User" {
					t.Fatalf("expected User, got %q", se.TypeName.Value)
				}
			},
		},
		{
			name:  "range_expression",
			input: "1..10 step 2",
			assert: func(t *testing.T, expr ast.Expression) {
				re, ok := expr.(*ast.RangeExpression)
				if !ok {
					t.Fatalf("expected RangeExpression, got %T", expr)
				}
				if re.Step == nil {
					t.Fatalf("expected step expression")
				}
			},
		},
		{
			name:  "query_expression",
			input: "from x in items where x > 0 orderby x select x",
			assert: func(t *testing.T, expr ast.Expression) {
				qe, ok := expr.(*ast.QueryExpression)
				if !ok {
					t.Fatalf("expected QueryExpression, got %T", expr)
				}
				if qe.Var.Value != "x" {
					t.Fatalf("expected x, got %q", qe.Var.Value)
				}
				if qe.Select == nil {
					t.Fatalf("expected select expression")
				}
			},
		},
		{
			name:  "race_expression",
			input: "| { taskA(), taskB() }",
			assert: func(t *testing.T, expr ast.Expression) {
				re, ok := expr.(*ast.RaceExpression)
				if !ok {
					t.Fatalf("expected RaceExpression, got %T", expr)
				}
				if len(re.Tasks) != 2 {
					t.Fatalf("expected 2 tasks, got %d", len(re.Tasks))
				}
			},
		},
		{
			name:  "spawn_expression",
			input: "& taskA()",
			assert: func(t *testing.T, expr ast.Expression) {
				se, ok := expr.(*ast.SpawnExpression)
				if !ok {
					t.Fatalf("expected SpawnExpression, got %T", expr)
				}
				if se.Task == nil {
					t.Fatalf("expected task expression")
				}
			},
		},
		{
			name:  "break_expression",
			input: "break 1",
			assert: func(t *testing.T, expr ast.Expression) {
				be, ok := expr.(*ast.BreakExpression)
				if !ok {
					t.Fatalf("expected BreakExpression, got %T", expr)
				}
				if be.Value == nil {
					t.Fatalf("expected break value")
				}
			},
		},
		{
			name:  "continue_expression",
			input: "continue",
			assert: func(t *testing.T, expr ast.Expression) {
				if _, ok := expr.(*ast.ContinueExpression); !ok {
					t.Fatalf("expected ContinueExpression, got %T", expr)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr := parseExpression(t, tc.input)
			tc.assert(t, expr)
		})
	}
}

func TestPatternKinds(t *testing.T) {
	input := `match value {
		case _ -> 0
		case 1..10 -> 1
		case { kind: "data", value } -> 2
		case [a, b] -> 3
		case [head, ...tail] -> 4
		case (x, y) -> 5
		case true -> 6
		case "done" -> 7
		case 'a' -> 8
	}`
	program := parseProgram(t, input)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	match, ok := stmt.Expression.(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expected MatchExpression, got %T", stmt.Expression)
	}
	if len(match.Arms) != 9 {
		t.Fatalf("expected 9 arms, got %d", len(match.Arms))
	}

	if _, ok := match.Arms[0].Pattern.(*ast.WildcardPattern); !ok {
		t.Fatalf("expected WildcardPattern, got %T", match.Arms[0].Pattern)
	}
	if _, ok := match.Arms[1].Pattern.(*ast.RangePattern); !ok {
		t.Fatalf("expected RangePattern, got %T", match.Arms[1].Pattern)
	}
	if _, ok := match.Arms[2].Pattern.(*ast.ObjectPattern); !ok {
		t.Fatalf("expected ObjectPattern, got %T", match.Arms[2].Pattern)
	}
	if _, ok := match.Arms[3].Pattern.(*ast.ArrayPattern); !ok {
		t.Fatalf("expected ArrayPattern, got %T", match.Arms[3].Pattern)
	}
	arrayWithRest, ok := match.Arms[4].Pattern.(*ast.ArrayPattern)
	if !ok {
		t.Fatalf("expected ArrayPattern, got %T", match.Arms[4].Pattern)
	}
	if arrayWithRest.Rest == nil {
		t.Fatalf("expected rest pattern")
	}
	if _, ok := match.Arms[5].Pattern.(*ast.TuplePattern); !ok {
		t.Fatalf("expected TuplePattern, got %T", match.Arms[5].Pattern)
	}
	if _, ok := match.Arms[6].Pattern.(*ast.BooleanLiteral); !ok {
		t.Fatalf("expected BooleanLiteral pattern, got %T", match.Arms[6].Pattern)
	}
	if _, ok := match.Arms[7].Pattern.(*ast.StringLiteral); !ok {
		t.Fatalf("expected StringLiteral pattern, got %T", match.Arms[7].Pattern)
	}
	if _, ok := match.Arms[8].Pattern.(*ast.CharLiteral); !ok {
		t.Fatalf("expected CharLiteral pattern, got %T", match.Arms[8].Pattern)
	}
}

func TestMatchGuardParsesWithoutLambda(t *testing.T) {
	input := `match sum_divs { case _ if sum_divs > n -> 1 case _ -> 0 }`
	program := parseProgram(t, input)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	match, ok := stmt.Expression.(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expected MatchExpression, got %T", stmt.Expression)
	}
	if len(match.Arms) != 2 {
		t.Fatalf("expected 2 arms, got %d", len(match.Arms))
	}
	if match.Arms[0].Guard == nil {
		t.Fatalf("expected guard expression")
	}
	guard, ok := match.Arms[0].Guard.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression guard, got %T", match.Arms[0].Guard)
	}
	if guard.Operator != ">" {
		t.Fatalf("expected '>' guard operator, got %q", guard.Operator)
	}
	right, ok := guard.Right.(*ast.Identifier)
	if !ok || right.Value != "n" {
		t.Fatalf("expected guard right identifier 'n', got %T", guard.Right)
	}
}

func TestMatchGuardNestedMatch(t *testing.T) {
	input := `match n { case _ if match n % 2 { case 0 -> true case _ -> false } -> "even" case _ -> "odd" }`
	program := parseProgram(t, input)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	match, ok := stmt.Expression.(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expected MatchExpression, got %T", stmt.Expression)
	}
	if len(match.Arms) != 2 {
		t.Fatalf("expected 2 arms, got %d", len(match.Arms))
	}
	if match.Arms[0].Guard == nil {
		t.Fatalf("expected guard expression")
	}
	if _, ok := match.Arms[0].Guard.(*ast.MatchExpression); !ok {
		t.Fatalf("expected MatchExpression guard, got %T", match.Arms[0].Guard)
	}
}

func TestPatternLetObjectDestructure(t *testing.T) {
	input := `
let { a, b } = foo
let { a, b, } = foo
let { a: a1, b: b1, } = foo
`
	program := parseProgram(t, input)
	if len(program.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(program.Statements))
	}

	for i, stmt := range program.Statements {
		letStmt, ok := stmt.(*ast.LetStatement)
		if !ok {
			t.Fatalf("expected LetStatement at %d, got %T", i, stmt)
		}
		obj, ok := letStmt.Name.(*ast.ObjectPattern)
		if !ok {
			t.Fatalf("expected ObjectPattern at %d, got %T", i, letStmt.Name)
		}
		if len(obj.Entries) != 2 {
			t.Fatalf("expected 2 object entries at %d, got %d", i, len(obj.Entries))
		}
	}
}

func TestPatternTrailingComma(t *testing.T) {
	input := `
let { a, b, } = foo
let [c, d, ] = bar
let [head, ...tail, ] = baz
let (x, y, ) = pair
`
	program := parseProgram(t, input)
	if len(program.Statements) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(program.Statements))
	}

	stmt0, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", program.Statements[0])
	}
	obj, ok := stmt0.Name.(*ast.ObjectPattern)
	if !ok {
		t.Fatalf("expected ObjectPattern, got %T", stmt0.Name)
	}
	if len(obj.Entries) != 2 {
		t.Fatalf("expected 2 object entries, got %d", len(obj.Entries))
	}

	stmt1, ok := program.Statements[1].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", program.Statements[1])
	}
	arr, ok := stmt1.Name.(*ast.ArrayPattern)
	if !ok {
		t.Fatalf("expected ArrayPattern, got %T", stmt1.Name)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 array elements, got %d", len(arr.Elements))
	}

	stmt2, ok := program.Statements[2].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", program.Statements[2])
	}
	arrRest, ok := stmt2.Name.(*ast.ArrayPattern)
	if !ok {
		t.Fatalf("expected ArrayPattern, got %T", stmt2.Name)
	}
	if arrRest.Rest == nil {
		t.Fatalf("expected rest pattern")
	}

	stmt3, ok := program.Statements[3].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", program.Statements[3])
	}
	tuple, ok := stmt3.Name.(*ast.TuplePattern)
	if !ok {
		t.Fatalf("expected TuplePattern, got %T", stmt3.Name)
	}
	if len(tuple.Elements) != 2 {
		t.Fatalf("expected 2 tuple elements, got %d", len(tuple.Elements))
	}
}
