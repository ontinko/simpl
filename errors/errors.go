package errors

import "fmt"

type RuntimeError struct {
	Message string
	Line    int
	Char    int
}

type SyntaxError struct {
	Message string
	Line    int
	Char    int
}

func (e *RuntimeError) Print() {
	fmt.Printf("Runtime error: %d:%d %s\n", e.Line, e.Char, e.Message)
}

func (e *SyntaxError) Print() {
	fmt.Printf("Syntax error: %d:%d %s\n", e.Line, e.Char, e.Message)
}
