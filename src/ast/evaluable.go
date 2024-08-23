package ast

import (
	"fmt"
	"math"
	"strconv"

	"github.com/faideww/glox/src/errors"
	"github.com/faideww/glox/src/token"
)

type EvaluableStmt interface {
	Evaluate(env *Environment) error
}

func (ps PrintStmt) Evaluate(env *Environment) error {
	result, err := ps.expression.(Evaluable).Evaluate(env)
	if err != nil {
		return err
	}

	fmt.Println(ToString(result))
	return nil
}

func (es ExpressionStmt) Evaluate(env *Environment) error {
	_, err := es.expression.(Evaluable).Evaluate(env)
	return err
}

func (vs VarStmt) Evaluate(env *Environment) error {
	var value LoxValue
	var err error

	if vs.initializer == nil {
		env.Declare(vs.name.Lexeme)
	} else {
		value, err = vs.initializer.(Evaluable).Evaluate(env)
		if err != nil {
			return err
		}
		env.Initialize(vs.name.Lexeme, value)
	}

	return nil
}

func (b Block) Evaluate(env *Environment) error {
	blockEnv := NewEnvironment(env)

	for _, statement := range b.statements {
		err := statement.(EvaluableStmt).Evaluate(&blockEnv)
		if err != nil {
			return err
		}
	}

	return nil
}

type Evaluable interface {
	Evaluate(env *Environment) (LoxValue, error)
}

func (l LiteralExpr) Evaluate(env *Environment) (LoxValue, error) {
	return l.value, nil
}

func (g GroupingExpr) Evaluate(env *Environment) (LoxValue, error) {
	return g.expression.(Evaluable).Evaluate(env)
}

func (v VariableExpr) Evaluate(env *Environment) (LoxValue, error) {
	value, err := env.Get(v.name)
	return value, err
}

func (u UnaryExpr) Evaluate(env *Environment) (LoxValue, error) {
	right, err := u.right.(Evaluable).Evaluate(env)
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

func (a AssignmentExpr) Evaluate(env *Environment) (LoxValue, error) {
	value, err := a.value.(Evaluable).Evaluate(env)
	if err != nil {
		return nil, err
	}

	err = env.Assign(a.name, value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (b BinaryExpr) Evaluate(env *Environment) (LoxValue, error) {
	left, leftErr := b.left.(Evaluable).Evaluate(env)
	right, rightErr := b.right.(Evaluable).Evaluate(env)

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

func (t TernaryExpr) Evaluate(env *Environment) (LoxValue, error) {
	cond, condErr := t.condition.(Evaluable).Evaluate(env)

	if condErr != nil {
		return cond, condErr
	}

	left, leftErr := t.left.(Evaluable).Evaluate(env)
	if leftErr != nil {
		return left, leftErr
	}

	right, rightErr := t.right.(Evaluable).Evaluate(env)

	if rightErr != nil {
		return right, rightErr
	}

	if isTruthy(cond) {
		return left, nil
	} else {
		return right, nil
	}
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

	return fmt.Sprintf("%v", value)
}
