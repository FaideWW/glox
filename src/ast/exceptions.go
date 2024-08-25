package ast

import (
	"github.com/faideww/glox/src/errors"
	"github.com/faideww/glox/src/token"
)

type BreakException struct {
	token token.Token
}

func (e *BreakException) Error() string {
	return errors.NewRuntimeError(e.token, "'break' called outside of loop").Error()
}

func NewBreakException(t token.Token) *BreakException {
	return &BreakException{t}
}

type ContinueException struct {
	token token.Token
}

func (e *ContinueException) Error() string {
	return errors.NewRuntimeError(e.token, "'continue' called outside of loop").Error()
}

func NewContinueException(t token.Token) *ContinueException {
	return &ContinueException{t}
}

type ReturnException struct {
	token token.Token
	value LoxValue
}

func (e *ReturnException) Error() string {
	return errors.NewRuntimeError(e.token, "'return' called outside of function").Error()
}

func NewReturnException(token token.Token, value LoxValue) *ReturnException {
	return &ReturnException{token, value}
}
