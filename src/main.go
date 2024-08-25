package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/faideww/glox/src/ast"
	"github.com/faideww/glox/src/errors"
)

var interpreter *ast.Interpreter

func main() {
	var err error
	if len(os.Args) > 2 {
		fmt.Printf("Usage: glox [script]\n")
		os.Exit(64)
	} else if len(os.Args) == 2 {
		err = runFile(os.Args[1])
	} else {
		err = runPrompt()
	}

	if err != nil {
		panic(err)
	}
}

func runFile(fp string) error {
	bytes, err := os.ReadFile(fp)
	if err != nil {
		return err
	}
	interpreter = ast.NewInterpreter()
	err = runProgram(string(bytes))
	if _, ok := err.(*errors.ParserError); ok {
		os.Exit(65)
	}
	if _, ok := err.(*errors.RuntimeError); ok {
		os.Exit(70)
	}
	return nil
}

func runPrompt() error {
	buffer := bufio.NewReader(os.Stdin)
	interpreter = ast.NewInterpreter()

	for {
		var err error
		fmt.Printf("> ")
		line, err := buffer.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		runRepl(line)
	}
	return nil
}

func runRepl(source string) error {
	scanner := NewScanner(source)
	tokens, scanErr := scanner.ScanTokens()
	if scanErr != nil {
		return scanErr
	}

	reporter := errors.NewErrorReporter()

	parser := ast.NewParser(tokens, reporter)

	// try to parse a single expression first
	expr, parseOk := parser.ParseExpression()

	if parseOk {
		// fmt.Printf("Expr: %+v\n", expr)
		value, runtimeErr := interpreter.InterpretExpression(expr)
		if runtimeErr != nil {
			return runtimeErr
		}

		fmt.Println(ast.ToString(value))
		return nil
	}

	// if that fails, try to parse it as statements instead
	reporter.Clear()
	return runProgram(source)
}

func runProgram(source string) error {
	scanner := NewScanner(source)
	tokens, scanErr := scanner.ScanTokens()
	if scanErr != nil {
		return scanErr
	}

	reporter := errors.NewErrorReporter()
	parser := ast.NewParser(tokens, reporter)
	statements, parseOk := parser.Parse()

	// fmt.Println("Stmts:")
	// for _, statement := range statements {
	// 	fmt.Printf("%#v\n", statement)
	// }

	if !parseOk {
		reporter.Report(os.Stdout)
		return reporter.Last()
	}

	resolver := ast.NewResolver(interpreter)
	resolveErr := resolver.Resolve(statements)

	if resolveErr != nil {
		fmt.Println(resolveErr)
		return resolveErr
	}

	runtimeErr := interpreter.Interpret(statements)
	if runtimeErr != nil {
		fmt.Println(runtimeErr)
		return runtimeErr
	}

	return nil
}
