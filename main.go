package main

import (
	"fmt"
	"os"
	"simpl/intpr"
	"simpl/lexer"
	"simpl/parser"
	"time"
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
	startTime := time.Now()
	tokens, errs := lexer.Tokenize(string(source), filename, 1)
	if len(errs) > 0 {
		for _, e := range errs {
			e.Print()
		}
		return
	}
	memory := intpr.NewMemory()

	parseSource := parser.New(tokens)
	program, error := parseSource.Parse(false)
	if error != nil {
		error.Print()
		os.Exit(64)
		return
	}
	if len(parseSource.Errors) > 0 {
		for _, e := range parseSource.Errors {
			e.Print()
		}
		os.Exit(64)
	}
	elapsed := time.Since(startTime)
	fmt.Println("Time elapsed for parsing:", elapsed)
	if execute {
		start := time.Now()
		for _, stmt := range program.Statements {
			err := stmt.Execute(memory)
			if err != nil {
				err.Print()
				os.Exit(64)
			}
		}
		elapsed := time.Since(start)
		fmt.Println("Elapsed:", elapsed)
		fmt.Println("Results:")
		memory.Print()
	} else {
		for _, stmt := range program.Statements {
			stmt.Visualize()
		}
	}
}
