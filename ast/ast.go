package ast

import (
	"fmt"
	"simpl/errors"
	"simpl/tokens"
)

type NodeType int

const (
	Default NodeType = iota
	Expression
	Statement
)

type Node struct {
	Token tokens.Token
	Type  NodeType
	Left  *Node
	Right *Node
}

type AST struct {
	Root       *Node
	prevParent *Node
	prev       *Node
}

func NewAST() AST {
	return AST{nil, nil, nil}
}

func (t *AST) Insert(node *Node) *errors.SyntaxError {
	var err *errors.SyntaxError
	switch node.Token.Type {
	case tokens.LEFT_BRACE, tokens.RIGHT_BRACE:
		return nil
	case tokens.SEMICOLON:
		if t.Root == nil || t.Root.Type != Statement || t.prev.Type != Default {
			return &errors.SyntaxError{Message: "Unexpected ;", Line: node.Token.Line, Char: node.Token.Char}
		}
	case tokens.COLON_EQUAL, tokens.EQUAL:
		if t.Root == nil || t.prevParent != nil || NodeType(t.prev.Token.Type) != NodeType(tokens.IDENTIFIER) {
			return &errors.SyntaxError{Message: "Unexpected assignment", Line: node.Token.Line, Char: node.Token.Char}
		}
		node.Left = t.prev
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SLASH:
		if t.Root == nil || t.Root.Type != Statement || t.prev.Type != Default {
			return &errors.SyntaxError{Message: "Unexpected operator", Line: node.Token.Line, Char: node.Token.Char}
		}
		node.Left = t.prev
		t.prevParent.Right = node
	case tokens.IDENTIFIER:
		if t.Root != nil && t.prev.Type == Default {
			return &errors.SyntaxError{Message: "Unexpected identifier", Line: node.Token.Line, Char: node.Token.Char}
		}
		if t.prev != nil {
			t.prev.Right = node
		}
	case tokens.NUMBER:
		if t.Root == nil || t.prev.Type == Default {
			return &errors.SyntaxError{Message: "Unexpected number", Line: node.Token.Line, Char: node.Token.Char}
		}
		t.prev.Right = node
	}
	if t.Root == nil || t.Root.Type != Statement {
		t.Root = node
	}
	t.prevParent = t.prev
	t.prev = node
	return err
}

func (t *AST) Traverse() {
	node := t.Root
	if node == nil {
		return
	}
	stack := []*Node{}
	for {
		if node != nil {
			stack = append(stack, node)
			node = node.Left
			continue
		}
		if len(stack) > 0 {
			node = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			node.Token.Print()
			node = node.Right
			continue
		}
		break
	}
}

func (root *Node) Visualize() {
	if root == nil {
		return
	}
	levels := [][]*Node{}
	queue := []*Node{root}
	for len(queue) > 0 {
		queueSize := len(queue)
		nextLevel := []*Node{}
		for i := 0; i < queueSize; i++ {
			node := queue[i]
			if node != nil {
				nextLevel = append(nextLevel, node.Left, node.Right)
			}
		}
		levels = append(levels, queue)
		queue = nextLevel
	}
	tab := 0
	linkTab := 0
	fmt.Println()
	for i := len(levels) - 2; i >= 0; i-- {
		for j := 0; j < tab; j++ {
			fmt.Print(" ")
		}
		tab = tab*2 + 1
		for _, n := range levels[i] {
			if n == nil {
				fmt.Print(" ")
			} else {
				val, found := tokens.Representations[n.Token.Type]
				if found {
					fmt.Print(val)
				} else if n.Token.Type == tokens.UNPERMITTED {
					fmt.Printf("X: %s", n.Token.Value)
				} else {
					fmt.Print(n.Token.Value)
				}
			}
			for j := 0; j < tab; j++ {
				fmt.Print(" ")
			}
		}
		fmt.Println()
		if i == 0 {
			break
		}
		for j := 0; j < linkTab; j++ {
			fmt.Print(" ")
		}
		linkTab = linkTab*2 + 1
		left := true
		for _, n := range levels[i] {
			if n == nil {
				fmt.Print(" ")
			} else {
				if left {
					fmt.Print("\\")
				} else {
					fmt.Print("/")
				}
				left = !left
			}
			for j := 0; j < linkTab; j++ {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}
