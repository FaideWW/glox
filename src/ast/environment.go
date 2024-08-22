package ast

import (
	"fmt"

	"github.com/faideww/glox/src/token"
)

type Environment struct {
	parent *Environment
	values map[string]LoxValue
}

func NewGlobalEnvironment() Environment {
	return Environment{
		parent: nil,
		values: make(map[string]LoxValue),
	}
}

func NewEnvironment(parent *Environment) Environment {
	return Environment{
		parent: parent,
		values: make(map[string]LoxValue),
	}
}

func (e *Environment) Define(name string, value LoxValue) {
	e.values[name] = value
}

func (e *Environment) Get(name token.Token) (LoxValue, error) {
	value, ok := e.values[name.Lexeme]
	if ok {
		return value, nil
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	return nil, NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'", name.Lexeme))
}

func (e *Environment) Assign(name token.Token, nextValue LoxValue) error {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = nextValue
		return nil
	}

	if e.parent != nil {
		return e.parent.Assign(name, nextValue)
	}

	return NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'", name.Lexeme))
}
