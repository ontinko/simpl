package intpr

import (
	"simpl/ast"
	"simpl/errors"
	"simpl/tokens"
	"strconv"
)

type Memory struct {
	Vars map[string]int
}

func NewMemory() *Memory {
	return &Memory{map[string]int{}}
}

func (m *Memory) Get(token tokens.Token) (int, *errors.RuntimeError) {
	value, found := m.Vars[token.Value]
	if !found {
		return 0, &errors.RuntimeError{Message: "Undefined variable", Line: token.Line, Char: token.Char}
	}
	return value, nil
}

func Run(mem *Memory, tree *ast.AST) *errors.RuntimeError {
	node := tree.Root
	switch node.Token.Type {
	case tokens.COLON_EQUAL:
		_, found := mem.Vars[node.Left.Token.Value]
		if found {
			return &errors.RuntimeError{Message: "Variable reassignment not allowed", Line: node.Token.Line, Char: node.Token.Char}
		}
		value, err := eval(mem, node.Right)
		if err != nil {
			return err
		}
		mem.Vars[node.Left.Token.Value] = value
	default:
		_, found := mem.Vars[node.Left.Token.Value]
		if !found {
			return &errors.RuntimeError{Message: "Undefined variable", Line: node.Token.Line, Char: node.Token.Char}
		}
		value, err := eval(mem, node.Right)
		if err != nil {
			return err
		}
		mem.Vars[node.Left.Token.Value] = value
	}
	return nil
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
