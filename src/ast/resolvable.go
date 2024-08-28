package ast

import "github.com/faideww/glox/src/errors"

type Resolvable interface {
	Resolve(r *Resolver) error
}

func (bs BlockStmt) Resolve(r *Resolver) error {
	r.beginScope()
	for _, stmt := range bs.statements {
		err := stmt.(Resolvable).Resolve(r)
		if err != nil {
			return err
		}
	}
	return r.endScope()
}

func (bs BreakStmt) Resolve(r *Resolver) error {
	if !r.inLoop {
		return errors.NewAnalysisError(bs.token, "Can't break outside of loop")
	}
	return nil
}

func (cs ClassStmt) Resolve(r *Resolver) error {
	enclosingClass := r.currentClass
	defer func() { r.currentClass = enclosingClass }()
	r.currentClass = CLASSTYPE_CLASS
	err := r.declare(cs.name)
	if err != nil {
		return err
	}
	r.define(cs.name)

	r.beginScope()
	r.scopes[len(r.scopes)-1]["this"] = ScopeVariable{
		declaration: nil,
		name:        "this",
		defined:     true,
		used:        true,
	}

	for _, method := range cs.methods {
		var declaration FunctionType = FNTYPE_METHOD
		if method.name.Lexeme == "init" {
			declaration = FNTYPE_INITIALIZER
		}
		fnErr := resolveFunction(r, method, declaration)
		if fnErr != nil {
			return fnErr
		}
	}
	err = r.endScope()
	if err != nil {
		return err
	}

	return nil
}

func (cs ContinueStmt) Resolve(r *Resolver) error {
	if !r.inLoop {
		return errors.NewAnalysisError(cs.token, "Can't continue outside of loop")
	}
	return nil
}

func (es ExpressionStmt) Resolve(r *Resolver) error {
	return es.expression.(Resolvable).Resolve(r)
}

func (fs FunctionStmt) Resolve(r *Resolver) error {
	err := r.declare(fs.name)
	if err != nil {
		return err
	}
	r.define(fs.name)

	return resolveFunction(r, fs, FNTYPE_FUNCTION)
}

func resolveFunction(r *Resolver, fs FunctionStmt, fnType FunctionType) error {
	enclosingFn := r.currentFunction
	r.currentFunction = fnType
	defer func() { r.currentFunction = enclosingFn }()
	r.beginScope()
	for _, param := range fs.params {
		err := r.declare(param)
		if err != nil {
			return err
		}
		r.define(param)
	}
	for _, stmt := range fs.body {
		err := stmt.(Resolvable).Resolve(r)
		if err != nil {
			return err
		}
	}
	err := r.endScope()
	return err
}

func (is IfStmt) Resolve(r *Resolver) error {
	err := is.condition.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	err = is.thenBranch.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	if is.elseBranch != nil {
		err = is.elseBranch.(Resolvable).Resolve(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ps PrintStmt) Resolve(r *Resolver) error {
	return ps.expression.(Resolvable).Resolve(r)
}

func (rs ReturnStmt) Resolve(r *Resolver) error {
	if r.currentFunction == FNTYPE_NONE {
		return errors.NewAnalysisError(rs.keyword, "Can't return from top-level code")
	}
	if rs.value != nil {
		if r.currentFunction == FNTYPE_INITIALIZER {
			return errors.NewAnalysisError(rs.keyword, "Can't return a value from a class initializer")
		}
		return rs.value.(Resolvable).Resolve(r)
	}
	return nil
}

func (vs VarStmt) Resolve(r *Resolver) error {
	err := r.declare(vs.name)
	if err != nil {
		return err
	}
	if vs.initializer != nil {
		err := vs.initializer.(Resolvable).Resolve(r)
		if err != nil {
			return err
		}
	}
	r.define(vs.name)
	return nil
}

func (ws WhileStmt) Resolve(r *Resolver) error {
	err := ws.condition.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}

	prevInLoop := r.inLoop
	r.inLoop = true
	err = ws.body.(Resolvable).Resolve(r)
	r.inLoop = prevInLoop
	return err
}

func (a AssignmentExpr) Resolve(r *Resolver) error {
	err := a.value.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	r.resolveLocal(a, a.name)
	return nil
}

func (b BinaryExpr) Resolve(r *Resolver) error {
	err := b.left.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	err = b.right.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	return nil
}

func (c CallExpr) Resolve(r *Resolver) error {
	err := c.callee.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	for _, arg := range c.arguments {
		err = arg.(Resolvable).Resolve(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g GetExpr) Resolve(r *Resolver) error {
	return g.object.(Resolvable).Resolve(r)
}

func (g GroupingExpr) Resolve(r *Resolver) error {
	return g.expression.(Resolvable).Resolve(r)
}

func (l LiteralExpr) Resolve(r *Resolver) error {
	return nil
}

func (l LogicalExpr) Resolve(r *Resolver) error {
	err := l.left.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	err = l.right.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	return nil
}

func (s SetExpr) Resolve(r *Resolver) error {
	err := s.value.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	err = s.obj.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	return nil
}

func (t TernaryExpr) Resolve(r *Resolver) error {
	err := t.condition.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	err = t.left.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	err = t.right.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}

	return nil
}

func (t ThisExpr) Resolve(r *Resolver) error {
	if r.currentClass == CLASSTYPE_NONE {
		return errors.NewAnalysisError(t.keyword, "Can't use 'this' outside of a class")
	}
	r.resolveLocal(t, t.keyword)
	return nil
}

func (u UnaryExpr) Resolve(r *Resolver) error {
	err := u.right.(Resolvable).Resolve(r)
	if err != nil {
		return err
	}
	return nil
}

func (v VariableExpr) Resolve(r *Resolver) error {
	if len(r.scopes) > 0 {
		def, ok := r.scopes[len(r.scopes)-1][v.name.Lexeme]
		if ok && !def.defined {
			return errors.NewAnalysisError(v.name, "Can't read local variable in its own initializer")
		}
	}

	r.resolveLocal(v, v.name)
	return nil
}
