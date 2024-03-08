package intpr

import (
	"simpl/errors"
	"simpl/tokens"
)

type DataType int

const (
	Invalid DataType = iota
	Bool
	Int
	Void
	Func
)

func (t DataType) View() string {
	switch t {
	case Invalid:
		return "invalid"
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Void:
		return "void"
	default:
		return "unknown"
	}
}

type Statement interface {
	Execute(*Memory) *errors.Error
	Visualize()
}

type Program struct {
	Statements []Statement
}

type Expression struct {
	DataType DataType
	Token    tokens.Token
	Left     *Expression
	Right    *Expression
}

type Assignment struct {
	Statement
	Explicit bool
	Scope    int
	DataType DataType
	Operator tokens.Token
	Var      tokens.Token
	Exp      *Expression
}

type Conditional struct {
	Statement
	Scope     int
	Token     tokens.Token
	Condition *Expression
	Then      *Program
	Else      *Program
}

type For struct {
	Statement
	Scope     int
	Token     tokens.Token
	Init      *Assignment
	Condition *Expression
	After     *Assignment
	Block     *Program
}

type Def struct {
	Statement
	Scope     int
	DataType  DataType
	Token     tokens.Token
	NameToken tokens.Token
	Params    []DefParam
	Body      *Program
}

type DefParam struct {
	NameToken tokens.Token
	DataType  DataType
}

type Function struct {
	Scope    int
	DataType DataType
	Params   []DefParam
	Body     *Program
}

type Break struct {
	Statement
}

type Continue struct {
	Statement
}

type Return struct {
	Statement
	Scope    int
	DataType DataType
	Exp      *Expression
}
