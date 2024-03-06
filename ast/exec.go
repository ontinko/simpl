package ast

import (
	"simpl/errors"
	"simpl/memory"
	"simpl/tokens"
	"strconv"
)

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
		return mem.GetInt(e.Token)
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
	case tokens.SLASH:
		if right == 0 {
			return 0, &errors.Error{Message: "zero division not allowed", Token: e.Token, Type: errors.RuntimeError}
		}
		return left / right, nil
	default:
		if right == 0 {
			return 0, &errors.Error{Message: "zero division not allowed", Token: e.Token, Type: errors.RuntimeError}
		}
		return left % right, nil
	}
}

func (e *Expression) evalBool(mem *memory.Memory) (bool, *errors.Error) {
	if e.DataType != Bool {
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
		mem.GetBool(e.Token)
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
	case tokens.LESS, tokens.LESS_EQUAL, tokens.GREATER, tokens.GREATER_EQUAL:
		left, err := e.Left.evalInt(mem)
		if err != nil {
			return false, err
		}
		right, err := e.Right.evalInt(mem)
		if err != nil {
			return false, err
		}
		switch e.Token.Type {
		case tokens.LESS:
			return left < right, nil
		case tokens.GREATER:
			return left > right, nil
		case tokens.LESS_EQUAL:
			return left <= right, nil
		default:
			return left >= right, nil
		}
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
		switch s.Operator.Type {
		case tokens.DOUBLE_PLUS:
			mem.IncInt(s.Var, 1)
			return nil
		case tokens.DOUBLE_MINUS:
			mem.DecInt(s.Var, 1)
			return nil
		}

		value, err := s.Exp.evalInt(mem)
		if err != nil {
			return err
		}
		switch s.Operator.Type {
		case tokens.EQUAL:
			if s.Explicit {
				mem.SetInt(s.Var, s.Operator, value)
			} else {
				mem.UpdateInt(s.Var, value)
			}
		case tokens.PLUS_EQUAL:
			mem.IncInt(s.Var, value)
		case tokens.MINUS_EQUAL:
			mem.DecInt(s.Var, value)
		case tokens.STAR_EQUAL:
			mem.MulInt(s.Var, value)
		case tokens.SLASH_EQUAL:
			if value == 0 {
				return &errors.Error{Message: "zero division not allowed", Token: s.Operator, Type: errors.RuntimeError}
			}
			mem.DivInt(s.Var, value)
		case tokens.MODULO_EQUAL:
			if value == 0 {
				return &errors.Error{Message: "zero division not allowed", Token: s.Operator, Type: errors.RuntimeError}
			}
			mem.ModInt(s.Var, value)
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
			if s.Explicit {
				mem.SetBool(s.Var, s.Operator, value)
			} else {
				mem.UpdateBool(s.Var, value)
			}
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
				err := stmt.Execute(mem)
				if err != nil {
					return err
				}
			}
		} else if s.Else != nil {
			for _, stmt := range s.Else.Statements {
				err := stmt.Execute(mem)
				if err != nil {
					return err
				}
			}
		}
	default:
		then := false
	WhileLoop:
		for {
			condition, err := s.Condition.evalBool(mem)
			if err != nil {
				return err
			}
			if condition {
				then = true
				for _, stmt := range s.Then.Statements {
					err := stmt.Execute(mem)
					if err != nil {
						switch err.Type {
						case errors.Break:
							return nil
						case errors.Continue:
							continue WhileLoop
						default:
							return err
						}
					}
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

func (s *For) Execute(mem *memory.Memory) *errors.Error {
	err := s.Init.Execute(mem)
	if err != nil {
		return err
	}
ForLoop:
	for {
		condition, err := s.Condition.evalBool(mem)
		if err != nil {
			return err
		}
		if condition {
			if s.Block != nil {
				for _, stmt := range s.Block.Statements {
					err := stmt.Execute(mem)
					if err != nil {
						switch err.Type {
						case errors.Break:
							return nil
						case errors.Continue:
							err := s.After.Execute(mem)
							if err != nil {
								return err
							}
							continue ForLoop
						}
						return err
					}
				}
			}
			err := s.After.Execute(mem)
			if err != nil {
				return err
			}
			continue
		}
		break
	}
	return nil
}

func (s *Break) Execute(mem *memory.Memory) *errors.Error {
	return &errors.Error{Type: errors.Break}
}

func (s *Continue) Execute(mem *memory.Memory) *errors.Error {
	return &errors.Error{Type: errors.Continue}
}
