package parser

import (
	"simpl/ast"
	"simpl/errors"
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

func (s *ParseSource) parseParens() (*ast.Node, *errors.SyntaxError) {
	s.current++
	node, err := s.parseExpression(sTokens.Precedences[0])
	if err != nil {
		return nil, err
	}
	nextToken := s.tokens[s.current+1]
	if s.tokens[s.current+1].Type != sTokens.RIGHT_PAREN {
		return nil, &errors.SyntaxError{Message: "Expected RIGHT_PAREN, got something else", Line: nextToken.Line, Char: nextToken.Char}
	}

	s.current++
	return node, nil
}

func (s *ParseSource) parsePrefix() (*ast.Node, *errors.SyntaxError) {
	token := s.tokens[s.current]
	switch token.Type {
	case sTokens.IDENTIFIER, sTokens.NUMBER:
		return &ast.Node{Token: token, Type: ast.Default}, nil
	case sTokens.LEFT_PAREN:
		return s.parseParens()
	default:
		return nil, &errors.SyntaxError{Message: "Unexpected token", Line: token.Line, Char: token.Char}
	}
}

func (s *ParseSource) parseExpression(precedence int) (*ast.Node, *errors.SyntaxError) {
	left, err := s.parsePrefix()
	if err != nil {
		return nil, err
	}
	for s.tokens[s.current+1].Type != sTokens.SEMICOLON && precedence < sTokens.Precedences[s.tokens[s.current+1].Type] {
		s.current++
		token := s.tokens[s.current]
		nextLeft := &ast.Node{Left: left, Token: token, Type: ast.Expression}
		prec := sTokens.Precedences[token.Type]
		s.current++
		right, err := s.parseExpression(prec)
		if err != nil {
			return nil, err
		}
		nextLeft.Right = right
		left = nextLeft
	}

	return left, nil
}

func (s *ParseSource) Parse() ([]*ast.AST, *errors.SyntaxError) {
	trees := []*ast.AST{}

	sourceSize := len(s.tokens) - 1
	scope := 0
	scopeStarts := []sTokens.Token{}

	for s.current < sourceSize {
		tree := ast.NewAST()
		t := s.tokens[s.current]
		switch t.Type {
		case sTokens.LEFT_BRACE:
			scopeStarts = append(scopeStarts, t)
			scope++
			s.current++
		case sTokens.RIGHT_BRACE:
			if scope == 0 {
				return []*ast.AST{}, &errors.SyntaxError{Message: "Unexpected RIGHT_BRACE", Line: t.Line, Char: t.Char}
			}
			scope--
			scopeStarts = scopeStarts[:scope]
			s.current++
		case sTokens.IDENTIFIER:
			expression, err := s.parseExpression(sTokens.Precedences[sTokens.EOF])
			if err != nil {
				return []*ast.AST{}, err
			}
			tree.Root = expression
            tree.Scope = scope
			trees = append(trees, &tree)
			s.current += 2
		case sTokens.EOF:
			if scope > 0 {
				brace := scopeStarts[0]
				return []*ast.AST{}, &errors.SyntaxError{Message: "Scope not closed", Line: brace.Line, Char: brace.Char}
			}
			break
		default:
			return []*ast.AST{}, &errors.SyntaxError{Message: "Unexpected token: not a statement start", Line: t.Line, Char: t.Char}
		}
	}

	return trees, nil
}
