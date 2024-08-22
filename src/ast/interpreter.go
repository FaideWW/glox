package ast

type Interpreter struct {
	environment *Environment
}

func NewInterpreter() Interpreter {
	env := NewGlobalEnvironment()
	return Interpreter{
		environment: &env,
	}
}

func (i *Interpreter) Interpret(statements []Stmt) error {
	for _, statement := range statements {
		err := statement.(EvaluableStmt).Evaluate(i.environment)
		if err != nil {
			return err
		}
	}

	return nil
}
