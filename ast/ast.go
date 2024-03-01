package ast

import (
	"fmt"
	"simpl/tokens"
)

type NodeType int

type DataType int

const (
	Value NodeType = iota
	Expression
	Statement
)

const (
	Number DataType = iota + 1
	Bool
)

type Node struct {
	Token    tokens.Token
	Type     NodeType
	DataType DataType
	Left     *Node
	Right    *Node
	Level    int
}

type AST struct {
	Scope int
	Root  *Node
}

func NewAST() AST {
	return AST{}
}

func (n *Node) SetTypes() {
	if n.Left == nil && n.Right == nil {
		switch n.Token.Type {
		case tokens.NUMBER:
			n.DataType = Number
		case tokens.TRUE, tokens.FALSE:
			n.DataType = Number
		}
		return
	}
	switch n.Type {
	}
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
			dataType := ""
			switch node.DataType {
			case Bool:
				dataType = "bool"
			case Number:
				dataType = "int"
			default:
				dataType = "none"
			}
			fmt.Printf("%s:%s\n", node.Token.View(), dataType)
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
				dataType := ""
				switch n.DataType {
				case Number:
					dataType = "int"
				case Bool:
					dataType = "bool"
				default:
					dataType = "none"
				}
				val, found := tokens.Representations[n.Token.Type]
				if found {
					fmt.Printf("%s: %s", val, dataType)
				} else if n.Token.Type == tokens.UNPERMITTED {
					fmt.Printf("X: %s", n.Token.Value)
				} else {
					fmt.Printf("%s: %s", n.Token.Value, dataType)
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
