package ast

import "github.com/faideww/glox/src/token"

type Stmt interface {
}

type ExpressionStmt struct {
	expression Expr
}

type PrintStmt struct {
	expression Expr
}

type VarStmt struct {
	name        token.Token
	initializer Expr
}

type Block struct {
	statements []Stmt
}
