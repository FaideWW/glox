package ast

import (
	"fmt"
	"math"
	"strconv"

	"github.com/faideww/glox/src/errors"
	"github.com/faideww/glox/src/token"
)

type EvaluableStmt interface {
	Evaluate(i *Interpreter) error
}

func (bs BreakStmt) Evaluate(i *Interpreter) error {
	return NewBreakException(bs.token)
}

func (cs ClassStmt) Evaluate(i *Interpreter) error {
	var superclassValue LoxValue = nil
	var superclass *LoxClass = nil
	if cs.superclass != nil {
		var superclassErr error
		superclassValue, superclassErr = cs.superclass.Evaluate(i)
		if superclassErr != nil {
			return superclassErr
		}

		var ok bool
		if superclass, ok = superclassValue.(*LoxClass); !ok {
			return errors.NewRuntimeError(cs.superclass.name, "Superclass must be a class")
		}
	}

	i.currentEnv.Define(cs.name.Lexeme, nil)

	if cs.superclass != nil {
		i.currentEnv = NewEnvironment(i.currentEnv)
		i.currentEnv.Define("super", superclass)
	}

	methods := make(map[string]LoxFunction)
	for _, method := range cs.methods {
		fn := NewLoxFunction(method, i.currentEnv, method.name.Lexeme == "init")
		methods[method.name.Lexeme] = fn
	}

	cls := NewLoxClass(cs.name.Lexeme, superclass, methods)

	if cs.superclass != nil {
		i.currentEnv = i.currentEnv.parent
	}

	return i.currentEnv.Assign(cs.name, cls)
}

func (cs ContinueStmt) Evaluate(i *Interpreter) error {
	return NewContinueException(cs.token)
}

func (fs FunctionStmt) Evaluate(i *Interpreter) error {
	function := NewLoxFunction(fs, i.currentEnv, false)
	i.currentEnv.Define(fs.name.Lexeme, function)
	return nil
}

