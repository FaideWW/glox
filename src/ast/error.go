package ast

import (
	"fmt"

	"github.com/faideww/glox/src/token"
)

type ParserError struct {
	token   token.Token
	message string
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("[line %d] Error at '%s': %s", e.token.Line, e.token.Lexeme, e.message)
}

func NewParserError(token token.Token, message string) *ParserError {
	return &ParserError{token, message}
}

type RuntimeError struct {
	token   token.Token
	message string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("%s\n[line %d]\n", e.message, e.token.Line)
}

func NewRuntimeError(token token.Token, message string) *RuntimeError {
	return &RuntimeError{token, message}
}
