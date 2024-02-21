package intpr

import (
	"simpl/ast"
	"simpl/errors"
	"simpl/tokens"
	"strconv"
)

type Memory []map[string]int

func NewMemory() *Memory {
	return &Memory{map[string]int{}}
}

func (m *Memory) Get(token tokens.Token) (int, *errors.RuntimeError) {
	for i := len(*m) - 1; i >= 0; i-- {
		val, found := (*m)[i][token.Value]
		if found {
			return val, nil
		}
	}
	return 0, &errors.RuntimeError{Message: "Undefined variable", Line: token.Line, Char: token.Char}
}

func (m *Memory) Set(token tokens.Token, opToken tokens.Token, value int) *errors.RuntimeError {
	name := token.Value
	_, defined := (*m)[len(*m)-1][name]
	if !defined {
		(*m)[len(*m)-1][name] = value
		return nil
	}
	return &errors.RuntimeError{Message: "Variable reassignment not allowed", Line: opToken.Line, Char: opToken.Char}
}

func (m *Memory) Update(token tokens.Token, value int) *errors.RuntimeError {
	name := token.Value
	for i := len(*m) - 1; i >= 0; i-- {
		_, found := (*m)[i][name]
		if found {
			(*m)[i][name] = value
			return nil
		}
	}
	return &errors.RuntimeError{Message: "Undefined variable", Line: token.Line, Char: token.Char}
}

func Run(mem *Memory, tree *ast.AST) *errors.RuntimeError {
	if tree.Scope < len(*mem)-1 {
		*mem = (*mem)[:tree.Scope+1]
	}
	for tree.Scope >= len(*mem) {
		*mem = append(*mem, map[string]int{})
	}
	node := tree.Root
	switch node.Token.Type {
	case tokens.COLON_EQUAL:
		value, err := eval(mem, node.Right)
		if err != nil {
			return err
		}
		return mem.Set(node.Left.Token, node.Token, value)
	default:
		value, err := eval(mem, node.Right)
		if err != nil {
			return err
		}
		return mem.Update(node.Left.Token, value)
	}
}

func eval(mem *Memory, node *ast.Node) (int, *errors.RuntimeError) {
	if node.Type == ast.Default {
		if node.Token.Type == tokens.IDENTIFIER {
			return mem.Get(node.Token)
		}
		val, err := strconv.Atoi(node.Token.Value)
		if err != nil {
			return 0, &errors.RuntimeError{Message: "Not a number", Line: node.Token.Line, Char: node.Token.Char}
		}
		return val, nil
	}
	left, err := eval(mem, node.Left)
	if err != nil {
		return 0, err
	}
	right, err := eval(mem, node.Right)
	if err != nil {
		return 0, err
	}
	switch node.Token.Type {
	case tokens.PLUS:
		return left + right, nil
	case tokens.MINUS:
		return left - right, nil
	case tokens.STAR:
		return left * right, nil
	default:
		return left / right, nil
	}
}