func (is IfStmt) Evaluate(i *Interpreter) error {
	cond, err := is.condition.(Evaluable).Evaluate(i)
	if err != nil {
		return err
	}
	if isTruthy(cond) {
		err := is.thenBranch.(EvaluableStmt).Evaluate(i)
		if err != nil {
			return err
		}
	} else if is.elseBranch != nil {
		err := is.elseBranch.(EvaluableStmt).Evaluate(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps PrintStmt) Evaluate(i *Interpreter) error {
	result, err := ps.expression.(Evaluable).Evaluate(i)
	if err != nil {
		return err
	}

	fmt.Println(ToString(result))
	return nil
}

func (rs ReturnStmt) Evaluate(i *Interpreter) error {
	var retVal LoxValue
	if rs.value != nil {
		var err error
		retVal, err = rs.value.(Evaluable).Evaluate(i)
		if err != nil {
			return err
		}
	}

	return NewReturnException(rs.keyword, retVal)
}

func (ws WhileStmt) Evaluate(i *Interpreter) error {
	cond, condErr := ws.condition.(Evaluable).Evaluate(i)
	for condErr == nil && isTruthy(cond) {
		bodyErr := ws.body.(EvaluableStmt).Evaluate(i)

		shouldBreak := false
		if bodyErr != nil {
			switch bodyErr.(type) {
			case *ContinueException:
				continue
			case *BreakException:
				shouldBreak = true
			default:
				return bodyErr
			}

		}

		if shouldBreak {
			break
		}

		cond, condErr = ws.condition.(Evaluable).Evaluate(i)
	}
	if condErr != nil {
		return condErr
	}

	return nil
}

func (es ExpressionStmt) Evaluate(i *Interpreter) error {
	_, err := es.expression.(Evaluable).Evaluate(i)
	return err
}

func (vs VarStmt) Evaluate(i *Interpreter) error {
	var value LoxValue
	var err error

	if vs.initializer == nil {
		i.currentEnv.Define(vs.name.Lexeme, nil)
	} else {
		value, err = vs.initializer.(Evaluable).Evaluate(i)
		if err != nil {
			return err
		}
		i.currentEnv.Define(vs.name.Lexeme, value)
	}

	return nil
}

func (b BlockStmt) Evaluate(i *Interpreter) error {
	prevEnv := i.currentEnv
	blockEnv := NewEnvironment(i.currentEnv)
	i.currentEnv = blockEnv

	var err error
	for _, statement := range b.statements {
		err = statement.(EvaluableStmt).Evaluate(i)
		if err != nil {
			break
		}
	}

	i.currentEnv = prevEnv

	return err
}

type Evaluable interface {
	Evaluate(i *Interpreter) (LoxValue, error)
}

func (a AssignmentExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	value, err := a.value.(Evaluable).Evaluate(i)
	if err != nil {
		return nil, err
	}

	if hops, ok := i.locals[a]; ok {
		i.currentEnv.AssignAt(hops, a.name, value)
	} else {
		err = i.globals.Assign(a.name, value)
	}

	if err != nil {
		return nil, err
	}
	return value, nil
}

func (b BinaryExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	left, leftErr := b.left.(Evaluable).Evaluate(i)
	right, rightErr := b.right.(Evaluable).Evaluate(i)

	if leftErr != nil {
		return left, leftErr
	}

	if rightErr != nil {
		return right, rightErr
	}

	lFloat, lOk := left.(float64)
	rFloat, rOk := right.(float64)

	switch b.operator.TokenType {
	case token.GREATER:
		if lOk && rOk {
			return lFloat > rFloat, nil
		}
	case token.GREATER_EQUAL:
		if lOk && rOk {
			return lFloat >= rFloat, nil
		}
	case token.LESS:
		if lOk && rOk {
			return lFloat < rFloat, nil
		}
	case token.LESS_EQUAL:
		if lOk && rOk {
			return lFloat <= rFloat, nil
		}
	case token.BANG_EQUAL:
		return left != right, nil
	case token.EQUAL_EQUAL:
		return left == right, nil
	case token.MINUS:
		if lOk && rOk {
			return lFloat - rFloat, nil
		}
	case token.SLASH:
		if lOk && rOk {
			if result, ok := safeDivide(lFloat, rFloat); ok {
				return result, nil
			}
			return nil, errors.NewRuntimeError(b.operator, "Divide by zero")
		}
	case token.STAR:
		if lOk && rOk {
			return lFloat * rFloat, nil
		}
	case token.PLUS:
		if lOk && rOk {
			return lFloat + rFloat, nil
		}

		_, lOk = left.(string)
		_, rOk = right.(string)

		if lOk || rOk {
			return fmt.Sprintf("%s%s", ToString(left), ToString(right)), nil
		}

		return nil, errors.NewRuntimeError(b.operator, "Operands must be two numbers or two strings")
	}

	// Unreachable
	return nil, errors.NewRuntimeError(b.operator, "Operands must be numbers")
}

func (c CallExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	callee, err := c.callee.(Evaluable).Evaluate(i)
	if err != nil {
		return nil, err
	}

	argValues := make([]LoxValue, len(c.arguments))
	for j, argExpr := range c.arguments {
		v, err := argExpr.(Evaluable).Evaluate(i)
		if err != nil {
			return nil, err
		}
		argValues[j] = v
	}

	fn, ok := callee.(Callable)
	if !ok {
		return nil, errors.NewRuntimeError(c.paren, "Can only call functions and classes")
	}

	if fn.Arity() != len(argValues) {
		return nil, errors.NewRuntimeError(c.paren, fmt.Sprintf("Expected %d arguments but got %d", fn.Arity(), len(argValues)))
	}

	return fn.Call(argValues, i)
}

func (g GetExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	obj, err := g.object.(Evaluable).Evaluate(i)
	if err != nil {
		return nil, err
	}

	if objInstance, ok := obj.(*LoxInstance); ok {
		return objInstance.Get(g.name)
	}

	return nil, errors.NewRuntimeError(g.name, "Only instances can have properties")
}

func (g GroupingExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	return g.expression.(Evaluable).Evaluate(i)
}

func (l LiteralExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	return l.value, nil
}

func (s SetExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	obj, err := s.obj.(Evaluable).Evaluate(i)
	if err != nil {
		return nil, err
	}

	if instanceObj, ok := obj.(*LoxInstance); ok {
		value, err := s.value.(Evaluable).Evaluate(i)
		if err != nil {
			return nil, err
		}
		instanceObj.Set(s.name, value)
		return value, nil
	}
	return nil, errors.NewRuntimeError(s.name, "Only instances have fields")
}

func (s SuperExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	distance := i.locals[s]
	superclass := i.currentEnv.GetAt(distance, "super").(*LoxClass)

	instance := i.currentEnv.GetAt(distance-1, "this").(*LoxInstance)

	method := superclass.findMethod(s.method.Lexeme)
	if method == nil {
		return nil, errors.NewRuntimeError(s.method, fmt.Sprintf("Undefined property '%s'", s.method.Lexeme))
	}
	return method.bind(instance), nil

}

func (t TernaryExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	cond, condErr := t.condition.(Evaluable).Evaluate(i)

	if condErr != nil {
		return cond, condErr
	}

	if isTruthy(cond) {
		left, leftErr := t.left.(Evaluable).Evaluate(i)
		if leftErr != nil {
			return left, leftErr
		}

		return left, nil
	} else {
		right, rightErr := t.right.(Evaluable).Evaluate(i)

		if rightErr != nil {
			return right, rightErr
		}

		return right, nil
	}
}

func (t ThisExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	return i.lookupVariable(t.keyword, t)
}

func (u UnaryExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	right, err := u.right.(Evaluable).Evaluate(i)
	if err != nil {
		return right, err
	}

	switch u.operator.TokenType {
	case token.BANG:
		return !isTruthy(right), nil
	case token.MINUS:
		if rFloat, ok := right.(float64); ok {
			return -(rFloat), nil
		}

		return nil, errors.NewRuntimeError(u.operator, "Operand must be a number")
	}

	// Unreachable
	return nil, nil
}

func (v VariableExpr) Evaluate(i *Interpreter) (LoxValue, error) {
	return i.lookupVariable(v.name, v)
}

func isTruthy(value LoxValue) bool {
	if value == nil {
		return false
	}
	if bValue, ok := value.(bool); ok {
		return bValue
	}

	return true
}

func safeDivide(a, b float64) (float64, bool) {
	if b == 0 {
		return math.NaN(), false
	}

	return a / b, true
}

func ToString(value LoxValue) string {
	if value == nil {
		return "nil"
	}

	if vFloat, ok := value.(float64); ok {
		return strconv.FormatFloat(vFloat, 'f', -1, 64)
	}

	if fn, ok := value.(Named); ok {
		return fn.String()
	}

	return fmt.Sprintf("%v", value)
}
