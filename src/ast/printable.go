package ast

import (
	"fmt"
	"strings"
)

type Printable interface {
	Print() string
}

func (l Literal) Print() string {
	if l.value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", l.value)
}

func (u Unary) Print() string {
	return parenthesize(u.operator.Lexeme, u.right.(Printable))
}

func (b Binary) Print() string {
	return parenthesize(b.operator.Lexeme, b.left.(Printable), b.right.(Printable))
}

func (t Ternary) Print() string {
	return parenthesize("?:", t.condition.(Printable), t.left.(Printable), t.right.(Printable))
}

func (g Grouping) Print() string {
	return parenthesize("group", g.expression.(Printable))
}

func parenthesize(name string, exprs ...Printable) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "(%s", name)
	for _, expr := range exprs {
		fmt.Fprintf(&sb, " %s", expr.Print())
	}
	fmt.Fprintf(&sb, ")")

	return sb.String()
}
