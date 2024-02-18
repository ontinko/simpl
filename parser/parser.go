package parser

import (
	"simpl/ast"
	"simpl/errors"
	sTokens "simpl/tokens"
)

func Parse(tokens *[]sTokens.Token) (*ast.AST, *errors.SyntaxError) {
	tree := ast.NewAST()
	for i, t := range *tokens {
		var nType ast.NodeType
		switch t.Type {
		case sTokens.PLUS, sTokens.MINUS, sTokens.STAR, sTokens.SLASH:
			nType = ast.Expression
		case sTokens.NUMBER, sTokens.IDENTIFIER:
			nType = ast.Default
		default:
			nType = ast.Statement
		}
		node := &ast.Node{Token: t, Type: nType, Left: nil, Right: nil}
		err := tree.Insert(node)
		if err != nil {
			return nil, err
		}
		if t.Type == sTokens.SEMICOLON {
			*tokens = (*tokens)[i+1:]
			break
		}
	}
	return &tree, nil
}
