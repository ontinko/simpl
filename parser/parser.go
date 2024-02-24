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

func (s *ParseSource) parsePrefix() (*ast.Node, *errors.SyntaxError) {
	token := s.tokens[s.current]
	switch token.Type {
	case sTokens.IDENTIFIER, sTokens.NUMBER:
		return &ast.Node{Token: token, Type: ast.Default}, nil
	default:
		return nil, &errors.SyntaxError{Message: "Unexpected token: parsing prefix", Line: token.Line, Char: token.Char}
	}
}

func (s *ParseSource) parseExpression(precedence int) (*ast.Node, *errors.SyntaxError) {
	left, err := s.parsePrefix()
	if err != nil {
		return nil, err
	}
	for s.tokens[s.current+1].Type != sTokens.SEMICOLON && precedence < sTokens.Priorities[s.tokens[s.current+1].Type] {
		s.current++
		token := s.tokens[s.current]
		nextLeft := &ast.Node{Left: left, Token: token, Type: ast.Expression}
		prec := sTokens.Priorities[token.Type]
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
			assignment_op := s.tokens[s.current+1]
			if assignment_op.Type != sTokens.EQUAL && assignment_op.Type != sTokens.COLON_EQUAL {
				return []*ast.AST{}, &errors.SyntaxError{Message: "Unexpected token: expecting = or :=", Line: assignment_op.Line, Char: assignment_op.Char}
			}
			s.current += 2
			expression, err := s.parseExpression(sTokens.Priorities[sTokens.EOF])
			if err != nil {
				return []*ast.AST{}, err
			}
			tree.Root = &ast.Node{Token: assignment_op, Type: ast.Statement}
			tree.Root.Left = &ast.Node{Token: t, Type: ast.Default}
			tree.Root.Right = expression
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
