package parser

import (
	"fmt"
	"simpl/ast"
	"simpl/errors"
	sTokens "simpl/tokens"
)

var permittedInfixes map[sTokens.TokenType]bool = map[sTokens.TokenType]bool{
	sTokens.PLUS:   true,
	sTokens.MINUS:  true,
	sTokens.STAR:   true,
	sTokens.SLASH:  true,
	sTokens.MODULO: true,

	sTokens.OR:           true,
	sTokens.AND:          true,
	sTokens.DOUBLE_EQUAL: true,
	sTokens.NOT_EQUAL:    true,

	sTokens.GREATER:       true,
	sTokens.GREATER_EQUAL: true,
	sTokens.LESS:          true,
	sTokens.LESS_EQUAL:    true,
}

type Cache struct {
	size int
	vars []map[string]ast.DataType
}

func (c *Cache) Resize(scope int) {
	for c.size <= scope {
		c.vars = append(c.vars, map[string]ast.DataType{})
		c.size++
	}
	if c.size > scope+1 {
		c.vars = c.vars[:scope+1]
		c.size = scope + 1
	}
}

func (c *Cache) GetVarType(name string) (ast.DataType, bool) {
	for i := c.size - 1; i >= 0; i-- {
		val, found := c.vars[i][name]
		if found {
			return val, true
		}
	}
	return ast.Invalid, false
}

func (c *Cache) SetVarType(name string, dataType ast.DataType) {
	c.vars[c.size-1][name] = dataType
}

type ParseSource struct {
	Errors  []*errors.Error
	cache   *Cache
	tokens  []sTokens.Token
	scope   int
	current int
}

func New(tokens []sTokens.Token) ParseSource {
	return ParseSource{
		cache:  &Cache{vars: []map[string]ast.DataType{}},
		tokens: tokens,
	}
}

func (s *ParseSource) Parse(inLoop bool) (*ast.Program, *errors.Error) {
	statements := []ast.Statement{}

	sourceSize := len(s.tokens) - 1
	scopeStarts := []sTokens.Token{}

MainLoop:
	for s.current < sourceSize {
		token := s.tokens[s.current]
		switch token.Type {
		case sTokens.LEFT_BRACE:
			s.current++
			s.scope++
			block, err := s.Parse(inLoop)
			s.scope--
			if err != nil {
				return nil, err
			}
			statements = append(statements, block.Statements...)
		case sTokens.RIGHT_BRACE:
			if s.scope == 0 {
				return nil, s.unexpectedError(token)
			}
			s.current++
			break MainLoop
		case sTokens.BOOL_TYPE, sTokens.INT_TYPE, sTokens.IDENTIFIER:
			stmt, err := s.parseAssignment(sTokens.SEMICOLON)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case sTokens.IF, sTokens.WHILE:
			var stmt ast.Conditional
			s.current++
			condition, err := s.parseExpression(sTokens.Precedences[sTokens.EOF], sTokens.LEFT_BRACE)
			if err != nil {
				return nil, err
			}
			s.scope++
			s.current += 2
			thenBlock, err := s.Parse(inLoop || token.Type == sTokens.WHILE)
			if err != nil {
				return nil, err
			}
			stmt.Token = token
			stmt.Condition = condition
			stmt.Then = thenBlock
			stmt.Scope = s.scope

			token = s.tokens[s.current]
			if token.Type == sTokens.ELSE {
				s.current += 2
				elseBlock, err := s.Parse(inLoop || token.Type == sTokens.WHILE)
				if err != nil {
					return nil, err
				}
				stmt.Else = elseBlock
			}
			s.scope--
			statements = append(statements, &stmt)
		case sTokens.FOR:
			s.current++
			s.scope++
			init, err := s.parseAssignment(sTokens.SEMICOLON)
			if err != nil {
				return nil, err
			}
			condition, err := s.parseExpression(sTokens.Precedences[sTokens.EOF], sTokens.SEMICOLON)
			if err != nil {
				return nil, err
			}
			s.current += 2
			after, err := s.parseAssignment(sTokens.LEFT_BRACE)
			if err != nil {
				return nil, err
			}
			s.scope++
			block, err := s.Parse(true)
			if err != nil {
				return nil, err
			}
			s.scope--
			statements = append(statements, &ast.For{Scope: s.scope, Init: init, Condition: condition, After: after, Block: block, Token: token})
			s.scope--
		case sTokens.BREAK:
			if s.tokens[s.current+1].Type != sTokens.SEMICOLON {
				return nil, s.unexpectedError(s.tokens[s.current+1])
			}
			if !inLoop {
				s.Errors = append(s.Errors, &errors.Error{Message: "break outside of loop body", Type: errors.SyntaxError, Token: token})
			}
			s.current += 2
			statements = append(statements, &ast.Break{})
		case sTokens.CONTINUE:
			if s.tokens[s.current+1].Type != sTokens.SEMICOLON {
				return nil, s.unexpectedError(s.tokens[s.current+1])
			}
			if !inLoop {
				s.Errors = append(s.Errors, &errors.Error{Message: "continue outside of loop body", Type: errors.SyntaxError, Token: token})
			}
			s.current += 2
			statements = append(statements, &ast.Continue{})
		case sTokens.EOF:
			if s.scope != 0 {
				brace := scopeStarts[0]
				return nil, &errors.Error{Message: "Scope not closed", Token: brace, Type: errors.SyntaxError}
			}
			break MainLoop
		default:
			return nil, s.unexpectedError(token)
		}
	}

	return &ast.Program{Statements: statements}, nil
}

