package ast

import "fmt"

type Callable interface {
	Arity() int
	Call(args []LoxValue, i *Interpreter) (LoxValue, error)
	String() string
}

type arityFn func() int
type callFn func(args []LoxValue, i *Interpreter) (LoxValue, error)

type NativeFunction struct {
	arity arityFn
	call  callFn
}

func NewNativeFunction(arity arityFn, call callFn) *NativeFunction {
	return &NativeFunction{
		arity,
		call}
}

func (f *NativeFunction) Arity() int {
	return f.arity()
}

func (f *NativeFunction) Call(args []LoxValue, i *Interpreter) (LoxValue, error) {
	return f.call(args, i)
}

func (f *NativeFunction) String() string {
	return "<native function>"
}

type LoxFunction struct {
	declaration FunctionStmt
	closure     *Environment
}

func NewLoxFunction(declaration FunctionStmt, closure *Environment) LoxFunction {
	return LoxFunction{declaration, closure}
}

func (f LoxFunction) Arity() int {
	return len(f.declaration.params)
}
func (f LoxFunction) Call(args []LoxValue, i *Interpreter) (LoxValue, error) {
	funcEnv := NewEnvironment(f.closure)

	for i := range f.declaration.params {
		funcEnv.Define(f.declaration.params[i].Lexeme, args[i])
	}

	var err error = nil
	prevEnv := i.currentEnv
	i.currentEnv = funcEnv
	for _, stmt := range f.declaration.body {
		err = stmt.(EvaluableStmt).Evaluate(i)
		if err != nil {
			switch err := err.(type) {
			case *ReturnException:
				i.currentEnv = prevEnv
				return err.value, nil
			default:
				break
			}
		}
	}
	i.currentEnv = prevEnv
	return nil, err
}
func (f LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", f.declaration.name.Lexeme)
}
