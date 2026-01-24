package parser

import (
	"fmt"
	"karl/ast"
	"karl/lexer"
	"karl/token"
	"strconv"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []ParseError

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

const (
	_ int = iota
	LOWEST
	ASSIGN
	OR
	AND
	EQUALS
	LESSGREATER
	RANGE
	SUM
	PRODUCT
	PREFIX
	POSTFIX
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:          ASSIGN,
	token.PLUS_ASSIGN:     ASSIGN,
	token.MINUS_ASSIGN:    ASSIGN,
	token.ASTERISK_ASSIGN: ASSIGN,
	token.SLASH_ASSIGN:    ASSIGN,
	token.PERCENT_ASSIGN:  ASSIGN,
	token.OR:              OR,
	token.AND:             AND,
	token.EQ:              EQUALS,
	token.NOT_EQ:          EQUALS,
	token.EQV:             EQUALS,
	token.LT:              LESSGREATER,
	token.LE:              LESSGREATER,
	token.GT:              LESSGREATER,
	token.GE:              LESSGREATER,
	token.DOTDOT:          RANGE,
	token.PLUS:            SUM,
	token.MINUS:           SUM,
	token.SLASH:           PRODUCT,
	token.ASTERISK:        PRODUCT,
	token.PERCENT:         PRODUCT,
	token.LPAREN:          POSTFIX,
	token.LBRACKET:        POSTFIX,
	token.DOT:             POSTFIX,
	token.INCREMENT:       POSTFIX,
	token.DECREMENT:       POSTFIX,
	token.QUESTION:        POSTFIX,
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []ParseError{}}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.CHAR, p.parseCharLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.NULL, p.parseNull)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.WAIT, p.parseWaitExpression)
	p.registerPrefix(token.IMPORT, p.parseImportExpression)
	p.registerPrefix(token.AMPERSAND, p.parseSpawnExpression)
	p.registerPrefix(token.PIPE, p.parseRaceExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedOrLambda)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.MATCH, p.parseMatchExpression)
	p.registerPrefix(token.FOR, p.parseForExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseBraceExpression)
	p.registerPrefix(token.FROM, p.parseQueryExpression)
	p.registerPrefix(token.BREAK, p.parseBreakExpression)
	p.registerPrefix(token.CONTINUE, p.parseContinueExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.PERCENT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.EQV, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LE, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.DOTDOT, p.parseRangeExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.PLUS_ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.MINUS_ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.ASTERISK_ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.SLASH_ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.PERCENT_ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.QUESTION, p.parseRecoverExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexOrSliceExpression)
	p.registerInfix(token.INCREMENT, p.parsePostfixExpression)
	p.registerInfix(token.DECREMENT, p.parsePostfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	if len(p.errors) == 0 {
		return nil
	}
	out := make([]string, len(p.errors))
	for i, err := range p.errors {
		out[i] = err.Message
	}
	return out
}

func (p *Parser) ErrorsDetailed() []ParseError {
	return p.errors
}

