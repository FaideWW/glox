package ast

import (
	"github.com/faideww/glox/src/errors"
	"github.com/faideww/glox/src/token"
)

type Parser struct {
	tokens   []token.Token
	current  int
	errored  bool
	reporter *errors.ErrorReporter
}

func NewParser(tokens []token.Token, reporter *errors.ErrorReporter) *Parser {
	return &Parser{tokens, 0, false, reporter}
}

func (p *Parser) Parse() ([]Stmt, bool) {
	statements := make([]Stmt, 0)
	for !p.atEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return statements, !p.errored
		}
		statements = append(statements, stmt)
	}

	return statements, !p.errored
}

func (p *Parser) ParseExpression() (Expr, bool) {
	expr, err := p.expression()
	if err != nil {
		return nil, !p.errored
	}

	return expr, !p.errored
}

func (p *Parser) declaration() (Stmt, error) {
	var value Stmt
	var err error
	if p.match(token.VAR) {
		value, err = p.varDeclaration()
	} else {
		value, err = p.statement()
	}
	if err != nil {
		p.synchronize()
		return nil, err
	}

	return value, err
}

func (p *Parser) varDeclaration() (Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect variable name")
	if err != nil {
		return nil, err
	}

	var initializer Expr
	if p.match(token.EQUAL) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(token.SEMICOLON, "Expect ';' after variable declaration")
	if err != nil {
		return nil, err
	}

	return VarStmt{name, initializer}, nil
}

func (p *Parser) statement() (Stmt, error) {
	if p.match(token.PRINT) {
		return p.printStatement()
	}

	if p.match(token.LEFT_BRACE) {
		block, err := p.block()
		if err != nil {
			return nil, err
		}
		return Block{block}, nil
	}

	return p.expressionStatement()
}

func (p *Parser) printStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return expr, err
	}
	p.consume(token.SEMICOLON, "Expect ';' after value")
	return PrintStmt{expr}, nil
}

func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return expr, err
	}
	p.consume(token.SEMICOLON, "Expect ';' after value")
	return ExpressionStmt{expr}, nil
}

func (p *Parser) block() ([]Stmt, error) {
	statements := make([]Stmt, 0)
	for !p.check(token.RIGHT_BRACE) && !p.atEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	p.consume(token.RIGHT_BRACE, "Expect '}' after block.")

	return statements, nil
}

func (p *Parser) match(types ...token.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t token.TokenType) bool {
	if p.atEnd() {
		return false
	}
	return p.peek().TokenType == t
}

func (p *Parser) advance() token.Token {
	if !p.atEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) atEnd() bool {
	return p.peek().TokenType == token.EOF
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) expression() (Expr, error) {
	expr, err := p.assignment()
	if err != nil {
		return expr, err
	}

	for p.match(token.COMMA) {
		expr, err = p.assignment()
	}

	return expr, err
}

func (p *Parser) assignment() (Expr, error) {
	expr, err := p.condition()
	if err != nil {
		return nil, err
	}

	if p.match(token.EQUAL) {
		tok := p.previous()
		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		// Check if the receiving expression is an l-value
		if receiver, ok := expr.(VariableExpr); ok {
			return AssignmentExpr{receiver.name, value}, nil
		}

		return nil, p.error(tok, "Invalid assignment target")
	}

	return expr, nil
}

func (p *Parser) condition() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	if p.match(token.QMARK) {
		left, err := p.condition()
		if err != nil {
			return nil, err
		}

		if p.match(token.COLON) {
			right, err := p.condition()
			if err != nil {
				return nil, err
			}

			expr = TernaryExpr{expr, left, right}
		} else {
			return nil, p.error(p.peek(), "expected : in ternary condition")
		}
	}

	return expr, nil
}

func (p *Parser) equality() (Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = BinaryExpr{expr, operator, right}
	}

	return expr, nil
}

func (p *Parser) comparison() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = BinaryExpr{expr, operator, right}

	}

	return expr, nil
}

func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}
	for p.match(token.MINUS, token.PLUS) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = BinaryExpr{expr, operator, right}

	}

	return expr, nil
}

func (p *Parser) factor() (Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(token.SLASH, token.STAR) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = BinaryExpr{expr, operator, right}

	}

	return expr, nil
}

func (p *Parser) unary() (Expr, error) {
	if p.match(token.BANG, token.MINUS) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return UnaryExpr{operator, right}, nil
	}

	return p.primary()
}

func (p *Parser) primary() (Expr, error) {
	if p.match(token.FALSE) {
		return LiteralExpr{false}, nil
	}
	if p.match(token.TRUE) {
		return LiteralExpr{true}, nil
	}
	if p.match(token.NIL) {
		return LiteralExpr{nil}, nil
	}

	if p.match(token.NUMBER, token.STRING) {
		return LiteralExpr{p.previous().Literal}, nil
	}

	if p.match(token.IDENTIFIER) {
		return VariableExpr{p.previous()}, nil
	}

	if p.match(token.LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after expression.")
		if err != nil {
			return nil, err
		}
		return GroupingExpr{expr}, nil
	}

	return nil, p.error(p.peek(), "expected expression")
}

func (p *Parser) consume(expect token.TokenType, err string) (token.Token, error) {
	if p.check(expect) {
		return p.advance(), nil
	}
	return token.Token{}, p.error(p.peek(), err)

}

func (p *Parser) error(t token.Token, message string) error {
	p.errored = true
	err := errors.NewParserError(t, message)
	p.reporter.Collect(err)
	return err
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.atEnd() {
		if p.previous().TokenType == token.SEMICOLON {
			return
		}

		switch p.peek().TokenType {
		case token.CLASS:
			fallthrough
		case token.FUN:
			fallthrough
		case token.VAR:
			fallthrough
		case token.FOR:
			fallthrough
		case token.IF:
			fallthrough
		case token.WHILE:
			fallthrough
		case token.PRINT:
			fallthrough
		case token.RETURN:
			return
		}

		p.advance()
	}
}
