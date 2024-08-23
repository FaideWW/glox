package errors

import (
	"fmt"
	"io"

	"github.com/faideww/glox/src/token"
)

type ErrorReporter struct {
	errors []error
}

func NewErrorReporter() *ErrorReporter {
	return &ErrorReporter{
		errors: make([]error, 0),
	}
}

func (r *ErrorReporter) Collect(err error) {
	r.errors = append(r.errors, err)
}

func (r *ErrorReporter) Clear() {
	clear(r.errors[:])
}

func (r *ErrorReporter) Report(w io.Writer) {
	for _, err := range r.errors {
		fmt.Fprintln(w, err)
	}
}

func (r *ErrorReporter) Last() error {
	if len(r.errors) < 1 {
		return nil
	}
	return r.errors[len(r.errors)-1]
}

type ParserError struct {
	token   token.Token
	message string
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("[line %d] Error at '%s': %s", e.token.Line, e.token.Lexeme, e.message)
}

func NewParserError(token token.Token, message string) *ParserError {
	err := &ParserError{token, message}
	return err
}

type RuntimeError struct {
	token   token.Token
	message string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("%s\n[line %d]\n", e.message, e.token.Line)
}

func NewRuntimeError(token token.Token, message string) *RuntimeError {
	err := &RuntimeError{token, message}
	return err
}
