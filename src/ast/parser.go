package ast

import (
	"fmt"

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
	if p.match(token.FUN) {
		value, err = p.function("function")
	} else if p.match(token.VAR) {
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

func (p *Parser) function(kind string) (Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, fmt.Sprintf("Expect %s name", kind))
	if err != nil {
		return nil, err
	}

	_, err = p.consume(token.LEFT_PAREN, fmt.Sprintf("Expect '(' after %s name", kind))
	if err != nil {
		return nil, err
	}

	params := make([]token.Token, 0)
	if !p.check(token.RIGHT_PAREN) {
		matchedComma := true
		for matchedComma {
			if len(params) >= 255 {
				p.error(p.peek(), "Can't have more than 255 parameters")
			}

			param, paramErr := p.consume(token.IDENTIFIER, "Expect parameter name")
			if paramErr != nil {
				return nil, err
			}

			params = append(params, param)
			matchedComma = p.match(token.COMMA)
		}
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after parameters")
	if err != nil {
		return nil, err
	}

	_, err = p.consume(token.LEFT_BRACE, fmt.Sprintf("Expect '{' before %s body", kind))
	if err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return FunctionStmt{name, params, body}, nil
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
	if p.match(token.BREAK) {
		t := p.previous()
		_, err := p.consume(token.SEMICOLON, "Expected ';' after 'break'")
		if err != nil {
			return nil, err
		}
		return BreakStmt{t}, nil
	}
	if p.match(token.CLASS) {
		return p.classDeclaration()
	}
	if p.match(token.CONTINUE) {
		t := p.previous()
		_, err := p.consume(token.SEMICOLON, "Expected ';' after 'break'")
		if err != nil {
			return nil, err
		}
		return ContinueStmt{t}, nil
	}
	if p.match(token.FOR) {
		return p.forStatement()
	}
	if p.match(token.IF) {
		return p.ifStatement()
	}
	if p.match(token.PRINT) {
		return p.printStatement()
	}
	if p.match(token.RETURN) {
		return p.returnStatement()
	}
	if p.match(token.WHILE) {
		return p.whileStatement()
	}
	if p.match(token.LEFT_BRACE) {
		block, err := p.block()
		if err != nil {
			return nil, err
		}
		return BlockStmt{block}, nil
	}

	return p.expressionStatement()
}

func (p *Parser) classDeclaration() (Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect class name")
	if err != nil {
		return nil, err
	}

	var superclass *VariableExpr = nil
	if p.match(token.LESS) {
		token, superclassErr := p.consume(token.IDENTIFIER, "Expect superclass name")
		if superclassErr != nil {
			return nil, superclassErr
		}
		superclass = &VariableExpr{token}
	}

	_, err = p.consume(token.LEFT_BRACE, "Expect '{' before class body")
	if err != nil {
		return nil, err
	}

	methods := make([]FunctionStmt, 0)
	for !p.check(token.RIGHT_BRACE) && !p.atEnd() {
		fn, methodErr := p.function("method")
		if methodErr != nil {
			return nil, methodErr
		}
		methods = append(methods, fn.(FunctionStmt))
	}

	_, err = p.consume(token.RIGHT_BRACE, "Expect '}' before class body")
	if err != nil {
		return nil, err
	}

	return ClassStmt{name, superclass, methods}, nil
}

func (p *Parser) forStatement() (Stmt, error) {
	_, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'for'")
	if err != nil {
		return nil, err
	}

	// initializer
	var initializer Stmt
	if p.match(token.SEMICOLON) {
		initializer = nil
	} else if p.match(token.VAR) {
		initializer, err = p.varDeclaration()
	} else {
		initializer, err = p.expressionStatement()
	}
	if err != nil {
		return nil, err
	}

	var condition Expr
	if !p.check(token.SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after loop condition")
	if err != nil {
		return nil, err
	}

	var increment Expr
	if !p.check(token.RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after for clauses")
	if err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	// desugar the loop construction into ({ initializer; while (condition) { body; increment; } })

	if increment != nil {
		body = BlockStmt{
			statements: []Stmt{body, ExpressionStmt{increment}},
		}
	}

	if condition != nil {
		body = WhileStmt{condition, body}
	}

	if initializer != nil {
		body = BlockStmt{
			statements: []Stmt{initializer, body},
		}
	}

	return body, nil
}

func (p *Parser) ifStatement() (Stmt, error) {
	p.consume(token.LEFT_PAREN, "Expect '(' after 'if'")
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after if condition")
	if err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch Stmt
	if p.match(token.ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return IfStmt{condition, thenBranch, elseBranch}, nil

}

func (p *Parser) printStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return expr, err
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after value")
	if err != nil {
		return nil, err
	}
	return PrintStmt{expr}, nil
}

func (p *Parser) returnStatement() (Stmt, error) {
	keyword := p.previous()
	var returnVal Expr = nil

	if !p.check(token.SEMICOLON) {
		var err error
		returnVal, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err := p.consume(token.SEMICOLON, "Expect ':' after return value")
	if err != nil {
		return nil, err
	}

	return ReturnStmt{keyword, returnVal}, nil
}

func (p *Parser) whileStatement() (Stmt, error) {
	_, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'while'")
	if err != nil {
		return nil, err
	}
	cond, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after while condition")
	if err != nil {
		return nil, err
	}
	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return WhileStmt{cond, body}, nil
}

func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return expr, err
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after value")
	if err != nil {
		return nil, err
	}
	return ExpressionStmt{expr}, nil
}

// block() assumes that the preceding left brace has already been consumed, and
// will not check for its presence.
func (p *Parser) block() ([]Stmt, error) {
	statements := make([]Stmt, 0)
	for !p.check(token.RIGHT_BRACE) && !p.atEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	_, err := p.consume(token.RIGHT_BRACE, "Expect '}' after block.")
	if err != nil {
		return nil, err
	}

	return statements, nil
}

// Attempts to match any of the given tokens, and consumes the token if found.
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
	return p.assignment()
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
		} else if getter, ok := expr.(GetExpr); ok {
			return SetExpr{getter.object, getter.name, value}, nil
		}

		return nil, p.error(tok, "Invalid assignment target")
	}

	return expr, nil
}

func (p *Parser) condition() (Expr, error) {
	expr, err := p.or()
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

func (p *Parser) or() (Expr, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	for p.match(token.OR) {
		operator := p.previous()
		right, err := p.and()
		if err != nil {
			return nil, err
		}
		expr = LogicalExpr{expr, operator, right}
	}

	return expr, nil
}

func (p *Parser) and() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(token.AND) {
		operator := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = LogicalExpr{expr, operator, right}
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

	return p.call()
}

func (p *Parser) call() (Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	// keep matching calls as long as they exist, using the previous expression
	// as the new callee
	for {
		if p.match(token.LEFT_PAREN) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(token.DOT) {
			name, err := p.consume(token.IDENTIFIER, "Expcet property name after '.'")
			if err != nil {
				return nil, err
			}
			expr = GetExpr{expr, name}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) finishCall(callee Expr) (Expr, error) {
	args := make([]Expr, 0)
	if !p.check(token.RIGHT_PAREN) {
		matchedComma := true
		for matchedComma {
			if len(args) >= 255 {
				// We deliberately don't raise the error here, as we don't need to
				// recover from an unknown state. We only want to report that it
				// happened
				p.error(p.peek(), "Maximum arguments reached (255)")
			}
			expr, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, expr)
			matchedComma = p.match(token.COMMA)
		}
	}

	token, err := p.consume(token.RIGHT_PAREN, "Expect ')' after arguments")
	if err != nil {
		return nil, err
	}

	return CallExpr{callee, token, args}, nil
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

	if p.match(token.THIS) {
		return ThisExpr{p.previous()}, nil
	}
	if p.match(token.SUPER) {
		keyword := p.previous()
		_, err := p.consume(token.DOT, "Expect '.' after 'super'")
		if err != nil {
			return nil, err
		}

		method, methodErr := p.consume(token.IDENTIFIER, "Expect superclass method name")
		if methodErr != nil {
			return nil, methodErr
		}
		return SuperExpr{keyword, method}, nil
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
