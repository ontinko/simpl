package parser

import (
	"simpl/ast"
	"simpl/errors"
	sTokens "simpl/tokens"
)

func Parse(tokens *[]sTokens.Token, scope *int) (*ast.AST, *errors.SyntaxError) {
	tree := ast.NewAST()
	for i, t := range *tokens {
		var nType ast.NodeType
		switch t.Type {
		// it would probably be better to have this kind of check in the tree
		case sTokens.LEFT_BRACE:
			if tree.Root != nil {
				return nil, &errors.SyntaxError{Message: "Unexpected {", Line: t.Line, Char: t.Char}
			}
			(*scope)++
			continue
		case sTokens.RIGHT_BRACE:
			(*scope)--
			continue
		case sTokens.PLUS, sTokens.MINUS, sTokens.STAR, sTokens.SLASH:
			nType = ast.Expression
		case sTokens.NUMBER, sTokens.IDENTIFIER:
			nType = ast.Default
		default:
			nType = ast.Statement
		}
		if i == len(*tokens)-1 && t.Type != sTokens.SEMICOLON {
			return nil, &errors.SyntaxError{Message: "Expecting semicolon at the end of a statement", Line: t.Line, Char: t.Char}
		}
		tree.Scope = (*scope)
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