func (p *Parser) addError(tok token.Token, msg string) {
	p.errors = append(p.errors, ParseError{Message: msg, Token: tok})
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	p.nextToken()
	stmt.Name = p.parsePattern()
	if stmt.Name == nil {
		return nil
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	if p.curToken.Literal == "_" {
		return &ast.Placeholder{Token: p.curToken}
	}
	if p.peekTokenIs(token.ARROW) {
		param := p.parsePatternFromToken(p.curToken)
		p.nextToken()
		return p.finishLambda([]ast.Pattern{param})
	}
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if p.peekTokenIs(token.LBRACE) && p.braceLooksLikeObject() {
		p.nextToken()
		obj := p.parseObjectLiteral().(*ast.ObjectLiteral)
		return &ast.StructInitExpression{Token: ident.Token, TypeName: ident, Value: obj}
	}
	return ident
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.addError(p.curToken, fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.addError(p.curToken, fmt.Sprintf("could not parse %q as float", p.curToken.Literal))
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseCharLiteral() ast.Expression {
	return &ast.CharLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNull() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseWaitExpression() ast.Expression {
	expr := &ast.AwaitExpression{Token: p.curToken}
	p.nextToken()
	expr.Value = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseImportExpression() ast.Expression {
	expr := &ast.ImportExpression{Token: p.curToken}
	if !p.expectPeek(token.STRING) {
		return nil
	}
	expr.Path = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	return expr
}

func (p *Parser) parseSpawnExpression() ast.Expression {
	expr := &ast.SpawnExpression{Token: p.curToken}
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		expr.Group = p.parseRaceOrSpawnGroup()
		return expr
	}
	p.nextToken()
	task := p.parseExpression(PREFIX)
	if !p.isCallExpression(task) {
		p.addError(p.curToken, "spawn target must be a call expression")
		return expr
	}
	expr.Task = task
	return expr
}

func (p *Parser) parseRaceExpression() ast.Expression {
	expr := &ast.RaceExpression{Token: p.curToken}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.Tasks = p.parseRaceOrSpawnGroup()
	return expr
}

func (p *Parser) parseRaceOrSpawnGroup() []ast.Expression {
	group := []ast.Expression{}
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return group
	}

	p.nextToken()
	for {
		task := p.parseExpression(PREFIX)
		if !p.isCallExpression(task) {
			p.addError(p.curToken, "group tasks must be call expressions")
		}
		group = append(group, task)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		break
	}

	if !p.expectPeek(token.RBRACE) {
		return group
	}

	return group
}

func (p *Parser) parseGroupedOrLambda() ast.Expression {
	if p.isLambdaParams() {
		params := p.parseLambdaParams()
		return p.finishLambda(params)
	}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return &ast.UnitLiteral{Token: p.curToken}
	}

	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockExpression()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		p.nextToken()
		if p.curTokenIs(token.IF) {
			expression.Alternative = p.parseIfExpression()
		} else if p.curTokenIs(token.LBRACE) {
			expression.Alternative = p.parseBlockExpression()
		} else {
			expression.Alternative = p.parseExpression(LOWEST)
		}
	}

	return expression
}

func (p *Parser) parseMatchExpression() ast.Expression {
	expression := &ast.MatchExpression{Token: p.curToken}

	p.nextToken()
	expression.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Arms = []ast.MatchArm{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if !p.curTokenIs(token.CASE) {
			p.addError(p.curToken, "expected case in match expression")
			return expression
		}
		arm := ast.MatchArm{Token: p.curToken}
		p.nextToken()
		arm.Pattern = p.parsePattern()

		if p.peekTokenIs(token.IF) {
			p.nextToken()
			p.nextToken()
			arm.Guard = p.parseExpression(LOWEST)
		}

		if !p.expectPeek(token.ARROW) {
			return expression
		}
		p.nextToken()
		arm.Body = p.parseExpression(LOWEST)
		expression.Arms = append(expression.Arms, arm)

		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		p.nextToken()
	}

	return expression
}

func (p *Parser) parseForExpression() ast.Expression {
	expression := &ast.ForExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.WITH) {
		p.nextToken()
		p.nextToken()
		expression.Bindings = p.parseBindings()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Body = p.parseBlockExpression()

	if p.peekTokenIs(token.THEN) {
		p.nextToken()
		p.nextToken()
		if p.curTokenIs(token.LBRACE) {
			expression.Then = p.parseBraceExpression()
		} else {
			expression.Then = p.parseExpression(LOWEST)
		}
	}

	return expression
}

func (p *Parser) parseBindings() []ast.Binding {
	bindings := []ast.Binding{}

	binding := ast.Binding{}
	binding.Pattern = p.parsePattern()
	if binding.Pattern == nil {
		return bindings
	}
	if !p.expectPeek(token.ASSIGN) {
		return bindings
	}
	p.nextToken()
	binding.Value = p.parseExpression(LOWEST)
	bindings = append(bindings, binding)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		binding := ast.Binding{}
		binding.Pattern = p.parsePattern()
		if binding.Pattern == nil {
			return bindings
		}
		if !p.expectPeek(token.ASSIGN) {
			return bindings
		}
		p.nextToken()
		binding.Value = p.parseExpression(LOWEST)
		bindings = append(bindings, binding)
	}

	return bindings
}

