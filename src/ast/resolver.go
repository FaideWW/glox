package ast

import (
	"fmt"

	"github.com/faideww/glox/src/errors"
	"github.com/faideww/glox/src/token"
)

type ScopeVariable struct {
	declaration *token.Token
	name        string
	defined     bool
	used        bool
}

type Scope map[string]ScopeVariable

type FunctionType int

const (
	FNTYPE_NONE = iota
	FNTYPE_FUNCTION
	FNTYPE_INITIALIZER
	FNTYPE_METHOD
)

type ClassType int

const (
	CLASSTYPE_NONE = iota
	CLASSTYPE_SUBCLASS
	CLASSTYPE_CLASS
)

type Resolver struct {
	interpreter     *Interpreter
	scopes          []Scope
	currentFunction FunctionType
	inLoop          bool
	currentClass    ClassType
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		interpreter:     interpreter,
		scopes:          make([]Scope, 0),
		currentFunction: FNTYPE_NONE,
		inLoop:          false,
		currentClass:    CLASSTYPE_NONE,
	}
}

func (r *Resolver) Resolve(statements []Stmt) error {
	for _, statement := range statements {
		err := statement.(Resolvable).Resolve(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(Scope))
}

func (r *Resolver) endScope() error {

	// scan for unused variables
	scope := r.scopes[len(r.scopes)-1]
	for _, v := range scope {
		if !v.used {
			return errors.NewAnalysisError(*v.declaration, fmt.Sprintf("Unused variable '%s'", v.name))
		}
	}

	// TODO: this will panic if we try to end a scope when we have none to end. Do we want that,
	// or should we silently ignore/log an error and recover?
	r.scopes = r.scopes[:len(r.scopes)-1]
	return nil
}

func (r *Resolver) declare(name token.Token) error {
	if len(r.scopes) == 0 {
		return nil
	}

	currentScope := r.scopes[len(r.scopes)-1]

	if _, ok := currentScope[name.Lexeme]; ok {
		errStr := fmt.Sprintf("Variable '%s' already exists in this scope", name.Lexeme)
		return errors.NewAnalysisError(name, errStr)
	}

	currentScope[name.Lexeme] = ScopeVariable{
		declaration: &name,
		name:        name.Lexeme,
		defined:     false,
		used:        false,
	}
	return nil
}

func (r *Resolver) define(name token.Token) {
	if len(r.scopes) == 0 {
		return
	}

	currentScope := r.scopes[len(r.scopes)-1]
	v := currentScope[name.Lexeme]
	v.defined = true
	currentScope[name.Lexeme] = v
}

// Walk back up the scope stack to find the nearest enclosing scope defining
// the provided variable name, then pass the depth to the interpreter so it can
// resolve its value later during runtime
func (r *Resolver) resolveLocal(expr Expr, name token.Token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if v, ok := r.scopes[i][name.Lexeme]; ok {
			r.interpreter.resolve(expr, len(r.scopes)-1-i)

			v.used = true
			r.scopes[i][name.Lexeme] = v

			return
		}
	}
}
