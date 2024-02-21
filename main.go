package main

import (
	"fmt"
	"os"
	"simpl/ast"
	"simpl/intpr"
	"simpl/lexer"
	"simpl/parser"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: simpl [script]")
		os.Exit(64)
	}
	source, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Println("File not found")
		os.Exit(64)
	}
	tokens, errs := lexer.Tokenize(string(source), 0)
	if len(errs) > 0 {
		for _, e := range errs {
			e.Print()
		}
		return
	}
	memory := intpr.NewMemory()

	logic := []*ast.AST{}
	scope := 0
	for len(tokens) != 0 {
		tree, astErr := parser.Parse(&tokens, &scope)
		if astErr != nil {
			astErr.Print()
			break
		}
		logic = append(logic, tree)
	}
	for _, tree := range logic {
		intprErr := intpr.Run(memory, tree)
		if intprErr != nil {
			intprErr.Print()
			break
		}
	}
	fmt.Println("Memory:")
	for _, m := range *memory {
		fmt.Println(m)
	}
}
