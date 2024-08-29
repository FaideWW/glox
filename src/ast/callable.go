package ast

import (
	"fmt"

	"github.com/faideww/glox/src/errors"
	"github.com/faideww/glox/src/token"
)

type Callable interface {
	Arity() int
	Call(args []LoxValue, i *Interpreter) (LoxValue, error)
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
	declaration   FunctionStmt
	closure       *Environment
	isInitializer bool
}

func NewLoxFunction(declaration FunctionStmt, closure *Environment, isInitializer bool) LoxFunction {
	return LoxFunction{declaration, closure, isInitializer}
}

func (f LoxFunction) Arity() int {
	return len(f.declaration.params)
}

func (f LoxFunction) bind(ctx *LoxInstance) LoxFunction {
	env := NewEnvironment(f.closure)
	env.Define("this", ctx)
	return NewLoxFunction(f.declaration, env, f.isInitializer)
}

func (f LoxFunction) Call(args []LoxValue, i *Interpreter) (LoxValue, error) {
	funcEnv := NewEnvironment(f.closure)

	for i := range f.declaration.params {
		funcEnv.Define(f.declaration.params[i].Lexeme, args[i])
	}

	var err error = nil
	prevEnv := i.currentEnv
	i.currentEnv = funcEnv
	defer func() { i.currentEnv = prevEnv }()
	for _, stmt := range f.declaration.body {
		err = stmt.(EvaluableStmt).Evaluate(i)
		if err != nil {
			switch err := err.(type) {
			case *ReturnException:
				if f.isInitializer {
					return f.closure.GetAt(0, "this"), nil
				}
				return err.value, nil
			default:
				break
			}
		}
	}
	if f.isInitializer {
		return f.closure.GetAt(0, "this"), nil
	}
	return nil, err
}
func (f LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", f.declaration.name.Lexeme)
}

type LoxClass struct {
	name       string
	superclass *LoxClass
	methods    map[string]LoxFunction
}

func NewLoxClass(name string, superclass *LoxClass, methods map[string]LoxFunction) *LoxClass {
	return &LoxClass{name, superclass, methods}
}

func (c *LoxClass) String() string {
	return c.name
}

func (c *LoxClass) Call(args []LoxValue, i *Interpreter) (LoxValue, error) {
	instance := NewLoxInstance(c)
	initializer := c.findMethod("init")
	if initializer != nil {
		initializer.bind(instance).Call(args, i)
	}

	return instance, nil
}

func (c *LoxClass) Arity() int {
	initializer := c.findMethod("init")
	if initializer == nil {
		return 0
	}
	return initializer.Arity()
}

func (c *LoxClass) findMethod(name string) *LoxFunction {
	if method, ok := c.methods[name]; ok {
		return &method
	}

	if c.superclass != nil {
		return c.superclass.findMethod(name)
	}

	return nil
}

type LoxInstance struct {
	cls    *LoxClass
	fields map[string]LoxValue
}

func NewLoxInstance(cls *LoxClass) *LoxInstance {
	return &LoxInstance{cls: cls, fields: make(map[string]LoxValue)}
}

func (i *LoxInstance) String() string {
	return fmt.Sprintf("%s instance", i.cls.String())
}

func (i *LoxInstance) Get(name token.Token) (LoxValue, error) {
	if value, ok := i.fields[name.Lexeme]; ok {
		return value, nil
	}

	method := i.cls.findMethod(name.Lexeme)
	if method != nil {
		return method.bind(i), nil
	}

	return nil, errors.NewRuntimeError(name, fmt.Sprintf("Undefined property '%s'", name.Lexeme))
}

func (i *LoxInstance) Set(name token.Token, value LoxValue) {
	i.fields[name.Lexeme] = value
}