func (s *ParseSource) parseAssignment(endToken sTokens.TokenType) (*ast.Assignment, *errors.Error) {
	s.cache.Resize(s.scope)
	var stmt ast.Assignment
	token := s.tokens[s.current]
	if token.Type == sTokens.IDENTIFIER {
		operator := s.tokens[s.current+1]
		switch operator.Type {
		case sTokens.COLON_EQUAL, sTokens.EQUAL, sTokens.PLUS_EQUAL, sTokens.MINUS_EQUAL, sTokens.STAR_EQUAL, sTokens.SLASH_EQUAL, sTokens.MODULO_EQUAL:
			s.current += 2
			exp, err := s.parseExpression(sTokens.Precedences[sTokens.EOF], endToken)
			if err != nil {
				return nil, err
			}
			stmt.Var = token
			stmt.Operator = operator
			stmt.Exp = exp
			stmt.Scope = s.scope

			switch operator.Type {
			case sTokens.COLON_EQUAL:
				_, defined := s.cache.vars[s.cache.size-1][token.Value]
				if defined {
					s.Errors = append(s.Errors, &errors.Error{Message: "variable reassignment not allowed", Type: errors.ReferenceError, Token: token})
				} else {
					s.cache.SetVarType(token.Value, exp.DataType)
					stmt.DataType = exp.DataType
				}
			default:
				dataType, defined := s.cache.GetVarType(token.Value)
				if !defined {
					s.Errors = append(s.Errors, &errors.Error{Message: "undefined variable", Type: errors.ReferenceError, Token: token})
				} else if dataType != exp.DataType && exp.DataType != ast.Invalid {
					s.Errors = append(s.Errors, &errors.Error{Message: "assigning wrong type", Type: errors.TypeError, Token: token})
				} else {
					stmt.DataType = exp.DataType
				}
			}
			s.current += 2
		case sTokens.DOUBLE_PLUS, sTokens.DOUBLE_MINUS:
			if s.tokens[s.current+2].Type != endToken {
				return nil, s.unexpectedError(s.tokens[s.current+2])
			}
			stmt.Var = token
			stmt.Operator = operator
			stmt.Scope = s.scope
			dataType, defined := s.cache.GetVarType(token.Value)
			if !defined {
				s.Errors = append(s.Errors, &errors.Error{Message: "undefined variable", Type: errors.ReferenceError, Token: token})
			} else if dataType != ast.Int {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: operator})
			}
			stmt.DataType = ast.Int
			s.current += 3
		default:
			return nil, s.unexpectedError(operator)
		}
		return &stmt, nil
	}

	stmt.Explicit = true
	switch token.Type {
	case sTokens.INT_TYPE:
		stmt.DataType = ast.Int
	default:
		stmt.DataType = ast.Bool
	}
	varToken := s.tokens[s.current+1]
	if varToken.Type != sTokens.IDENTIFIER {
		return nil, s.unexpectedError(varToken)
	}
	stmt.Var = varToken
	operator := s.tokens[s.current+2]
	if operator.Type != sTokens.EQUAL {
		return nil, s.unexpectedError(operator)
	}
	stmt.Operator = operator
	s.current += 3
	exp, err := s.parseExpression(sTokens.Precedences[sTokens.EOF], endToken)
	s.current += 2
	if err != nil {
		return nil, err
	}
	stmt.Exp = exp
	if exp.DataType != stmt.DataType && exp.DataType != ast.Invalid {
		s.Errors = append(s.Errors, &errors.Error{Message: "assigning wrong type", Type: errors.TypeError, Token: operator})
	}
	return &stmt, nil
}

