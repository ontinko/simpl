package errors

import (
	"fmt"
	"simpl/tokens"
)

type ErrorType int

const (
	SyntaxError ErrorType = iota + 1
	RuntimeError
	TypeError
	ReferenceError
)

type Error struct {
	Type    ErrorType
	Message string
	Token   tokens.Token
}

func (e *Error) Print() {
	token := e.Token
	var errorType string
	switch e.Type {
	case SyntaxError:
		errorType = "syntax error"
	case TypeError:
		errorType = "type error"
	case ReferenceError:
		errorType = "reference error"
	default:
		errorType = "runtime error"
	}
	fmt.Printf("%s:%d:%d: %s: %s\n", token.Filename, token.Line, token.Char, errorType, e.Message)
}
