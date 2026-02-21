package lexer

import (
	"strconv"
	"strings"

	"karl/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else if l.ch != 0 {
		l.column++
	}
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespaceAndComments()

	var tok token.Token
	startLine := l.line
	startColumn := l.column
	startOffset := l.position

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		switch l.peekChar() {
		case '+':
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.INCREMENT, Literal: literal}
		case '=':
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.PLUS_ASSIGN, Literal: literal}
		default:
			tok = newToken(token.PLUS, l.ch)
		}
	case '-':
		switch l.peekChar() {
		case '-':
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.DECREMENT, Literal: literal}
		case '=':
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.MINUS_ASSIGN, Literal: literal}
		case '>':
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.ARROW, Literal: literal}
		default:
			tok = newToken(token.MINUS, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.RACE, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.NextToken()
		}
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.SLASH_ASSIGN, Literal: literal}
		} else {
			tok = newToken(token.SLASH, l.ch)
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.ASTERISK_ASSIGN, Literal: literal}
		} else {
			tok = newToken(token.ASTERISK, l.ch)
		}
	case '%':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.PERCENT_ASSIGN, Literal: literal}
		} else {
			tok = newToken(token.PERCENT, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.LE, Literal: literal}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.GE, Literal: literal}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal}
		} else {
			tok = newToken(token.AMPERSAND, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.OR, Literal: literal}
		} else {
			tok = newToken(token.PIPE, l.ch)
		}
	case '?':
		tok = newToken(token.QUESTION, l.ch)
	case '.':
		if l.peekChar() == '.' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				literal := string(ch) + string('.') + string(l.ch)
				tok = token.Token{Type: token.DOTDOTDOT, Literal: literal}
			} else {
				literal := string(ch) + string(l.ch)
				tok = token.Token{Type: token.DOTDOT, Literal: literal}
			}
		} else {
			tok = newToken(token.DOT, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '\'':
		tok.Type = token.CHAR
		tok.Literal = l.readCharLiteral()
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Line = startLine
			tok.Column = startColumn
			tok.Offset = startOffset
			return tok
		} else if isDigit(l.ch) {
			literal, tokType := l.readNumber()
			tok.Type = tokType
			tok.Literal = literal
			tok.Line = startLine
			tok.Column = startColumn
			tok.Offset = startOffset
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	tok.Line = startLine
	tok.Column = startColumn
	tok.Offset = startOffset
	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() (string, token.TokenType) {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && l.peekChar() != '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
		return l.input[position:l.position], token.FLOAT
	}
	return l.input[position:l.position], token.INT
}

func (l *Lexer) readString() string {
	l.readChar()
	var out strings.Builder
	for l.ch != '"' && l.ch != 0 {
		if l.ch != '\\' {
			out.WriteByte(l.ch)
			l.readChar()
			continue
		}

		l.readChar()
		if l.ch == 0 {
			break
		}
		switch l.ch {
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		case 'b':
			out.WriteByte('\b')
		case 'f':
			out.WriteByte('\f')
		case '\\':
			out.WriteByte('\\')
		case '"':
			out.WriteByte('"')
		case '\'':
			out.WriteByte('\'')
		case 'u':
			r, ok := l.readUnicodeEscape()
			if ok {
				out.WriteRune(r)
			} else {
				out.WriteString("\\u")
			}
		default:
			out.WriteByte(l.ch)
		}
		l.readChar()
	}
	return out.String()
}

func (l *Lexer) readCharLiteral() string {
	l.readChar()
	if l.ch == '\\' {
		l.readChar()
		if l.ch == 0 {
			return ""
		}
		var out rune
		switch l.ch {
		case 'n':
			out = '\n'
		case 'r':
			out = '\r'
		case 't':
			out = '\t'
		case 'b':
			out = '\b'
		case 'f':
			out = '\f'
		case '\\':
			out = '\\'
		case '"':
			out = '"'
		case '\'':
			out = '\''
		case 'u':
			r, ok := l.readUnicodeEscape()
			if ok {
				out = r
			} else {
				out = 'u'
			}
		default:
			out = rune(l.ch)
		}
		l.readChar()
		return string(out)
	}
	ch := l.ch
	l.readChar()
	return string(ch)
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}

func (l *Lexer) readUnicodeEscape() (rune, bool) {
	var hex [4]byte
	for i := 0; i < 4; i++ {
		l.readChar()
		if l.ch == 0 || !isHexDigit(l.ch) {
			return 0, false
		}
		hex[i] = l.ch
	}
	value, err := strconv.ParseInt(string(hex[:]), 16, 32)
	if err != nil {
		return 0, false
	}
	return rune(value), true
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
			continue
		}
		if l.ch == '/' && l.peekChar() == '/' {
			l.skipLineComment()
			continue
		}
		break
	}
}

func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}
