package errors

import (
	"fmt"
	"io"
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