func (p *Parser) parseLambdaParams() []ast.Pattern {
	params := []ast.Pattern{}
	p.nextToken()

	if p.curTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	params = append(params, p.parsePattern())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		params = append(params, p.parsePattern())
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

func (p *Parser) finishLambda(params []ast.Pattern) ast.Expression {
	if !p.curTokenIs(token.ARROW) {
		if !p.expectPeek(token.ARROW) {
			return nil
		}
	}
	lambda := &ast.LambdaExpression{Token: p.curToken, Params: params}
	p.nextToken()
	lambda.Body = p.parseExpression(LOWEST)
	return lambda
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseObjectLiteral() ast.Expression {
	object := &ast.ObjectLiteral{Token: p.curToken}
	object.Entries = []ast.ObjectEntry{}

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return object
	}

	p.nextToken()
	for {
		if p.curTokenIs(token.RBRACE) {
			return object
		}
		entry := ast.ObjectEntry{Token: p.curToken}
		if p.curTokenIs(token.DOTDOTDOT) {
			p.nextToken()
			entry.Spread = true
			entry.Value = p.parseExpression(LOWEST)
			object.Entries = append(object.Entries, entry)
		} else {
			if !p.curTokenIs(token.IDENT) {
				p.addError(p.curToken, "object keys must be identifiers")
				return object
			}
			entry.Key = p.curToken.Literal
			if p.peekTokenIs(token.COLON) {
				p.nextToken()
				p.nextToken()
				entry.Value = p.parseExpression(LOWEST)
			} else {
				entry.Shorthand = true
				entry.Value = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			}
			object.Entries = append(object.Entries, entry)
		}

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			if p.curTokenIs(token.RBRACE) {
				return object
			}
			continue
		}
		break
	}

	if !p.expectPeek(token.RBRACE) {
		return object
	}
	return object
}

func (p *Parser) parseBraceExpression() ast.Expression {
	if p.braceLooksLikeObject() {
		return p.parseObjectLiteral()
	}
	return p.parseBlockExpression()
}

