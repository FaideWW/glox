package ast

import (
	"fmt"

	"github.com/faideww/glox/src/errors"
	"github.com/faideww/glox/src/token"
)

type Environment struct {
	parent    *Environment
	variables map[string]LoxValue
}

func NewGlobalEnvironment() Environment {
	return Environment{
		parent:    nil,
		variables: make(map[string]LoxValue),
	}
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:    parent,
		variables: make(map[string]LoxValue),
	}
}

func (e *Environment) Define(name string, value LoxValue) {
	e.variables[name] = value
}

func (e *Environment) Get(name token.Token) (LoxValue, error) {
	value, ok := e.variables[name.Lexeme]
	if ok {
		return value, nil
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	fmt.Printf("%#v\n", e)

	return nil, errors.NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'", name.Lexeme))
}

func (e *Environment) ancestor(distance int) *Environment {
	currentEnv := e
	for i := 0; i < distance; i++ {
		currentEnv = currentEnv.parent
	}
	return currentEnv
}

func (e *Environment) GetAt(distance int, name string) LoxValue {
	value := e.ancestor(distance).variables[name]
	return value
}

func (e *Environment) Assign(name token.Token, nextValue LoxValue) error {
	if _, ok := e.variables[name.Lexeme]; ok {
		e.variables[name.Lexeme] = nextValue
		return nil
	}

	if e.parent != nil {
		return e.parent.Assign(name, nextValue)
	}

	return errors.NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'", name.Lexeme))
}

func (e *Environment) AssignAt(distance int, name token.Token, nextValue LoxValue) {
	e.ancestor(distance).Assign(name, nextValue)
}
