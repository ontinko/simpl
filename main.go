package main

import (
	"fmt"
	"os"
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
	tokens, errs := lexer.Tokenize(string(source), 1)
	if len(errs) > 0 {
		for _, e := range errs {
			e.Print()
		}
		return
	}
	memory := intpr.NewMemory()

	parseSource := parser.New(tokens)
	logic, error := parseSource.Parse()
	if error != nil {
		error.Print()
		os.Exit(64)
		return
	}
	for _, tree := range logic {
		intprErr := intpr.Run(memory, tree)
		if intprErr != nil {
			intprErr.Print()
			os.Exit(64)
		}
	}
	fmt.Println("Results:")
	for _, m := range *memory {
		for k, v := range m {
			fmt.Printf("%s = %d\n", k, v)
		}
	}
}
