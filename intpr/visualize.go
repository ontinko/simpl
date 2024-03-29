package intpr

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
				fmt.Printf("%s: %s", n.Token.View(), dType)
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
		case Bool:
			fmt.Print("bool ")
		default:
			fmt.Print("unknown")
		}
	}
	fmt.Printf("%s:", s.Var.View())
	switch s.DataType {
	case Int:
		fmt.Print("int ")
	case Bool:
		fmt.Print("bool ")
	default:
		fmt.Print("unknown ")
	}
	fmt.Println(s.Operator.View(), "SCOPE", s.VarScope)
	s.Exp.Visualize()
}

func (s *Conditional) Visualize() {
	fmt.Printf("%s:\n", s.Token.View())
	s.Condition.Visualize()
	fmt.Println("then:")
	if s.Then != nil {
		for _, stmt := range s.Then.Statements {
			stmt.Visualize()
		}
	}
	fmt.Println("else:")
	if s.Else != nil {
		for _, stmt := range s.Else.Statements {
			stmt.Visualize()
		}
	}
}

func (s *For) Visualize() {
	fmt.Println("For statemt")
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
	fmt.Println("ForEnd")
}

func (s *Def) Visualize() {
	fmt.Printf("def %s %s, scope: %d\n", s.NameToken.Value, s.DataType.View(), s.Scope)
	fmt.Print("params: ")
	paramsLen := len(s.Params)
	for i, p := range s.Params {
		fmt.Printf("%s %s", p.DataType.View(), p.NameToken.Value)
		if i != paramsLen-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println()
	fmt.Println("body:")
	if s.Body != nil {
		for _, s := range s.Body.Statements {
			s.Visualize()
		}
	}
	fmt.Println("defEnd")
}

func (f *Function) Visualize() {
	fmt.Printf("%s: (", f.DataType.View())
	for i, a := range f.Params {
		fmt.Printf("%s %s", a.DataType.View(), a.NameToken.Value)
		if i < len(f.Params)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println(")")
	fmt.Println("Body:")
	if f.Body != nil {
		for _, s := range f.Body.Statements {
			s.Visualize()
		}
	}
	fmt.Println("funcEnd")
}

func (s *Return) Visualize() {
	fmt.Printf("return: %s\n", s.DataType.View())
}

func (s *Break) Visualize() {
	fmt.Println("break")
}

func (s *Continue) Visualize() {
	fmt.Println("continue")
}

func (s *VoidCall) Visualize() {
	fmt.Printf("%s(): void", s.NameToken.Value)
	fmt.Println("Arguments:")
	for _, a := range s.Args {
		a.Visualize()
	}
	fmt.Println("voidcallEnd")
}

func (s *OpenScope) Visualize() {
	fmt.Println("{")
}

func (s *CloseScope) Visualize() {
	fmt.Println("}")
}
