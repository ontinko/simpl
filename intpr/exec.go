package intpr

import (
	"simpl/errors"
	"simpl/tokens"
	"strconv"
)

func (e *Expression) evalInt(mem *Memory) (int, *errors.Error) {
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
		if e.Args == nil {
			return mem.GetInt(e.Token, e.Scope)
		}
		fn, err := mem.GetFunc(e.Token, e.Scope)
		if err != nil {
			return 0, err
		}
		mem.Extend()
		for i, p := range fn.Params {
			switch p.DataType {
			case Int:
				val, err := e.Args[i].evalInt(mem)
				if err != nil {
					return 0, err
				}
				mem.SetInt(p.NameToken, val)
			case Bool:
				val, err := e.Args[i].evalBool(mem)
				if err != nil {
					return 0, err
				}
				mem.SetBool(p.NameToken, val)
			default:
				return 0, &errors.Error{Message: "unexpected argument type", Type: errors.RuntimeError, Token: p.NameToken}
			}
		}
		var returnResult int
		var returnErr *errors.Error
		for _, s := range fn.Body.Statements {
			err := s.Execute(mem)
			if err != nil {
				if err.Type == errors.Return {
					exp := fn.Returns[err.MessageId]
					returnResult, returnErr = exp.evalInt(mem)
				}
				break
			}
		}
		mem.Shrink()
		return returnResult, returnErr
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

func (e *Expression) evalBool(mem *Memory) (bool, *errors.Error) {
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
		if e.Args == nil {
			return mem.GetBool(e.Token, e.Scope)
		}
		fn, err := mem.GetFunc(e.Token, e.Scope)
		if err != nil {
			return false, err
		}
		mem.Extend()
		for i, p := range fn.Params {
			switch p.DataType {
			case Int:
				val, err := e.Args[i].evalInt(mem)
				if err != nil {
					return false, err
				}
				mem.SetInt(p.NameToken, val)
			case Bool:
				val, err := e.Args[i].evalBool(mem)
				if err != nil {
					return false, err
				}
				mem.SetBool(p.NameToken, val)
			default:
				return false, &errors.Error{Message: "unexpected argument type", Type: errors.RuntimeError, Token: p.NameToken}
			}
		}
		var returnResult bool
		var returnErr *errors.Error
		for _, s := range fn.Body.Statements {
			err := s.Execute(mem)
			if err != nil {
				if err.Type == errors.Return {
					exp := fn.Returns[err.MessageId]
					returnResult, returnErr = exp.evalBool(mem)
				}
				break
			}
		}
		mem.Shrink()
		return returnResult, returnErr
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

func (s *Assignment) Execute(mem *Memory) *errors.Error {
	switch s.DataType {
	case Int:
		switch s.Operator.Type {
		case tokens.DOUBLE_PLUS:
			mem.IncInt(s.Var, 1, s.VarScope)
			return nil
		case tokens.DOUBLE_MINUS:
			mem.DecInt(s.Var, 1, s.VarScope)
			return nil
		}

		value, err := s.Exp.evalInt(mem)
		if err != nil {
			return err
		}
		switch s.Operator.Type {
		case tokens.EQUAL:
			if s.Explicit {
				mem.SetInt(s.Var, value)
			} else {
				mem.UpdateInt(s.Var, value, s.VarScope)
			}
		case tokens.PLUS_EQUAL:
			mem.IncInt(s.Var, value, s.VarScope)
		case tokens.MINUS_EQUAL:
			mem.DecInt(s.Var, value, s.VarScope)
		case tokens.STAR_EQUAL:
			mem.MulInt(s.Var, value, s.VarScope)
		case tokens.SLASH_EQUAL:
			if value == 0 {
				return &errors.Error{Message: "zero division not allowed", Token: s.Operator, Type: errors.RuntimeError}
			}
			mem.DivInt(s.Var, value, s.VarScope)
		case tokens.MODULO_EQUAL:
			if value == 0 {
				return &errors.Error{Message: "zero division not allowed", Token: s.Operator, Type: errors.RuntimeError}
			}
			mem.ModInt(s.Var, value, s.VarScope)
		case tokens.COLON_EQUAL:
			mem.SetInt(s.Var, value)
		}
	case Bool:
		value, err := s.Exp.evalBool(mem)
		if err != nil {
			return err
		}
		switch s.Operator.Type {
		case tokens.EQUAL:
			if s.Explicit {
				mem.SetBool(s.Var, value)
			} else {
				mem.UpdateBool(s.Var, value, s.VarScope)
			}
		case tokens.COLON_EQUAL:
			mem.SetBool(s.Var, value)
		}
	}
	return nil
}

func (s *Conditional) Execute(mem *Memory) *errors.Error {
	switch s.Token.Type {
	case tokens.IF:
		condition, err := s.Condition.evalBool(mem)
		if err != nil {
			return err
		}
		if condition {
            mem.Extend()
			for _, stmt := range s.Then.Statements {
				err := stmt.Execute(mem)
				if err != nil {
					return err
				}
			}
            mem.Shrink()
		} else if s.Else != nil {
            mem.Extend()
			for _, stmt := range s.Else.Statements {
				err := stmt.Execute(mem)
				if err != nil {
					return err
				}
			}
            mem.Shrink()
		}
	default:
		then, _ := s.Condition.evalBool(mem)
	WhileLoop:
		for {
			condition, err := s.Condition.evalBool(mem)
			if err != nil {
				return err
			}
			if condition {
				for _, stmt := range s.Then.Statements {
                    mem.Extend()
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
                    mem.Shrink()
				}
			} else {
				break
			}

		}
		if !then && s.Else != nil {
            mem.Extend()
			for _, stmt := range s.Else.Statements {
				stmt.Execute(mem)
			}
            mem.Shrink()
		}
	}
	return nil
}

func (s *For) Execute(mem *Memory) *errors.Error {
	mem.Extend()
	err := s.Init.Execute(mem)
	if err != nil {
		return err
	}
	mem.Extend()
	if s.Block == nil {
		for {
			condition, err := s.Condition.evalBool(mem)
			if err != nil {
				return err
			}
			if condition {
				err := s.After.Execute(mem)
				if err != nil {
					return err
				}
			} else {
				break
			}
		}
		mem.Shrink()
		return nil
	}
ForLoop:
	for {
		condition, err := s.Condition.evalBool(mem)
		if err != nil {
			return err
		}
		if condition {
			for _, stmt := range s.Block.Statements {
				err := stmt.Execute(mem)
				if err != nil {
					switch err.Type {
					case errors.Break:
						break ForLoop
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
			err := s.After.Execute(mem)
			if err != nil {
				return err
			}
			continue
		}
		break
	}
	mem.Shrink()
	mem.Shrink()
	return nil
}

func (s *Def) Execute(mem *Memory) *errors.Error {
	fun := Function{
		Scope:    s.Scope,
		Params:   s.Params,
		DataType: s.DataType,
		Body:     s.Body,
		Returns:  s.ReturnBranches,
	}
	mem.SetFunc(s.NameToken, &fun)
	return nil
}

func (s *Return) Execute(mem *Memory) *errors.Error {
	return &errors.Error{Type: errors.Return, MessageId: s.Id}
}

func (s *Break) Execute(mem *Memory) *errors.Error {
	return &errors.Error{Type: errors.Break}
}

func (s *Continue) Execute(mem *Memory) *errors.Error {
	return &errors.Error{Type: errors.Continue}
}

func (s *VoidCall) Execute(mem *Memory) *errors.Error {
	fn, err := mem.GetFunc(s.NameToken, s.Scope)
	if err != nil {
		return err
	}
	if fn.Body == nil {
		return nil
	}
	mem.Extend()
	for i, a := range s.Args {
		switch a.DataType {
		case Int:
			val, err := a.evalInt(mem)
			if err != nil {
				return err
			}
			mem.SetInt(fn.Params[i].NameToken, val)
		case Bool:
			val, err := a.evalBool(mem)
			if err != nil {
				return err
			}
			mem.SetBool(fn.Params[i].NameToken, val)
		default:
			return &errors.Error{Message: "invalid argument type", Type: errors.RuntimeError, Token: s.NameToken}
		}
	}
	for _, s := range fn.Body.Statements {
		err := s.Execute(mem)
		if err != nil {
			if err.Type == errors.Return {
				break
			}
			return err
		}
	}
	mem.Shrink()
	return nil
}

func (s *OpenScope) Execute(mem *Memory) *errors.Error {
	mem.Extend()
	return nil
}

func (s *CloseScope) Execute(mem *Memory) *errors.Error {
	mem.Shrink()
	return nil
}
