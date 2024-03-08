package parser

import (
	"fmt"
	"simpl/errors"
	"simpl/intpr"
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
	size  int
	vars  []map[string]intpr.DataType
	funcs []map[string]FuncCache
}

type FuncCache struct {
	DataType intpr.DataType
	Params   []intpr.DefParam
	Returns  bool
}

func (c *Cache) Resize(scope int) {
	for c.size <= scope {
		c.vars = append(c.vars, map[string]intpr.DataType{})
		c.funcs = append(c.funcs, map[string]FuncCache{})
		c.size++
	}
	if c.size > scope+1 {
		c.vars = c.vars[:scope+1]
		c.funcs = c.funcs[:scope+1]
		c.size = scope + 1
	}
}

func (c *Cache) GetVarType(name string) (intpr.DataType, bool) {
	for i := c.size - 1; i >= 0; i-- {
		val, found := c.vars[i][name]
		if found {
			return val, true
		}
	}
	return intpr.Invalid, false
}

func (c *Cache) SetVarType(name string, dataType intpr.DataType) {
	c.vars[c.size-1][name] = dataType
}

func (c *Cache) SetFuncCache(name string, cache FuncCache) {
	c.funcs[c.size-1][name] = cache
}

type ParseSource struct {
	Errors          []*errors.Error
	cache           *Cache
	tokens          []sTokens.Token
	scope           int
	current         int
	currentFunction *FuncCache
}

func New(tokens []sTokens.Token) ParseSource {
	return ParseSource{
		cache:  &Cache{vars: []map[string]intpr.DataType{}, funcs: []map[string]FuncCache{}},
		tokens: tokens,
	}
}

