package tests

import (
	"karl/ast"
	"karl/lexer"
	"karl/parser"
	"testing"
)

func TestObjectDisambiguation(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		wantObject bool
	}{
		{
			name:       "object_with_colon",
			input:      `let o = { x: 1, y: 2 }`,
			wantObject: true,
		},
		{
			name:       "object_with_spread",
			input:      `let o = { ...other }`,
			wantObject: true,
		},
		{
			name:       "object_with_trailing_comma",
			input:      `let o = { x, y, }`,
			wantObject: true,
		},
		{
			name:       "empty_object",
			input:      `let o = { }`,
			wantObject: true,
		},
		{
			name:       "object_in_parens_with_trailing_comma",
			input:      `let o = ({ x, y, })`,
			wantObject: true,
		},
		{
			name:       "block_expression",
			input:      `let o = { x }`,
			wantObject: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.input))
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt, ok := program.Statements[0].(*ast.LetStatement)
			if !ok {
				t.Fatalf("expected LetStatement, got %T", program.Statements[0])
			}
			switch stmt.Value.(type) {
			case *ast.ObjectLiteral:
				if !tc.wantObject {
					t.Fatalf("expected block, got object")
				}
			case *ast.BlockExpression:
				if tc.wantObject {
					t.Fatalf("expected object, got block")
				}
			default:
				t.Fatalf("unexpected expression type %T", stmt.Value)
			}
		})
	}
}
