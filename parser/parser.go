package parser

import (
	"fmt"
	"simpl/ast"
	"simpl/errors"
	"simpl/tokens"
	sTokens "simpl/tokens"
)

type ParseSource struct {
	tokens  []sTokens.Token
	current int
}

func New(tokens []sTokens.Token) ParseSource {
	return ParseSource{
		tokens:  tokens,
		current: 0,
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

func (s *ParseSource) parsePrefix() (*ast.Expression, *errors.Error) {
	token := s.tokens[s.current]
	switch token.Type {
	case sTokens.BANG:
		node := &ast.Expression{Token: token}
		s.current++
		expression, err := s.parsePrefix()
		if err != nil {
			return nil, err
		}
		node.Left = expression
		return node, nil
	case sTokens.IDENTIFIER, sTokens.NUMBER, sTokens.TRUE, sTokens.FALSE:
		return &ast.Expression{Token: token}, nil
	case sTokens.LEFT_PAREN:
		return s.parseParens()
	default:
		return nil, &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
	}
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
		var nextLeft *ast.Expression
		switch token.Type {
		case sTokens.STAR, sTokens.SLASH, sTokens.PLUS, sTokens.MINUS, sTokens.AND, sTokens.OR, sTokens.DOUBLE_EQUAL, sTokens.NOT_EQUAL, sTokens.LESS, sTokens.GREATER, sTokens.LESS_EQUAL, sTokens.GREATER_EQUAL:
			nextLeft = &ast.Expression{Token: token}
		default:
			return nil, &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
		}
		nextLeft.Left = left
		prec := sTokens.Precedences[token.Type]
		s.current++
		right, err := s.parseExpression(prec, endToken)
		if err != nil {
			return nil, err
		}
		nextLeft.Right = right
		left = nextLeft
	}

	return left, nil
}

func (s *ParseSource) unexpectedError(token sTokens.Token) *errors.Error {
	return &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
}

func (s *ParseSource) Parse(baseScope int) (*ast.Program, *errors.Error) {
	statements := []ast.Statement{}

	sourceSize := len(s.tokens) - 1
	scope := baseScope
	scopeStarts := []sTokens.Token{}

MainLoop:
	for s.current < sourceSize {
		token := s.tokens[s.current]
		switch token.Type {
		case sTokens.LEFT_BRACE:
			s.current++
			block, err := s.Parse(scope + 1)
			if err != nil {
				return nil, err
			}
			statements = append(statements, block.Statements...)
		case sTokens.RIGHT_BRACE:
			if scope == 0 {
				return nil, s.unexpectedError(token)
			}
			s.current++
			break MainLoop
		case sTokens.IDENTIFIER:
			var stmt ast.Assignment
			operator := s.tokens[s.current+1]
			if operator.Type != sTokens.COLON_EQUAL && operator.Type != sTokens.EQUAL {
				return nil, s.unexpectedError(operator)
			}
			s.current += 2
			exp, err := s.parseExpression(sTokens.Precedences[sTokens.EOF], sTokens.SEMICOLON)
			if err != nil {
				return nil, err
			}
			stmt.Var = token
			stmt.Operator = operator
			stmt.Exp = exp
			stmt.Scope = scope

			statements = append(statements, &stmt)
			s.current += 2
		case sTokens.IF, sTokens.WHILE:
			var stmt ast.Conditional
			s.current++
			condition, err := s.parseExpression(sTokens.Precedences[sTokens.EOF], sTokens.LEFT_BRACE)
			if err != nil {
				return nil, err
			}
			s.current += 2

			thenBlock, err := s.Parse(scope + 1)
			if err != nil {
				return nil, err
			}
			stmt.Token = token
			stmt.Condition = condition
			stmt.Then = thenBlock
			stmt.Scope = scope

			token = s.tokens[s.current]
			if token.Type == tokens.ELSE {
				s.current += 2
				elseBlock, err := s.Parse(scope + 1)
				if err != nil {
					return nil, err
				}
				stmt.Else = elseBlock
			}
			statements = append(statements, &stmt)
		case sTokens.EOF:
			if scope != 0 {
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
