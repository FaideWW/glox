package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/faideww/glox/src/ast"
)

var hadError bool
var interpreter ast.Interpreter

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
	err = run(string(bytes))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	if _, ok := err.(*ast.ParserError); ok {
		os.Exit(65)
	}
	if _, ok := err.(*ast.RuntimeError); ok {
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
		fmt.Printf("%+v", line)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		err = run(line)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func run(source string) error {
	scanner := NewScanner(source)
	tokens, scanErr := scanner.ScanTokens()
	if scanErr != nil {
		return scanErr
	}

	parser := ast.NewParser(tokens)
	statements, parseErr := parser.Parse()

	// for _, statement := range statements {
	// 	fmt.Printf("%+v\n", statement)
	// }

	if parseErr != nil {
		return parseErr
	}

	runtimeErr := interpreter.Interpret(statements)
	if runtimeErr != nil {
		return runtimeErr
	}

	return nil
}
