package ast

import (
	"fmt"

	"github.com/faideww/glox/src/errors"
	"github.com/faideww/glox/src/token"
)

type LoxVariable struct {
	initialized bool
	value       LoxValue
}

type Environment struct {
	parent    *Environment
	variables map[string]LoxVariable
}

func NewGlobalEnvironment() Environment {
	return Environment{
		parent:    nil,
		variables: make(map[string]LoxVariable),
	}
}

func NewEnvironment(parent *Environment) Environment {
	return Environment{
		parent:    parent,
		variables: make(map[string]LoxVariable),
	}
}

func (e *Environment) Declare(name string) {
	e.variables[name] = LoxVariable{initialized: false, value: nil}
}

func (e *Environment) Initialize(name string, value LoxValue) {
	e.variables[name] = LoxVariable{initialized: true, value: value}
}

func (e *Environment) Get(name token.Token) (LoxValue, error) {
	value, ok := e.variables[name.Lexeme]
	if ok {
		if !value.initialized {
			return nil, errors.NewRuntimeError(name, fmt.Sprintf("Variable '%s' is uninitialized", name.Lexeme))
		}
		return value.value, nil
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	return nil, errors.NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'", name.Lexeme))
}

func (e *Environment) Assign(name token.Token, nextValue LoxValue) error {
	if v, ok := e.variables[name.Lexeme]; ok {
		v.value = nextValue
		v.initialized = true
		e.variables[name.Lexeme] = v
		return nil
	}

	if e.parent != nil {
		return e.parent.Assign(name, nextValue)
	}

	return errors.NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'", name.Lexeme))
}
