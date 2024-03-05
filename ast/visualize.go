package ast

import (
	"fmt"
)

func (e *Expression) Visualize() {
	if e == nil {
		return
	}
	levels := [][]*Expression{}
	queue := []*Expression{e}
	for len(queue) > 0 {
		queueSize := len(queue)
		nextLevel := []*Expression{}
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
	for i := len(levels) - 2; i >= 0; i-- {
		for j := 0; j < tab; j++ {
			fmt.Print(" ")
		}
		tab = tab*2 + 1
		for _, n := range levels[i] {
			if n == nil {
				fmt.Print(" ")
			} else {
				dType := ""
				switch n.DataType {
				case Int:
					dType = "int"
				default:
					dType = "bool"
				}
				fmt.Printf("%s -> %s", n.Token.View(), dType)
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
	fmt.Println()
}

func (s *Assignment) Visualize() {
	if s.Explicit {
		switch s.DataType {
		case Int:
			fmt.Print("int ")
		default:
			fmt.Print("bool ")
		}
	}
	fmt.Printf("%s %s\n", s.Var.View(), s.Operator.View())
	s.Exp.Visualize()
}

func (s *Conditional) Visualize() {
	fmt.Printf("%s:\n", s.Token.View())
	s.Condition.Visualize()
	fmt.Println("then:")
	for _, stmt := range s.Then.Statements {
		stmt.Visualize()
	}
	fmt.Println("else:")
	for _, stmt := range s.Else.Statements {
		stmt.Visualize()
	}
}

func (s *For) Visualize() {
	fmt.Println("Init:")
	s.Init.Visualize()
	fmt.Println("Condition:")
	s.Condition.Visualize()
	fmt.Println("After:")
	s.After.Visualize()
	fmt.Println("Block:")
	if s.Block != nil {
		for _, stmt := range s.Block.Statements {
			stmt.Visualize()
		}
	}
}

func (s *Break) Visualize() {
	fmt.Println("break")
}

func (s *Continue) Visualize() {
	fmt.Println("continue")
}
