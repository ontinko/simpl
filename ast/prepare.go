package ast

import (
	"simpl/errors"
	"simpl/tokens"
)

func resizeCache(cache *[]map[string]DataType, scope int) {
	for len(*cache) <= scope {
		*cache = append(*cache, map[string]DataType{})
	}
	if len(*cache) > scope+1 {
		*cache = (*cache)[:scope+1]
	}
}

func (e *Expression) Prepare(cache *[]map[string]DataType) []*errors.Error {
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
		for i := len(*cache) - 1; i >= 0; i-- {
			value, ok := (*cache)[i][e.Token.Value]
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
	case tokens.STAR, tokens.SLASH, tokens.PLUS, tokens.MINUS, tokens.MODULO:
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
	case tokens.LESS, tokens.GREATER, tokens.LESS_EQUAL, tokens.GREATER_EQUAL:
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

func (s *Assignment) Prepare(cache *[]map[string]DataType) []*errors.Error {
	resizeCache(cache, s.Scope)
	errs := []*errors.Error{}
	expErrs := s.Exp.Prepare(cache)
	if s.Operator.Type == tokens.COLON_EQUAL {
		_, defined := (*cache)[len(*cache)-1][s.Var.Value]
		if defined {
			errs = append(errs, &errors.Error{Message: "variable reassignment not allowed", Token: s.Var, Type: errors.SyntaxError})
		} else {
			(*cache)[len(*cache)-1][s.Var.Value] = s.Exp.DataType
			s.DataType = s.Exp.DataType
		}
		errs = append(errs, expErrs...)
		return errs
	}
	defined := false
	var dataType DataType
	for i := len(*cache) - 1; i >= 0; i-- {
		value, ok := (*cache)[i][s.Var.Value]
		if ok {
			defined = true
			dataType = value
			break
		}
	}
	if !defined {
		errs = append(errs, &errors.Error{Message: "undefined variable", Token: s.Var, Type: errors.ReferenceError})
		return errs
	}

	switch s.Operator.Type {
	case tokens.DOUBLE_MINUS, tokens.DOUBLE_PLUS:
		s.DataType = Int
		if dataType != Int && dataType != Invalid {
			errs = append(errs, &errors.Error{Message: "int/dec on wrong type", Token: s.Var, Type: errors.TypeError})
		}
	case tokens.MODULO:
		s.DataType = Int
		if dataType != Int && dataType != Invalid || s.Exp.DataType != Int && s.Exp.DataType != Invalid {
			errs = append(errs, &errors.Error{Message: "assigning wrong type", Token: s.Var, Type: errors.TypeError})
		}
	default:
		if dataType != s.Exp.DataType && s.Exp.DataType != Invalid {
			errs = append(errs, &errors.Error{Message: "assigning wrong type", Token: s.Var, Type: errors.TypeError})
		} else {
			s.DataType = s.Exp.DataType
		}
	}

	return errs
}

func (s *Conditional) Prepare(cache *[]map[string]DataType) []*errors.Error {
	resizeCache(cache, s.Scope)
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

func (p *Program) Prepare(cache *[]map[string]DataType) []*errors.Error {
	if p == nil {
		return nil
	}

	errs := []*errors.Error{}
	for _, s := range p.Statements {
		errs = append(errs, s.Prepare(cache)...)
	}

	return errs
}

func (s *For) Prepare(cache *[]map[string]DataType) []*errors.Error {
	for len(*cache) <= s.Scope {
		*cache = append(*cache, map[string]DataType{})
	}
	if len(*cache) > s.Scope+1 {
		*cache = (*cache)[:s.Scope+1]
	}
	errs := []*errors.Error{}
	errs = append(errs, s.Init.Prepare(cache)...)
	errs = append(errs, s.Condition.Prepare(cache)...)
	errs = append(errs, s.After.Prepare(cache)...)
	errs = append(errs, s.Block.Prepare(cache)...)

	return errs
}

func (s *Break) Prepare(cache *[]map[string]DataType) []*errors.Error {
	return nil
}

func (s *Continue) Prepare(cache *[]map[string]DataType) []*errors.Error {
	return nil
}
