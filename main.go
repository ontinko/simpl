package main

import (
	"fmt"
	"os"
	"time"

	"simpl/analyzer"
	"simpl/intpr"
	"simpl/lexer"
	"simpl/parser"
)

func main() {
	execute := true
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: simpl [script]")
		os.Exit(64)
	}
	filename := args[0]
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("File not found")
		os.Exit(64)
	}
	tokens, errs := lexer.Tokenize(string(source), filename, 1)
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
	staticErrs := analyzer.Prepare(logic)
	if len(staticErrs) > 0 {
		for _, e := range staticErrs {
			e.Print()
		}
		os.Exit(64)
	}
	if execute {
		start := time.Now()
		for _, tree := range logic {
			intprErr := intpr.Run(memory, tree)
			if intprErr != nil {
				intprErr.Print()
				os.Exit(64)
			}
		}
		elapsed := time.Since(start)
		fmt.Println("Elapsed:", elapsed)
		fmt.Println("Results:")
		memory.Print()
	} else {
		for _, tree := range logic {
			tree.Root.Visualize()
		}
	}
}