func (p *Parser) parseBlockExpression() *ast.BlockExpression {
	block := &ast.BlockExpression{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseQueryExpression() ast.Expression {
	qe := &ast.QueryExpression{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	qe.Var = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IN) {
		return nil
	}
	p.nextToken()
	qe.Source = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.WHERE) {
		p.nextToken()
		p.nextToken()
		qe.Where = append(qe.Where, p.parseExpression(LOWEST))
	}

	if p.peekTokenIs(token.ORDERBY) {
		p.nextToken()
		p.nextToken()
		qe.OrderBy = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.SELECT) {
		return nil
	}
	p.nextToken()
	qe.Select = p.parseExpression(LOWEST)

	return qe
}

func (p *Parser) parseBreakExpression() ast.Expression {
	expr := &ast.BreakExpression{Token: p.curToken}
	if p.peekTokenIs(token.SEMICOLON) || p.peekTokenIs(token.RBRACE) {
		return expr
	}
	p.nextToken()
	expr.Value = p.parseExpression(LOWEST)
	return expr
}

func (p *Parser) parseContinueExpression() ast.Expression {
	return &ast.ContinueExpression{Token: p.curToken}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{Token: p.curToken, Operator: p.curToken.Literal, Left: left}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	if !isAssignable(left) {
		p.addError(p.curToken, "invalid assignment target")
	}
	expression := &ast.AssignExpression{Token: p.curToken, Operator: p.curToken.Literal, Left: left}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence - 1)
	return expression
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	expression := &ast.RangeExpression{Token: p.curToken, Start: left}
	p.nextToken()
	expression.End = p.parseExpression(RANGE)
	if _, ok := left.(*ast.FloatLiteral); ok {
		p.addError(left.(*ast.FloatLiteral).Token, "float ranges are not allowed")
	}
	if _, ok := expression.End.(*ast.FloatLiteral); ok {
		p.addError(expression.End.(*ast.FloatLiteral).Token, "float ranges are not allowed")
	}
	if p.peekTokenIs(token.STEP) || (p.peekTokenIs(token.IDENT) && p.peekToken.Literal == "step") {
		p.nextToken()
		p.nextToken()
		expression.Step = p.parseExpression(LOWEST)
	}
	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseRecoverExpression(left ast.Expression) ast.Expression {
	recoverToken := p.curToken
	if _, ok := left.(*ast.CallExpression); !ok {
		p.addError(recoverToken, "recover block only allowed after call expression")
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	fallback := p.parseBraceExpression()
	return &ast.RecoverExpression{Token: recoverToken, Target: left, Fallback: fallback}
}

func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	if !(p.peekTokenIs(token.IDENT) || p.peekTokenIs(token.THEN)) {
		p.addError(p.peekToken, "expected member name after '.'")
		return nil
	}
	p.nextToken()
	property := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return &ast.MemberExpression{Token: p.curToken, Object: left, Property: property}
}

func (p *Parser) parseIndexOrSliceExpression(left ast.Expression) ast.Expression {
	startToken := p.curToken
	p.nextToken()
	if p.curTokenIs(token.DOTDOT) {
		var end ast.Expression
		if !p.peekTokenIs(token.RBRACKET) {
			p.nextToken()
			end = p.parseExpression(LOWEST)
		}
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		return &ast.SliceExpression{Token: startToken, Left: left, Start: nil, End: end}
	}

	start := p.parseExpression(RANGE)
	if p.peekTokenIs(token.DOTDOT) {
		p.nextToken()
		var end ast.Expression
		if !p.peekTokenIs(token.RBRACKET) {
			p.nextToken()
			end = p.parseExpression(LOWEST)
		}
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		return &ast.SliceExpression{Token: startToken, Left: left, Start: start, End: end}
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return &ast.IndexExpression{Token: startToken, Left: left, Index: start}
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	return &ast.PostfixExpression{Token: p.curToken, Left: left, Operator: p.curToken.Literal}
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if p.peekTokenIs(end) {
			p.nextToken()
			return list
		}
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parsePattern() ast.Pattern {
	return p.parsePatternFromToken(p.curToken)
}

func (p *Parser) parsePatternFromToken(tok token.Token) ast.Pattern {
	switch tok.Type {
	case token.IDENT:
		if tok.Literal == "_" {
			return &ast.WildcardPattern{Token: tok}
		}
		return &ast.Identifier{Token: tok, Value: tok.Literal}
	case token.INT:
		p.curToken = tok
		lit := p.parseIntegerLiteral().(*ast.IntegerLiteral)
		if p.peekTokenIs(token.DOTDOT) {
			p.nextToken()
			p.nextToken()
			end := p.parsePattern()
			return &ast.RangePattern{Token: tok, Start: lit, End: end}
		}
		return lit
	case token.FLOAT:
		p.curToken = tok
		lit := p.parseFloatLiteral().(*ast.FloatLiteral)
		if p.peekTokenIs(token.DOTDOT) {
			p.nextToken()
			p.nextToken()
			end := p.parsePattern()
			return &ast.RangePattern{Token: tok, Start: lit, End: end}
		}
		return lit
	case token.STRING:
		lit := &ast.StringLiteral{Token: tok, Value: tok.Literal}
		if p.peekTokenIs(token.DOTDOT) {
			p.nextToken()
			p.nextToken()
			end := p.parsePattern()
			return &ast.RangePattern{Token: tok, Start: lit, End: end}
		}
		return lit
	case token.CHAR:
		lit := &ast.CharLiteral{Token: tok, Value: tok.Literal}
		if p.peekTokenIs(token.DOTDOT) {
			p.nextToken()
			p.nextToken()
			end := p.parsePattern()
			return &ast.RangePattern{Token: tok, Start: lit, End: end}
		}
		return lit
	case token.TRUE, token.FALSE:
		return &ast.BooleanLiteral{Token: tok, Value: tok.Type == token.TRUE}
	case token.NULL:
		return &ast.NullLiteral{Token: tok}
	case token.LBRACE:
		p.curToken = tok
		return p.parseObjectPattern()
	case token.LBRACKET:
		p.curToken = tok
		return p.parseArrayPattern()
	case token.LPAREN:
		p.curToken = tok
		return p.parseTuplePattern()
	default:
		p.addError(tok, fmt.Sprintf("unexpected pattern token: %s", tok.Type))
		return nil
	}
}

func (p *Parser) parseCallPattern(tok token.Token) ast.Pattern {
	name := &ast.Identifier{Token: tok, Value: tok.Literal}
	args := []ast.Pattern{}
	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		return &ast.CallPattern{Token: tok, Name: name, Args: args}
	}
	args = append(args, p.parsePattern())
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parsePattern())
	}
	if !p.expectPeek(token.RPAREN) {
		return &ast.CallPattern{Token: tok, Name: name, Args: args}
	}
	return &ast.CallPattern{Token: tok, Name: name, Args: args}
}