func (s *ParseSource) Parse(inLoop bool) (*intpr.Program, *errors.Error) {
	s.cache.Resize(s.scope)
	statements := []intpr.Statement{}

	sourceSize := len(s.tokens) - 1
	scopeStarts := []sTokens.Token{}

MainLoop:
	for s.current < sourceSize {
		token := s.tokens[s.current]
		switch token.Type {
		case sTokens.LEFT_BRACE:
			s.current++
			s.scope++
			s.cache.Resize(s.scope)
			block, err := s.Parse(inLoop)
			s.scope--
			if err != nil {
				return nil, err
			}
			s.cache.Resize(s.scope)
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
			var stmt intpr.Conditional
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
			s.cache.Resize(s.scope)
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
			s.cache.Resize(s.scope)
			block, err := s.Parse(true)
			if err != nil {
				return nil, err
			}
			s.scope--
			s.cache.Resize(s.scope)
			statements = append(statements, &intpr.For{Scope: s.scope, Init: init, Condition: condition, After: after, Block: block, Token: token})
			s.scope--
			s.cache.Resize(s.scope)
		case sTokens.BREAK:
			if s.tokens[s.current+1].Type != sTokens.SEMICOLON {
				return nil, s.unexpectedError(s.tokens[s.current+1])
			}
			if !inLoop {
				s.Errors = append(s.Errors, &errors.Error{Message: "break outside of loop body", Type: errors.SyntaxError, Token: token})
			}
			s.current += 2
			statements = append(statements, &intpr.Break{})
		case sTokens.CONTINUE:
			if s.tokens[s.current+1].Type != sTokens.SEMICOLON {
				return nil, s.unexpectedError(s.tokens[s.current+1])
			}
			if !inLoop {
				s.Errors = append(s.Errors, &errors.Error{Message: "continue outside of loop body", Type: errors.SyntaxError, Token: token})
			}
			s.current += 2
			statements = append(statements, &intpr.Continue{})
		case sTokens.DEF:
			stmt := intpr.Def{Scope: s.scope}
			name := s.tokens[s.current+1]
			if name.Type != sTokens.IDENTIFIER {
				return nil, s.unexpectedError(name)
			}
			stmt.NameToken = name
			openParen := s.tokens[s.current+2]
			if openParen.Type != sTokens.LEFT_PAREN {
				return nil, s.unexpectedError(openParen)
			}
			s.current += 3
			stmt.Params = []intpr.DefParam{}
			if s.tokens[s.current].Type != sTokens.RIGHT_PAREN {
			ParamsLoop:
				for {
					param := intpr.DefParam{}
					paramType := s.tokens[s.current]
					switch paramType.Type {
					case sTokens.INT_TYPE:
						param.DataType = intpr.Int
					case sTokens.BOOL_TYPE:
						param.DataType = intpr.Bool
					default:
						return nil, s.unexpectedError(paramType)
					}
					s.current++
					paramName := s.tokens[s.current]
					if paramName.Type != sTokens.IDENTIFIER {
						return nil, s.unexpectedError(paramName)
					}
					param.NameToken = paramName
					s.current++
					delimitter := s.tokens[s.current]
					stmt.Params = append(stmt.Params, param)
					switch delimitter.Type {
					case sTokens.COMMA:
						s.current++
					case sTokens.RIGHT_PAREN:
						break ParamsLoop
					default:
						return nil, s.unexpectedError(delimitter)
					}
				}

			}
			s.current++
			nextToken := s.tokens[s.current]
			switch nextToken.Type {
			case sTokens.INT_TYPE:
				stmt.DataType = intpr.Int
				s.current++
			case sTokens.BOOL_TYPE:
				stmt.DataType = intpr.Bool
				s.current++
			case sTokens.LEFT_BRACE:
				stmt.DataType = intpr.Void
			default:
				return nil, s.unexpectedError(nextToken)
			}
			funcCache := FuncCache{
				DataType: stmt.DataType,
				Params:   stmt.Params,
			}
			_, defined := s.cache.vars[s.cache.size-1][stmt.NameToken.Value]
			if defined {
				s.Errors = append(s.Errors, &errors.Error{Message: "variable reassignment not allowed", Type: errors.ReferenceError, Token: stmt.NameToken})
			} else {
				s.cache.SetVarType(stmt.NameToken.Value, intpr.Func)
				s.cache.SetFuncCache(stmt.NameToken.Value, funcCache)
			}
			if s.tokens[s.current].Type != sTokens.LEFT_BRACE {
				return nil, s.unexpectedError(s.tokens[s.current])
			}
			s.current++
			s.scope++
			s.cache.Resize(s.scope)
			s.currentFunction = &funcCache
			body, err := s.Parse(false)
			if err != nil {
				return nil, err
			}
			s.scope--
			s.cache.Resize(s.scope)
			if stmt.DataType != intpr.Void && !s.currentFunction.Returns {
				s.Errors = append(s.Errors, &errors.Error{Message: "missing return", Type: errors.TypeError, Token: token})
			}
			s.currentFunction = nil
			stmt.Body = body
			statements = append(statements, &stmt)
		case sTokens.RETURN:
			stmt := intpr.Return{Scope: s.scope}
			s.current++
			nextToken := s.tokens[s.current]
			switch nextToken.Type {
			case sTokens.SEMICOLON:
				stmt.DataType = intpr.Void
				if s.currentFunction != nil && s.currentFunction.DataType != intpr.Void {
					s.Errors = append(s.Errors, &errors.Error{Message: fmt.Sprintf("wrong return type: expected %s", s.currentFunction.DataType.View()), Type: errors.TypeError, Token: nextToken})
				}
				s.current++
			default:
				exp, err := s.parseExpression(sTokens.Precedences[sTokens.EOF], sTokens.SEMICOLON)
				if err != nil {
					return nil, err
				}
				s.current += 2
				if s.currentFunction != nil && exp.DataType != s.currentFunction.DataType {
					s.Errors = append(s.Errors, &errors.Error{Message: fmt.Sprintf("wrong return type: expected %s", s.currentFunction.DataType.View()), Type: errors.TypeError, Token: nextToken})
				}
				stmt.DataType = exp.DataType
				stmt.Exp = exp
			}
			statements = append(statements, &stmt)
			if s.currentFunction == nil {
				s.Errors = append(s.Errors, &errors.Error{Message: "return outside of function body", Type: errors.SyntaxError, Token: token})
			} else {
				s.currentFunction.Returns = true
			}
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

	return &intpr.Program{Statements: statements}, nil
}

func (s *ParseSource) parseAssignment(endToken sTokens.TokenType) (*intpr.Assignment, *errors.Error) {
	var stmt intpr.Assignment
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
				if s.currentFunction != nil {
					defined := false
					for _, p := range s.currentFunction.Params {
						if p.NameToken.Value == token.Value {
							s.Errors = append(s.Errors, &errors.Error{Message: "variable reassignment not allowed", Type: errors.ReferenceError, Token: token})
							defined = true
						}
					}
					if !defined {
						s.cache.SetVarType(token.Value, exp.DataType)
						stmt.DataType = exp.DataType
					}
				} else {
					_, defined := s.cache.vars[s.cache.size-1][token.Value]
					if defined {
						s.Errors = append(s.Errors, &errors.Error{Message: "variable reassignment not allowed", Type: errors.ReferenceError, Token: token})
					} else {
						s.cache.SetVarType(token.Value, exp.DataType)
						stmt.DataType = exp.DataType
					}
				}
			default:
				var defined bool
				var dataType intpr.DataType
				if s.currentFunction != nil {
					for _, a := range s.currentFunction.Params {
						if a.NameToken.Value == token.Value {
							defined = true
							dataType = a.DataType
							break
						}
					}
				}
				if !defined {
					dataType, defined = s.cache.GetVarType(token.Value)
				}
				if !defined {
					s.Errors = append(s.Errors, &errors.Error{Message: "undefined variable", Type: errors.ReferenceError, Token: token})
				} else if dataType != exp.DataType && exp.DataType != intpr.Invalid {
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
			} else if dataType != intpr.Int {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: operator})
			}
			stmt.DataType = intpr.Int
			s.current += 3
		default:
			return nil, s.unexpectedError(operator)
		}
		return &stmt, nil
	}

	stmt.Explicit = true
	switch token.Type {
	case sTokens.INT_TYPE:
		stmt.DataType = intpr.Int
	default:
		stmt.DataType = intpr.Bool
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
	if exp.DataType != stmt.DataType && exp.DataType != intpr.Invalid {
		s.Errors = append(s.Errors, &errors.Error{Message: "assigning wrong type", Type: errors.TypeError, Token: operator})
	}
	return &stmt, nil
}

