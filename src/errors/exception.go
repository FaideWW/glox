package errors

import "github.com/faideww/glox/src/token"

type BreakException struct {
	token token.Token
}

func (e *BreakException) Error() string {
	return NewRuntimeError(e.token, "'break' called outside of loop").Error()
}

func NewBreakException(t token.Token) *BreakException {
	return &BreakException{t}
}

type ContinueException struct {
	token token.Token
}

func (e *ContinueException) Error() string {
	return NewRuntimeError(e.token, "'continue' called outside of loop").Error()
}

func NewContinueException(t token.Token) *ContinueException {
	return &ContinueException{t}
}
