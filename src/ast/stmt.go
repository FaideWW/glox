package ast

import "github.com/faideww/glox/src/token"

type Stmt interface {
}

type BlockStmt struct {
	statements []Stmt
}

type BreakStmt struct{ token token.Token }

type ClassStmt struct {
	name       token.Token
	superclass *VariableExpr
	methods    []FunctionStmt
}

type ContinueStmt struct{ token token.Token }

type ExpressionStmt struct {
	expression Expr
}

type FunctionStmt struct {
	name   token.Token
	params []token.Token
	body   []Stmt
}

type IfStmt struct {
	condition  Expr
	thenBranch Stmt
	elseBranch Stmt
}

type PrintStmt struct {
	expression Expr
}

type ReturnStmt struct {
	keyword token.Token
	value   Expr
}

type VarStmt struct {
	name        token.Token
	initializer Expr
}

type WhileStmt struct {
	condition Expr
	body      Stmt
}