func (s *ParseSource) parseExpression(precedence int, endToken sTokens.TokenType) (*intpr.Expression, *errors.Error) {
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
		nextLeft := &intpr.Expression{Token: token}
		prec := sTokens.Precedences[token.Type]
		s.current++
		right, err := s.parseExpression(prec, endToken)
		if err != nil {
			return nil, err
		}
		switch token.Type {
		case sTokens.STAR, sTokens.SLASH, sTokens.PLUS, sTokens.MINUS, sTokens.MODULO:
			nextLeft.DataType = intpr.Int
			if left.DataType != intpr.Int && left.DataType != intpr.Invalid || right.DataType != intpr.Int && right.DataType != intpr.Invalid {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: token})
			}
		case sTokens.LESS, sTokens.GREATER, sTokens.LESS_EQUAL, sTokens.GREATER_EQUAL:
			nextLeft.DataType = intpr.Bool
			if left.DataType != intpr.Int && left.DataType != intpr.Invalid || right.DataType != intpr.Int && right.DataType != intpr.Invalid {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: token})
			}

		case sTokens.AND, sTokens.OR:
			nextLeft.DataType = intpr.Bool
			if left.DataType != intpr.Bool && left.DataType != intpr.Invalid || right.DataType != intpr.Bool && right.DataType != intpr.Invalid {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: token})
			}
		case sTokens.NOT_EQUAL, sTokens.DOUBLE_EQUAL:
			nextLeft.DataType = intpr.Bool
			if left.DataType != right.DataType && left.DataType != intpr.Invalid && right.DataType != intpr.Invalid || left.DataType == intpr.Func || right.DataType == intpr.Func {
				s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: token})
			}
		}
		nextLeft.Left = left
		nextLeft.Right = right
		left = nextLeft
	}

	return left, nil
}

func (s *ParseSource) parsePrefix() (*intpr.Expression, *errors.Error) {
	token := s.tokens[s.current]
	switch token.Type {
	case sTokens.BANG:
		node := &intpr.Expression{Token: token, DataType: intpr.Bool}
		s.current++
		expression, err := s.parsePrefix()
		if err != nil {
			return nil, err
		}
		node.Left = expression
		return node, nil
	case sTokens.NUMBER:
		return &intpr.Expression{Token: token, DataType: intpr.Int}, nil
	case sTokens.TRUE, sTokens.FALSE:
		return &intpr.Expression{Token: token, DataType: intpr.Bool}, nil
	case sTokens.IDENTIFIER:
		var dataType intpr.DataType
		var defined bool
		if s.currentFunction != nil {
			for _, a := range s.currentFunction.Params {
				if a.NameToken.Value == token.Value {
					defined = true
					dataType = a.DataType
					break
				}
			}
		}
		if !defined {
			dataType, defined = s.cache.GetVarType(token.Value)
		}
		if !defined {
			s.Errors = append(s.Errors, &errors.Error{Message: "undefined variable", Type: errors.ReferenceError, Token: token})
		}
		return &intpr.Expression{Token: token, DataType: dataType}, nil
	case sTokens.LEFT_PAREN:
		return s.parseParens()
	default:
		return nil, &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
	}
}

func (s *ParseSource) parseParens() (*intpr.Expression, *errors.Error) {
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
