package token

import "fmt"

type TokenType int

const (
	// Single-character tokens
	LEFT_PAREN = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	LEFT_BLOCK_COMMENT
	RIGHT_BLOCK_COMMENT
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR
	QMARK // ternary "?"
	COLON // ternary ":"

	// 1-2 char tokens
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// literals
	IDENTIFIER
	STRING
	NUMBER

	// keywords
	AND
	BREAK
	CLASS
	CONTINUE
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE

	EOF
)

type LiteralObject interface{}

type Token struct {
	TokenType TokenType
	Lexeme    string
	Literal   LiteralObject
	Line      int
	tokenId   int
}

func NewToken(tokenType TokenType, lexeme string, literal LiteralObject, line int, tokenId int) Token {
	return Token{tokenType, lexeme, literal, line, tokenId}
}

func (t Token) String() string {
	return fmt.Sprintf("%d %s %x", t.TokenType, t.Lexeme, t.Literal)
}
