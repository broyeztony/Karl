package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
	Offset  int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	FLOAT  = "FLOAT"
	STRING = "STRING"
	CHAR   = "CHAR"

	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	PERCENT  = "%"

	LT = "<"
	GT = ">"
	LE = "<="
	GE = ">="

	EQ     = "=="
	NOT_EQ = "!="
	EQV    = "EQV"

	AND = "&&"
	OR  = "||"

	PLUS_ASSIGN     = "+="
	MINUS_ASSIGN    = "-="
	ASTERISK_ASSIGN = "*="
	SLASH_ASSIGN    = "/="
	PERCENT_ASSIGN  = "%="

	INCREMENT = "++"
	DECREMENT = "--"

	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	DOT       = "."
	DOTDOT    = ".."
	DOTDOTDOT = "..."
	ARROW     = "->"
	QUESTION  = "?"
	AMPERSAND = "&"
	PIPE      = "|"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	LET      = "LET"
	IMPORT   = "IMPORT"
	IF       = "IF"
	ELSE     = "ELSE"
	MATCH    = "MATCH"
	CASE     = "CASE"
	FOR      = "FOR"
	WITH     = "WITH"
	THEN     = "THEN"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	NULL     = "NULL"
	FROM     = "FROM"
	IN       = "IN"
	WHERE    = "WHERE"
	ORDERBY  = "ORDERBY"
	SELECT   = "SELECT"
	STEP     = "STEP"
	WAIT     = "WAIT"
)

var keywords = map[string]TokenType{
	"let":      LET,
	"import":   IMPORT,
	"if":       IF,
	"else":     ELSE,
	"match":    MATCH,
	"case":     CASE,
	"for":      FOR,
	"with":     WITH,
	"then":     THEN,
	"break":    BREAK,
	"continue": CONTINUE,
	"true":     TRUE,
	"false":    FALSE,
	"null":     NULL,
	"from":     FROM,
	"in":       IN,
	"where":    WHERE,
	"orderby":  ORDERBY,
	"select":   SELECT,
	"eqv":      EQV,
	"wait":     WAIT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
