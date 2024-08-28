package ast

import (
	"github.com/faideww/glox/src/token"
)

type LoxValue interface{}

type Expr interface {
}

type AssignmentExpr struct {
	name  token.Token
	value Expr
}

type BinaryExpr struct {
	left     Expr
	operator token.Token
	right    Expr
}

type CallExpr struct {
	callee    Expr
	paren     token.Token
	arguments []Expr
}

type GetExpr struct {
	object Expr
	name   token.Token
}

type GroupingExpr struct {
	expression Expr
}

type LiteralExpr struct {
	value LoxValue
}

type LogicalExpr struct {
	left     Expr
	operator token.Token
	right    Expr
}

type SetExpr struct {
	obj   Expr
	name  token.Token
	value Expr
}

type TernaryExpr struct {
	condition Expr
	left      Expr
	right     Expr
}

type ThisExpr struct {
	keyword token.Token
}

type UnaryExpr struct {
	operator token.Token
	right    Expr
}

type VariableExpr struct {
	name token.Token
}
