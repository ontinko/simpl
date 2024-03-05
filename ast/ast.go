package ast

import (
	"simpl/errors"
	"simpl/memory"
	"simpl/tokens"
)

type DataType int

const (
	Invalid DataType = iota
	Bool
	Int
)

type Statement interface {
	Execute(*memory.Memory) *errors.Error
	Prepare(*[]map[string]DataType) []*errors.Error
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

type Break struct {
	Statement
}

type Continue struct {
	Statement
}

func (s *Assignment) GetScope() int {
	return s.Scope
}

func (s *Conditional) GetScope() int {
	return s.Scope
}

func (s *For) GetScope() int {
	return s.Scope
}
