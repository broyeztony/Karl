package tests

import (
	"karl/lexer"
	"karl/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `
let add = (x, y) -> x + y
let s = "hi"
let c = 'a'
let r = 1..10 step 2
x eqv y
x += 1
x++
arr[1..]
obj.field
jsonDecode("{}") ? { foo: "bar", }
& taskA()
!& { taskA(), taskB() }
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.ARROW, "->"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},

		{token.LET, "let"},
		{token.IDENT, "s"},
		{token.ASSIGN, "="},
		{token.STRING, "hi"},

		{token.LET, "let"},
		{token.IDENT, "c"},
		{token.ASSIGN, "="},
		{token.CHAR, "a"},

		{token.LET, "let"},
		{token.IDENT, "r"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.DOTDOT, ".."},
		{token.INT, "10"},
		{token.IDENT, "step"},
		{token.INT, "2"},

		{token.IDENT, "x"},
		{token.EQV, "eqv"},
		{token.IDENT, "y"},

		{token.IDENT, "x"},
		{token.PLUS_ASSIGN, "+="},
		{token.INT, "1"},

		{token.IDENT, "x"},
		{token.INCREMENT, "++"},

		{token.IDENT, "arr"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.DOTDOT, ".."},
		{token.RBRACKET, "]"},

		{token.IDENT, "obj"},
		{token.DOT, "."},
		{token.IDENT, "field"},

		{token.IDENT, "jsonDecode"},
		{token.LPAREN, "("},
		{token.STRING, "{}"},
		{token.RPAREN, ")"},
		{token.QUESTION, "?"},
		{token.LBRACE, "{"},
		{token.IDENT, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.COMMA, ","},
		{token.RBRACE, "}"},

		{token.AMPERSAND, "&"},
		{token.IDENT, "taskA"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},

		{token.RACE, "!&"},
		{token.LBRACE, "{"},
		{token.IDENT, "taskA"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.COMMA, ","},
		{token.IDENT, "taskB"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.RBRACE, "}"},

		{token.EOF, ""},
	}

	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestBangNotEqAndRaceTokens(t *testing.T) {
	input := `
!flag
a != b
!& { taskA(), taskB() }
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.BANG, "!"},
		{token.IDENT, "flag"},

		{token.IDENT, "a"},
		{token.NOT_EQ, "!="},
		{token.IDENT, "b"},

		{token.RACE, "!&"},
		{token.LBRACE, "{"},
		{token.IDENT, "taskA"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.COMMA, ","},
		{token.IDENT, "taskB"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.RBRACE, "}"},

		{token.EOF, ""},
	}

	l := lexer.New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStringAndCharEscapes(t *testing.T) {
	input := `
let s1 = "line1\nline2"
let s2 = "quote: \" backslash: \\"
let c1 = '\n'
let c2 = '\u0041'
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "s1"},
		{token.ASSIGN, "="},
		{token.STRING, "line1\nline2"},

		{token.LET, "let"},
		{token.IDENT, "s2"},
		{token.ASSIGN, "="},
		{token.STRING, "quote: \" backslash: \\"},

		{token.LET, "let"},
		{token.IDENT, "c1"},
		{token.ASSIGN, "="},
		{token.CHAR, "\n"},

		{token.LET, "let"},
		{token.IDENT, "c2"},
		{token.ASSIGN, "="},
		{token.CHAR, "A"},

		{token.EOF, ""},
	}

	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