func (s *ParseSource) parseExpression(precedence int, endToken sTokens.TokenType) (*ast.Expression, *errors.Error) {
	left, err := s.parsePrefix()
	if err != nil {
		return nil, err
	}
	for {
		if s.tokens[s.current+1].Type == sTokens.EOF {
			return nil, &errors.Error{Message: "expected ;", Token: s.tokens[s.current], Type: errors.SyntaxError}
		}
		if s.tokens[s.current+1].Type == endToken || precedence >= sTokens.Precedences[s.tokens[s.current+1].Type] {
			break
		}
		s.current++
		token := s.tokens[s.current]
		if !permittedInfixes[token.Type] {
			return nil, &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
		}
		nextLeft := &ast.Expression{Token: token}
		prec := sTokens.Precedences[token.Type]
		s.current++
		right, err := s.parseExpression(prec, endToken)
		if err != nil {
			return nil, err
		}
		switch token.Type {
		case sTokens.STAR, sTokens.SLASH, sTokens.PLUS, sTokens.MINUS, sTokens.MODULO:
			nextLeft.DataType = ast.Int
			if left.DataType != ast.Int && left.DataType != ast.Invalid || right.DataType != ast.Int && right.DataType != ast.Invalid {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: token})
			}
		case sTokens.LESS, sTokens.GREATER, sTokens.LESS_EQUAL, sTokens.GREATER_EQUAL:
			nextLeft.DataType = ast.Bool
			if left.DataType != ast.Int && left.DataType != ast.Invalid || right.DataType != ast.Int && right.DataType != ast.Invalid {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: token})
			}

		case sTokens.AND, sTokens.OR:
			nextLeft.DataType = ast.Bool
			if left.DataType != ast.Bool && left.DataType != ast.Invalid || right.DataType != ast.Bool && right.DataType != ast.Invalid {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: token})
			}
		case sTokens.NOT_EQUAL, sTokens.DOUBLE_EQUAL:
			nextLeft.DataType = ast.Bool
			if left.DataType != right.DataType && left.DataType != ast.Invalid && right.DataType != ast.Invalid {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: token})
			}
		}
		nextLeft.Left = left
		nextLeft.Right = right
		left = nextLeft
	}

	return left, nil
}

func (s *ParseSource) parsePrefix() (*ast.Expression, *errors.Error) {
	token := s.tokens[s.current]
	switch token.Type {
	case sTokens.BANG:
		node := &ast.Expression{Token: token, DataType: ast.Bool}
		s.current++
		expression, err := s.parsePrefix()
		if err != nil {
			return nil, err
		}
		node.Left = expression
		return node, nil
	case sTokens.NUMBER:
		return &ast.Expression{Token: token, DataType: ast.Int}, nil
	case sTokens.TRUE, sTokens.FALSE:
		return &ast.Expression{Token: token, DataType: ast.Bool}, nil
	case sTokens.IDENTIFIER:
		dataType, defined := s.cache.GetVarType(token.Value)
		if !defined {
			s.Errors = append(s.Errors, &errors.Error{Message: "undefined variable", Type: errors.ReferenceError, Token: token})
		}
		return &ast.Expression{Token: token, DataType: dataType}, nil
	case sTokens.LEFT_PAREN:
		return s.parseParens()
	default:
		return nil, &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
	}
}

func (s *ParseSource) parseParens() (*ast.Expression, *errors.Error) {
	s.current++
	node, err := s.parseExpression(0, sTokens.RIGHT_PAREN)
	if err != nil {
		return nil, err
	}
	nextToken := s.tokens[s.current+1]
	if s.tokens[s.current+1].Type != sTokens.RIGHT_PAREN {
		expected, got := sTokens.Representations[sTokens.RIGHT_PAREN], sTokens.Representations[nextToken.Type]
		return nil, &errors.Error{Message: fmt.Sprintf("expected %s, got %s", expected, got), Token: nextToken, Type: errors.SyntaxError}
	}
	s.current++
	return node, nil
}

func (s *ParseSource) unexpectedError(token sTokens.Token) *errors.Error {
	return &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
}
