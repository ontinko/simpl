package intpr

import (
	"fmt"
	"simpl/ast"
	"simpl/errors"
	"simpl/tokens"
	"strconv"
)

type Memory struct {
	maxScope int
	ints     []map[string]int
	bools    []map[string]bool
}

func NewMemory() *Memory {
	return &Memory{maxScope: 0, ints: []map[string]int{{}}, bools: []map[string]bool{{}}}
}

// TODO: remove duplication if possible

func (m *Memory) getBool(token tokens.Token) bool {
	data := m.bools
	var result bool
	for i := len(data) - 1; i >= 0; i-- {
		val, found := data[i][token.Value]
		if found {
			result = val
		}
	}
	return result
}

func (m *Memory) getNumber(token tokens.Token) int {
	data := m.ints
	var result int
	for i := len(data) - 1; i >= 0; i-- {
		val, found := data[i][token.Value]
		if found {
			result = val
		}
	}
	return result
}

func (m *Memory) setInt(token tokens.Token, opToken tokens.Token, value int) {
	name := token.Value
	m.ints[len(m.ints)-1][name] = value
}

func (m *Memory) setBool(token tokens.Token, opToken tokens.Token, value bool) {
	name := token.Value
	m.bools[len(m.bools)-1][name] = value
}

func (m *Memory) updateInt(token tokens.Token, value int) {
	name := token.Value
	for i := len(m.ints) - 1; i >= 0; i-- {
		_, found := m.ints[i][name]
		if found {
			m.ints[i][name] = value
			break
		}
	}
}

func (m *Memory) updateBool(token tokens.Token, value bool) {
	name := token.Value
	for i := len(m.bools) - 1; i >= 0; i-- {
		_, found := m.bools[i][name]
		if found {
			m.bools[i][name] = value
			break
		}
	}
}

func (m *Memory) Print() {
	for _, data := range m.ints {
		for k, v := range data {
			fmt.Printf("%s: %d\n", k, v)
		}
	}
	for _, data := range m.bools {
		for k, v := range data {
			fmt.Printf("%s: %t\n", k, v)
		}
	}
}

// Fix: variables don't update

func Run(mem *Memory, tree *ast.AST) *errors.Error {
	if tree.Scope < mem.maxScope {
		mem.ints = mem.ints[:tree.Scope+1]
		mem.bools = mem.bools[:tree.Scope+1]
		mem.maxScope = tree.Scope
	}
	for tree.Scope > mem.maxScope {
		mem.ints = append(mem.ints, map[string]int{})
		mem.bools = append(mem.bools, map[string]bool{})
		mem.maxScope++
	}
	node := tree.Root
	switch node.Left.DataType {
	case ast.Number:
		value, err := evalNum(mem, node.Right)
		if err != nil {
			return err
		}
		if node.Token.Type == tokens.COLON_EQUAL {
			mem.setInt(node.Left.Token, node.Token, value)
		} else if node.Token.Type == tokens.EQUAL {
			mem.updateInt(node.Left.Token, value)
		}
	case ast.Bool:
		value, err := evalBool(mem, node.Right)
		if err != nil {
			return err
		}
		if node.Token.Type == tokens.COLON_EQUAL {
			mem.setBool(node.Left.Token, node.Token, value)
		} else if node.Token.Type == tokens.EQUAL {
			mem.updateBool(node.Left.Token, value)
		}
	}
	return nil
}

func evalBool(mem *Memory, node *ast.Node) (bool, *errors.Error) {
	if node.Token.Type == tokens.TRUE {
		return true, nil
	}
	return false, nil
}

func evalNum(mem *Memory, node *ast.Node) (int, *errors.Error) {
	if node.Type == ast.Value {
		if node.Token.Type == tokens.IDENTIFIER {
			return mem.getNumber(node.Token), nil
		}
		val, err := strconv.Atoi(node.Token.Value)
		if err != nil {
			return 0, &errors.Error{Message: "Not a number", Token: node.Token, Type: errors.RuntimeError}
		}
		return val, nil
	}
	left, err := evalNum(mem, node.Left)
	if err != nil {
		return 0, err
	}
	right, err := evalNum(mem, node.Right)
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
