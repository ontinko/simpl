package ast

import (
	"fmt"
	"simpl/errors"
	"simpl/memory"
	"simpl/tokens"
	"strconv"
)

type DataType int

const (
	Invalid DataType = iota
	Bool
	Int
)

type Statement interface {
	Execute(*memory.Memory) *errors.Error
	GetScope() int
	Prepare([]map[string]DataType) []*errors.Error
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

type Conditional struct {
	Statement
	Scope     int
	Token     tokens.Token
	Condition *Expression
	Then      *Program
	Else      *Program
}

type Assignment struct {
	Statement
	Scope    int
	DataType DataType
	Operator tokens.Token
	Var      tokens.Token
	Exp      *Expression
}

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

func (e *Expression) evalInt(mem *memory.Memory) (int, *errors.Error) {
	if e.DataType != Int {
		return 0, &errors.Error{Message: "Expected int", Type: errors.TypeError, Token: e.Token}
	}
	switch e.Token.Type {
	case tokens.NUMBER:
		val, err := strconv.Atoi(e.Token.Value)
		if err != nil {
			return 0, &errors.Error{Message: "NaN", Type: errors.TypeError, Token: e.Token}
		}
		return val, nil
	case tokens.IDENTIFIER:
		return mem.GetInt(e.Token), nil
	}

	left, err := e.Left.evalInt(mem)
	if err != nil {
		return 0, err
	}
	right, err := e.Right.evalInt(mem)
	if err != nil {
		return 0, err
	}

	switch e.Token.Type {
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

func (e *Expression) evalBool(mem *memory.Memory) (bool, *errors.Error) {
	if e.DataType != Bool {
		fmt.Println(e.Token.View())
		fmt.Println("Invalid?", e.DataType == Invalid)
		fmt.Println("Int?", e.DataType == Int)
		return false, &errors.Error{Message: "Expected bool", Type: errors.TypeError, Token: e.Token}
	}
	switch e.Token.Type {
	case tokens.TRUE:
		return true, nil
	case tokens.FALSE:
		return false, nil
	case tokens.BANG:
		value, err := e.Left.evalBool(mem)
		if err != nil {
			return false, err
		}
		return !value, nil
	case tokens.IDENTIFIER:
		return mem.GetBool(e.Token), nil
	case tokens.DOUBLE_EQUAL, tokens.NOT_EQUAL:
		if e.Left.DataType == Bool {
			left, err := e.Left.evalBool(mem)
			if err != nil {
				return false, err
			}
			right, err := e.Right.evalBool(mem)
			if err != nil {
				return false, err
			}
			if e.Token.Type == tokens.DOUBLE_EQUAL {
				return left == right, nil
			}
			return left != right, nil
		} else {
			left, err := e.Left.evalInt(mem)
			if err != nil {
				return false, err
			}
			right, err := e.Right.evalInt(mem)
			if err != nil {
				return false, err
			}
			if e.Token.Type == tokens.DOUBLE_EQUAL {
				return left == right, nil
			}
			return left != right, nil
		}
	case tokens.LESS:
		left, err := e.Left.evalInt(mem)
		if err != nil {
			return false, err
		}
		right, err := e.Right.evalInt(mem)
		if err != nil {
			return false, err
		}
		return left < right, nil
	case tokens.GREATER:
		left, err := e.Left.evalInt(mem)
		if err != nil {
			return false, err
		}
		right, err := e.Right.evalInt(mem)
		if err != nil {
			return false, err
		}
		return left > right, nil
	}

	left, err := e.Left.evalBool(mem)
	if err != nil {
		return false, err
	}
	right, err := e.Right.evalBool(mem)
	if err != nil {
		return false, err
	}

	switch e.Token.Type {
	case tokens.OR:
		return left || right, nil
	default:
		return left && right, nil
	}
}

func (s *Assignment) Execute(mem *memory.Memory) *errors.Error {
	mem.Resize(s.Scope)
	switch s.DataType {
	case Int:
		value, err := s.Exp.evalInt(mem)
		if err != nil {
			return err
		}
		switch s.Operator.Type {
		case tokens.EQUAL:
			mem.UpdateInt(s.Var, value)
		case tokens.COLON_EQUAL:
			mem.SetInt(s.Var, s.Operator, value)
		}
	case Bool:
		value, err := s.Exp.evalBool(mem)
		if err != nil {
			return err
		}
		switch s.Operator.Type {
		case tokens.EQUAL:
			mem.UpdateBool(s.Var, value)
		case tokens.COLON_EQUAL:
			mem.SetBool(s.Var, s.Operator, value)
		}
	}
	return nil
}

func (s *Conditional) Execute(mem *memory.Memory) *errors.Error {
	mem.Resize(s.Scope)
	switch s.Token.Type {
	case tokens.IF:
		condition, err := s.Condition.evalBool(mem)
		if err != nil {
			return err
		}
		if condition {
			for _, stmt := range s.Then.Statements {
				stmt.Execute(mem)
			}
		} else if s.Else != nil {
			for _, stmt := range s.Else.Statements {
				stmt.Execute(mem)
			}
		}
	default:
		then := false
		for {
			condition, err := s.Condition.evalBool(mem)
			if err != nil {
				return err
			}
			if condition {
                then = true
				for _, stmt := range s.Then.Statements {
					stmt.Execute(mem)
				}
			} else {
				break
			}

		}
		if !then && s.Else != nil {
			for _, stmt := range s.Else.Statements {
				stmt.Execute(mem)
			}
		}
	}
	return nil
}

func (s *Assignment) GetScope() int {
	return s.Scope
}

func (s *Conditional) GetScope() int {
	return s.Scope
}

func (e *Expression) Prepare(cache []map[string]DataType) []*errors.Error {
	if e == nil {
		return nil
	}

	errs := []*errors.Error{}
	errs = append(errs, e.Left.Prepare(cache)...)
	rightErrs := e.Right.Prepare(cache)
	switch e.Token.Type {
	case tokens.NUMBER:
		e.DataType = Int
	case tokens.TRUE, tokens.FALSE:
		e.DataType = Bool
	case tokens.IDENTIFIER:
		defined := false
		dataType := Invalid
		for i := len(cache) - 1; i >= 0; i-- {
			value, ok := cache[i][e.Token.Value]
			if ok {
				defined = true
				dataType = value
				break
			}
		}
		if !defined {
			errs = append(errs, &errors.Error{Message: "undefined variable", Token: e.Token, Type: errors.ReferenceError})
		}
		e.DataType = dataType
	case tokens.STAR, tokens.SLASH, tokens.PLUS, tokens.MINUS:
		e.DataType = Int
		if e.Left.DataType != Int && e.Left.DataType != Invalid || e.Right.DataType != Int && e.Right.DataType != Invalid {
			errs = append(errs, &errors.Error{Message: "invalid operation", Token: e.Token, Type: errors.TypeError})
		}
	case tokens.AND, tokens.OR:
		e.DataType = Bool
		if e.Left.DataType != Bool && e.Left.DataType != Invalid || e.Right.DataType != Bool && e.Right.DataType != Invalid {
			errs = append(errs, &errors.Error{Message: "invalid operation", Token: e.Token, Type: errors.TypeError})
		}
	case tokens.DOUBLE_EQUAL, tokens.NOT_EQUAL:
		e.DataType = Bool
		if e.Left.DataType != e.Right.DataType && e.Left.DataType != Invalid && e.Right.DataType != Invalid {
			errs = append(errs, &errors.Error{Message: "cannot compare values of different types", Token: e.Token, Type: errors.TypeError})
		}
	case tokens.LESS, tokens.GREATER:
		e.DataType = Bool
		if e.Left.DataType != Int && e.Left.DataType != Invalid || e.Right.DataType != Int && e.Right.DataType != Invalid {
			errs = append(errs, &errors.Error{Message: "invalid operation", Token: e.Token, Type: errors.TypeError})
		}
	case tokens.BANG:
		e.DataType = Bool
		if e.Left.DataType != Bool && e.Left.DataType != Invalid {
			errs = append(errs, &errors.Error{Message: "invalid negation: expected bool", Token: e.Token, Type: errors.TypeError})
		}
	}

	errs = append(errs, rightErrs...)
	return errs
}

func (s *Assignment) Prepare(cache []map[string]DataType) []*errors.Error {
	errs := []*errors.Error{}
	expErrs := s.Exp.Prepare(cache)
	switch s.Operator.Type {
	case tokens.COLON_EQUAL:
		_, defined := cache[len(cache)-1][s.Var.Value]
		if defined {
			errs = append(errs, &errors.Error{Message: "variable reassignment not allowed", Token: s.Var, Type: errors.SyntaxError})
		} else {
			cache[len(cache)-1][s.Var.Value] = s.Exp.DataType
			s.DataType = s.Exp.DataType
		}
	case tokens.EQUAL:
		defined := false
		var dataType DataType
		for i := len(cache) - 1; i >= 0; i-- {
			value, ok := cache[i][s.Var.Value]
			if ok {
				defined = true
				dataType = value
				break
			}
		}
		if !defined {
			errs = append(errs, &errors.Error{Message: "undefined variable", Token: s.Var, Type: errors.ReferenceError})
		} else if dataType != s.Exp.DataType && s.Exp.DataType != Invalid {
			errs = append(errs, &errors.Error{Message: "assigning wrong type", Token: s.Var, Type: errors.TypeError})
		} else {
			s.DataType = s.Exp.DataType
		}
	}
	errs = append(errs, expErrs...)
	return errs
}

func (s *Conditional) Prepare(cache []map[string]DataType) []*errors.Error {
	errs := []*errors.Error{}
	condErrs := s.Condition.Prepare(cache)
	if s.Condition.DataType != Bool {
		errs = append(errs, &errors.Error{Message: "invalid condition type: expected bool", Token: s.Condition.Token, Type: errors.SyntaxError})
	}
	errs = append(errs, condErrs...)
	errs = append(errs, s.Condition.Prepare(cache)...)
	errs = append(errs, s.Then.Prepare(cache)...)
	errs = append(errs, s.Else.Prepare(cache)...)

	return errs
}

func (p *Program) Prepare(cache []map[string]DataType) []*errors.Error {
	if p == nil {
		return nil
	}

	errs := []*errors.Error{}
	statements := p.Statements

	for _, s := range statements {
		scope := s.GetScope()
		for len(cache) <= scope {
			cache = append(cache, map[string]DataType{})
		}
		if len(cache) > scope+1 {
			cache = cache[:scope+1]
		}
		errs = append(errs, s.Prepare(cache)...)
	}

	return errs
}