func (p *Parser) parseObjectPattern() ast.Pattern {
	pattern := &ast.ObjectPattern{Token: p.curToken, Entries: []ast.PatternEntry{}}

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return pattern
	}

	p.nextToken()
	for {
		if !p.curTokenIs(token.IDENT) {
			p.addError(p.curToken, "object pattern keys must be identifiers")
			return pattern
		}
		entry := ast.PatternEntry{Token: p.curToken, Key: p.curToken.Literal}
		if p.peekTokenIs(token.COLON) {
			p.nextToken()
			p.nextToken()
			entry.Pattern = p.parsePattern()
		} else {
			entry.Pattern = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		}
		pattern.Entries = append(pattern.Entries, entry)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		break
	}

	if !p.expectPeek(token.RBRACE) {
		return pattern
	}
	return pattern
}

func (p *Parser) parseArrayPattern() ast.Pattern {
	pattern := &ast.ArrayPattern{Token: p.curToken, Elements: []ast.Pattern{}}

	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return pattern
	}

	p.nextToken()
	for {
		if p.curTokenIs(token.DOTDOTDOT) {
			p.nextToken()
			pattern.Rest = p.parsePattern()
			break
		}
		pattern.Elements = append(pattern.Elements, p.parsePattern())

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		break
	}

	if !p.expectPeek(token.RBRACKET) {
		return pattern
	}
	return pattern
}

func (p *Parser) parseTuplePattern() ast.Pattern {
	pattern := &ast.TuplePattern{Token: p.curToken, Elements: []ast.Pattern{}}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return pattern
	}

	p.nextToken()
	pattern.Elements = append(pattern.Elements, p.parsePattern())
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		pattern.Elements = append(pattern.Elements, p.parsePattern())
	}

	if !p.expectPeek(token.RPAREN) {
		return pattern
	}
	return pattern
}

func (p *Parser) isLambdaParams() bool {
	depth := 1
	peek := p.peekToken
	lcopy := *p.l
	for {
		switch peek.Type {
		case token.LPAREN:
			depth++
		case token.RPAREN:
			depth--
			if depth == 0 {
				next := lcopy.NextToken()
				return next.Type == token.ARROW
			}
		}
		peek = lcopy.NextToken()
		if peek.Type == token.EOF {
			return false
		}
	}
}

func (p *Parser) braceLooksLikeObject() bool {
	peek := p.peekToken
	lcopy := *p.l

	if peek.Type == token.LBRACE {
		peek = lcopy.NextToken()
		if peek.Type == token.RBRACE {
			// Empty literal: treat as object.
			return true
		}
	} else if peek.Type == token.RBRACE {
		// Empty literal: treat as object.
		return true
	}

	depthParen := 0
	depthBrace := 0
	depthBracket := 0
	tok := peek
	for tok.Type != token.EOF {
		if depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
			switch tok.Type {
			case token.COLON, token.DOTDOTDOT:
				return true
			case token.COMMA:
				next := lcopy.NextToken()
				if next.Type == token.RBRACE {
					return true
				}
				tok = next
				continue
			case token.RBRACE:
				return false
			}
		}

		switch tok.Type {
		case token.LPAREN:
			depthParen++
		case token.RPAREN:
			if depthParen > 0 {
				depthParen--
			}
		case token.LBRACE:
			depthBrace++
		case token.RBRACE:
			if depthBrace == 0 {
				return false
			}
			depthBrace--
		case token.LBRACKET:
			depthBracket++
		case token.RBRACKET:
			if depthBracket > 0 {
				depthBracket--
			}
		}

		tok = lcopy.NextToken()
	}
	return false
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.addError(p.peekToken, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(p.curToken, msg)
}

func (p *Parser) isCallExpression(expr ast.Expression) bool {
	_, ok := expr.(*ast.CallExpression)
	return ok
}

func isAssignable(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.Identifier, *ast.MemberExpression, *ast.IndexExpression:
		return true
	default:
		return false
	}
}
