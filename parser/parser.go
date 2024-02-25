package parser

import (
	"fmt"
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

func (s *ParseSource) parseParens() (*ast.Node, *errors.Error) {
	s.current++
	node, err := s.parseExpression(0)
	if err != nil {
		return nil, err
	}
	nextToken := s.tokens[s.current+1]
	if s.tokens[s.current+1].Type != sTokens.RIGHT_PAREN {
		expected, got := sTokens.Representations[sTokens.RIGHT_PAREN], sTokens.Representations[nextToken.Type]
		return nil, &errors.Error{Message: fmt.Sprintf("Expected %s, got %s", expected, got), Token: nextToken, Type: errors.SyntaxError}
	}
	s.current++
	return node, nil
}

func (s *ParseSource) parsePrefix() (*ast.Node, *errors.Error) {
	token := s.tokens[s.current]
	switch token.Type {
	case sTokens.IDENTIFIER, sTokens.NUMBER:
		return &ast.Node{Token: token, Type: ast.Default}, nil
	case sTokens.LEFT_PAREN:
		return s.parseParens()
	default:
		return nil, &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
	}
}

func (s *ParseSource) parseExpression(precedence int) (*ast.Node, *errors.Error) {
	left, err := s.parsePrefix()
	if err != nil {
		return nil, err
	}
	for s.tokens[s.current+1].Type != sTokens.SEMICOLON && precedence < sTokens.Precedences[s.tokens[s.current+1].Type] {
		s.current++
		token := s.tokens[s.current]
		var nextLeft *ast.Node
		switch token.Type {
		case sTokens.STAR, sTokens.SLASH, sTokens.PLUS, sTokens.MINUS:
			nextLeft = &ast.Node{Token: token, Type: ast.Expression}
		default:
			return nil, &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
		}
		nextLeft.Left = left
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

func (s *ParseSource) UnexpectedError(token sTokens.Token) *errors.Error {
	return &errors.Error{Message: fmt.Sprintf("unexpected %s", token.View()), Token: token, Type: errors.SyntaxError}
}

func (s *ParseSource) Parse() ([]*ast.AST, *errors.Error) {
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
				return []*ast.AST{}, s.UnexpectedError(t)
			}
			scope--
			scopeStarts = scopeStarts[:scope]
			s.current++
		case sTokens.IDENTIFIER:
			assignment := s.tokens[s.current+1]
			if assignment.Type != sTokens.COLON_EQUAL && assignment.Type != sTokens.EQUAL {
				return []*ast.AST{}, s.UnexpectedError(assignment)
			}
			tree.Scope = scope
			tree.Root = &ast.Node{Token: assignment, Type: ast.Statement}
			tree.Root.Left = &ast.Node{Token: t, Type: ast.Default}
			s.current += 2
			right, err := s.parseExpression(sTokens.Precedences[sTokens.EOF])
			if err != nil {
				return []*ast.AST{}, err
			}
			tree.Root.Right = right
			trees = append(trees, &tree)
			s.current += 2
		case sTokens.EOF:
			if scope > 0 {
				brace := scopeStarts[0]
				return []*ast.AST{}, &errors.Error{Message: "Scope not closed", Token: brace, Type: errors.SyntaxError}
			}
			break
		default:
			return []*ast.AST{}, s.UnexpectedError(t)
		}
	}

	return trees, nil
}
