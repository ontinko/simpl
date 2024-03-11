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

func NewCache() *Cache {
	return &Cache{vars: []map[string]intpr.DataType{{}}, funcs: []map[string]FuncCache{{}}, size: 1}
}

type FuncCache struct {
	DataType       intpr.DataType
	Params         []intpr.DefParam
	Returns        bool
	ReturnBranches []*intpr.Expression
	LocalScope     *Cache
}

func (c *Cache) Extend() {
	c.vars = append(c.vars, map[string]intpr.DataType{})
	c.funcs = append(c.funcs, map[string]FuncCache{})
	c.size++
}

func (c *Cache) Shrink() {
	c.size--
	c.vars = c.vars[:c.size]
	c.funcs = c.funcs[:c.size]
}

func (c *Cache) GetVarType(name string) (intpr.DataType, int, bool) {
	for i := c.size - 1; i >= 0; i-- {
		val, found := c.vars[i][name]
		if found {
			return val, i, true
		}
	}
	return intpr.Invalid, -1, false
}

func (c *Cache) GetFuncCache(name string) *FuncCache {
	for i := c.size - 1; i >= 0; i-- {
		cache, found := c.funcs[i][name]
		if found {
			return &cache
		}
	}
	return nil
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

type FunctionCall struct {
	Identifier sTokens.Token
	Args       []*intpr.Expression
}

func New(tokens []sTokens.Token) ParseSource {
	return ParseSource{
		cache:  NewCache(),
		tokens: tokens,
	}
}

func (s *ParseSource) Parse(inLoop bool) (*intpr.Program, *errors.Error) {
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
			s.cache.Extend()
			block, err := s.Parse(inLoop)
			s.scope--
			if err != nil {
				return nil, err
			}
			s.cache.Shrink()
			statements = append(statements, &intpr.OpenScope{})
			statements = append(statements, block.Statements...)
			statements = append(statements, &intpr.CloseScope{})
		case sTokens.RIGHT_BRACE:
			if s.scope == 0 {
				return nil, s.unexpectedError(token)
			}
			s.current++
			break MainLoop
		case sTokens.BOOL_TYPE, sTokens.INT_TYPE, sTokens.IDENTIFIER:
			stmt, err := s.parseOneliner(sTokens.SEMICOLON)
			if err != nil {
				return nil, err
			}
			if s.tokens[s.current+1].Type != sTokens.SEMICOLON {
				return nil, s.unexpectedError(s.tokens[s.current+1])
			}
			s.current += 2
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
			s.cache.Extend()
			init, err := s.parseOneliner(sTokens.SEMICOLON)
			if err != nil {
				return nil, err
			}
			s.current += 2
			condition, err := s.parseExpression(sTokens.Precedences[sTokens.EOF], sTokens.SEMICOLON)
			if err != nil {
				return nil, err
			}
			s.current += 2
			after, err := s.parseOneliner(sTokens.LEFT_BRACE)
			if err != nil {
				return nil, err
			}
			s.current += 2
			s.scope++
			block, err := s.Parse(true)
			if err != nil {
				return nil, err
			}
			statements = append(statements, &intpr.For{Init: init, Condition: condition, After: after, Block: block, Token: token})
			s.scope -= 2
			s.cache.Shrink()
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
			if s.currentFunction != nil {
				return nil, s.unexpectedError(token)
			}
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
			fnCache := FuncCache{
				DataType:   stmt.DataType,
				Params:     stmt.Params,
				LocalScope: NewCache(),
			}
			_, defined := s.cache.vars[s.cache.size-1][stmt.NameToken.Value]
			if defined {
				s.Errors = append(s.Errors, &errors.Error{Message: "variable reassignment not allowed", Type: errors.ReferenceError, Token: stmt.NameToken})
			} else {
				s.cache.SetVarType(stmt.NameToken.Value, intpr.Func)
				s.cache.SetFuncCache(stmt.NameToken.Value, fnCache)
			}
			if s.tokens[s.current].Type != sTokens.LEFT_BRACE {
				return nil, s.unexpectedError(s.tokens[s.current])
			}
			s.currentFunction = &fnCache
			s.current++
			s.scope++
			s.cache.Extend()
			body, err := s.Parse(false)
			if err != nil {
				return nil, err
			}
			s.scope--
			if stmt.DataType != intpr.Void && !s.currentFunction.Returns {
				s.Errors = append(s.Errors, &errors.Error{Message: "missing return", Type: errors.TypeError, Token: token})
			}
			stmt.Body = body
			stmt.ReturnBranches = s.currentFunction.ReturnBranches
			s.currentFunction = nil
			statements = append(statements, &stmt)
		case sTokens.RETURN:
			stmt := intpr.Return{}
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
				stmt.Id = len(s.currentFunction.ReturnBranches)
				s.currentFunction.ReturnBranches = append(s.currentFunction.ReturnBranches, exp)
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

func (s *ParseSource) parseOneliner(endToken sTokens.TokenType) (intpr.Statement, *errors.Error) {
	token := s.tokens[s.current]
	scope := s.scope
	if s.currentFunction != nil {
		scope = -1
	}
	if token.Type == sTokens.IDENTIFIER && s.tokens[s.current+1].Type == sTokens.LEFT_PAREN {
		fnCall, err := s.parseFunctionCall()
		if err != nil {
			return nil, err
		}
		fnCache := s.cache.GetFuncCache(fnCall.Identifier.Value)
		if fnCache != nil && fnCache.DataType != intpr.Void {
			s.Errors = append(s.Errors, &errors.Error{Message: fmt.Sprintf("%s is not a void function", fnCall.Identifier.Value), Type: errors.ReferenceError, Token: fnCall.Identifier})
		}
		return &intpr.VoidCall{
			NameToken: fnCall.Identifier,
			Args:      fnCall.Args,
			Scope:     scope,
		}, nil
	}

	stmt := intpr.Assignment{}
	if token.Type != sTokens.IDENTIFIER {
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
		if err != nil {
			return nil, err
		}
		stmt.Exp = exp
		if exp.DataType != stmt.DataType && exp.DataType != intpr.Invalid {
			s.Errors = append(s.Errors, &errors.Error{Message: "assigning wrong type", Type: errors.TypeError, Token: operator})
		}
		return &stmt, nil
	}
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

		switch operator.Type {
		case sTokens.COLON_EQUAL:
			if s.currentFunction != nil {
				stmt.VarScope = -1
				defined := false
				_, _, defined = s.currentFunction.LocalScope.GetVarType(token.Value)
				if defined {
					s.Errors = append(s.Errors, &errors.Error{Message: "variable reassignment not allowed", Type: errors.ReferenceError, Token: token})
				} else {
					for _, p := range s.currentFunction.Params {
						if p.NameToken.Value == token.Value {
							s.Errors = append(s.Errors, &errors.Error{Message: "variable reassignment not allowed", Type: errors.ReferenceError, Token: token})
							defined = true
						}
					}
					if !defined {
						s.currentFunction.LocalScope.SetVarType(token.Value, exp.DataType)
						stmt.DataType = exp.DataType
					}
				}
			} else {
				stmt.VarScope = s.scope
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
			scope := s.scope
			if s.currentFunction != nil {
				scope = -1
				for _, a := range s.currentFunction.Params {
					if a.NameToken.Value == token.Value {
						defined = true
						dataType = a.DataType
						break
					}
				}
				if !defined {
					dataType, _, defined = s.currentFunction.LocalScope.GetVarType(token.Value)
				}
			}
			if !defined {
				dataType, scope, defined = s.cache.GetVarType(token.Value)
			}
			if !defined {
				s.Errors = append(s.Errors, &errors.Error{Message: "undefined variable", Type: errors.ReferenceError, Token: token})
			} else if dataType != exp.DataType && exp.DataType != intpr.Invalid {
				s.Errors = append(s.Errors, &errors.Error{Message: "assigning wrong type", Type: errors.TypeError, Token: token})
			} else {
				stmt.DataType = exp.DataType
				stmt.VarScope = scope
			}
		}
	case sTokens.DOUBLE_PLUS, sTokens.DOUBLE_MINUS:
		if s.tokens[s.current+2].Type != endToken {
			return nil, s.unexpectedError(s.tokens[s.current+2])
		}
		stmt.Var = token
		stmt.Operator = operator
		var dataType intpr.DataType
		var scope int
		var defined bool
		if s.currentFunction != nil {
			scope = -1
			dataType, _, defined = s.currentFunction.LocalScope.GetVarType(token.Value)
			if !defined {
				for _, p := range s.currentFunction.Params {
					if p.NameToken.Value == token.Value {
						dataType = p.DataType
						defined = true
						break
					}
				}
			}
		}
		if !defined {
			dataType, scope, defined = s.cache.GetVarType(token.Value)
		}
		if !defined {
			s.Errors = append(s.Errors, &errors.Error{Message: "undefined variable", Type: errors.ReferenceError, Token: token})
		} else if dataType != intpr.Int {
			s.Errors = append(s.Errors, &errors.Error{Message: "invalid operation", Type: errors.TypeError, Token: operator})
		}
		// if s.currentFunction != nil {
		// 	scope = -1
		// }
		stmt.DataType = intpr.Int
		stmt.VarScope = scope
		s.current++
	default:
		return nil, s.unexpectedError(operator)
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
		scope := s.scope
		if s.currentFunction != nil {
			scope = -1
		}
		nextLeft := &intpr.Expression{Token: token, Scope: scope}
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
		scope := s.scope
		if s.currentFunction != nil {
			scope = -1
		}
		node := &intpr.Expression{Token: token, DataType: intpr.Bool, Scope: scope}
		s.current++
		expression, err := s.parsePrefix()
		if err != nil {
			return nil, err
		}
		node.Left = expression
		return node, nil
	case sTokens.NUMBER:
		return &intpr.Expression{Token: token, DataType: intpr.Int, Scope: s.scope}, nil
	case sTokens.TRUE, sTokens.FALSE:
		return &intpr.Expression{Token: token, DataType: intpr.Bool, Scope: s.scope}, nil
	case sTokens.IDENTIFIER:
		if s.tokens[s.current+1].Type == sTokens.LEFT_PAREN {
			fnCall, err := s.parseFunctionCall()
			if err != nil {
				return nil, err
			}
			fnCache := s.cache.GetFuncCache(fnCall.Identifier.Value)
			if fnCache == nil {
				return &intpr.Expression{Token: token, DataType: intpr.Invalid, Args: fnCall.Args, Scope: s.scope}, nil
			}
			if fnCache.DataType == intpr.Void {
				s.Errors = append(s.Errors, &errors.Error{Message: fmt.Sprintf("function %s does not return a value", token.Value), Type: errors.TypeError, Token: token})
			}
			return &intpr.Expression{Token: token, DataType: fnCache.DataType, Args: fnCall.Args, Scope: s.scope}, nil
		}
		var dataType intpr.DataType
		var defined bool
		scope := s.scope
		if s.currentFunction != nil {
			for _, a := range s.currentFunction.Params {
				if a.NameToken.Value == token.Value {
					defined = true
					dataType = a.DataType
					scope = -1
					break
				}
			}
			if !defined {
				dataType, _, defined = s.currentFunction.LocalScope.GetVarType(token.Value)
			}
		}
		if !defined {
			dataType, scope, defined = s.cache.GetVarType(token.Value)
		} else {
			scope = -1
		}
		if !defined {
			s.Errors = append(s.Errors, &errors.Error{Message: "undefined variable", Type: errors.ReferenceError, Token: token})
		}
		return &intpr.Expression{Token: token, DataType: dataType, Scope: scope}, nil
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

func (s *ParseSource) parseFunctionCall() (*FunctionCall, *errors.Error) {
	identifier := s.tokens[s.current]
	s.current += 2
	args := []*intpr.Expression{}
	if s.tokens[s.current].Type != sTokens.RIGHT_PAREN {
		for {
			exp, err := s.parseExpression(sTokens.Precedences[sTokens.RIGHT_PAREN], sTokens.COMMA)
			if err != nil {
				return nil, err
			}
			args = append(args, exp)
			if s.tokens[s.current+1].Type == sTokens.COMMA {
				s.current += 2
				continue
			} else {
				s.current++
				break
			}
		}
	}
	dataType, _, defined := s.cache.GetVarType(identifier.Value)
	if !defined {
		s.Errors = append(s.Errors, &errors.Error{Message: "undefined function", Type: errors.ReferenceError, Token: identifier})
	} else if dataType != intpr.Func {
		s.Errors = append(s.Errors, &errors.Error{Message: fmt.Sprintf("%s is not a function", identifier.Value), Type: errors.ReferenceError, Token: identifier})
	} else {
		fnCache := s.cache.GetFuncCache(identifier.Value)
		if fnCache == nil {
			s.Errors = append(s.Errors, &errors.Error{Message: fmt.Sprintf("function %s is missing", identifier.Value), Type: errors.ReferenceError, Token: identifier})
		} else if len(fnCache.Params) != len(args) {
			s.Errors = append(s.Errors, &errors.Error{Message: fmt.Sprintf("wrong set of arguments for function %s", identifier.Value), Type: errors.ReferenceError, Token: identifier})
		} else {
			for i := 0; i < len(args); i++ {
				param := fnCache.Params[i]
				if param.DataType != args[i].DataType {
					s.Errors = append(s.Errors, &errors.Error{Message: fmt.Sprintf("wrong type for parameter %s for function %s", identifier.Value, identifier.Value), Type: errors.ReferenceError, Token: identifier})
				}
			}
		}
	}

	return &FunctionCall{Identifier: identifier, Args: args}, nil
}

func (s *ParseSource) unexpectedError(token sTokens.Token) *errors.Error {
	return &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
}
