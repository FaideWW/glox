package ast

// Supports resolving non-primitive types (functions,
// classes) to name strings, so that they can be nicely printed
// to output
type Named interface {
	String() string
}
