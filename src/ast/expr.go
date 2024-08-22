package ast

import (
	"github.com/faideww/glox/src/token"
)

type LoxValue interface{}

type Expr interface {
}

type Literal struct {
	value LoxValue
}

type Unary struct {
	operator token.Token
	right    Expr
}

type Variable struct {
	name token.Token
}

type Assignment struct {
	name  token.Token
	value Expr
}

type Binary struct {
	left     Expr
	operator token.Token
	right    Expr
}

type Grouping struct {
	expression Expr
}

type Ternary struct {
	condition Expr
	left      Expr
	right     Expr
}
