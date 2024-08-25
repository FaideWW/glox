package ast

import (
	"time"

	"github.com/faideww/glox/src/token"
)

type Interpreter struct {
	globals    *Environment
	currentEnv *Environment
	locals     map[Expr]int
}

func NewInterpreter() *Interpreter {
	globalEnv := NewGlobalEnvironment()

	globalEnv.Define("clock", NewNativeFunction(
		func() int { return 0 },
		func(args []LoxValue, _ *Interpreter) (LoxValue, error) {
			return float64(time.Now().Unix()), nil
		},
	))

	return &Interpreter{
		globals:    &globalEnv,
		currentEnv: &globalEnv,
		locals:     make(map[Expr]int),
	}
}

func (i *Interpreter) Interpret(statements []Stmt) error {
	for _, statement := range statements {
		err := statement.(EvaluableStmt).Evaluate(i)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) InterpretExpression(expression Expr) (LoxValue, error) {
	return expression.(Evaluable).Evaluate(i)

}
func (i *Interpreter) resolve(expr Expr, depth int) {
	i.locals[expr] = depth
}

func (i *Interpreter) lookupVariable(name token.Token, expr Expr) (LoxValue, error) {
	if hops, ok := i.locals[expr]; ok {
		// If the resolver has been run, this is guaranteed to find a value
		return i.currentEnv.GetAt(hops, name.Lexeme), nil
	} else {
		return i.globals.Get(name)
	}
}
