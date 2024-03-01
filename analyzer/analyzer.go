package analyzer

import (
	"simpl/ast"
	sErrors "simpl/errors"
	"simpl/tokens"
)

// Performs type checking and static analysis
func Prepare(trees []*ast.AST) []*sErrors.Error {
	types := []map[string]ast.DataType{}
	errors := []*sErrors.Error{}

	var setType func(node *ast.Node, scope int)
	setType = func(node *ast.Node, scope int) {
		switch node.Token.Type {
		case tokens.NUMBER:
			node.DataType = ast.Number
		case tokens.TRUE, tokens.FALSE:
			node.DataType = ast.Bool
		case tokens.IDENTIFIER:
			for i := len(types) - 1; i >= 0; i-- {
				val, found := types[i][node.Token.Value]
				if found {
					node.DataType = val
					return
				}
			}
			errors = append(errors, &sErrors.Error{Message: "undefined variable", Type: sErrors.ReferenceError, Token: node.Token})
		case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SLASH:
			node.DataType = ast.Number
			setType(node.Left, scope)
			setType(node.Right, scope)
			if node.Left.DataType != ast.Number {
				errors = append(errors, &sErrors.Error{Message: "invalid operation", Type: sErrors.TypeError, Token: node.Token})
			}
			if node.Right.DataType != ast.Number {
				errors = append(errors, &sErrors.Error{Message: "expected number", Type: sErrors.TypeError, Token: node.Right.Token})
			}
		case tokens.COLON_EQUAL:
			_, found := types[len(types)-1][node.Left.Token.Value]
			setType(node.Right, scope)
			if !found {
				node.Left.DataType = node.Right.DataType
				types[len(types)-1][node.Left.Token.Value] = node.Left.DataType
			} else {
				errors = append(errors, &sErrors.Error{Message: "variable reassignment not allowed", Type: sErrors.ReferenceError, Token: node.Token})
			}
		case tokens.EQUAL:
			var dType ast.DataType
			for i := len(types) - 1; i >= 0; i-- {
				val, found := types[i][node.Left.Token.Value]
				if found {
					dType = val
					break
				}
			}
			if dType == 0 {
				errors = append(errors, &sErrors.Error{Message: "undefined variable", Type: sErrors.ReferenceError, Token: node.Left.Token})
			}
			node.Left.DataType = dType
			setType(node.Right, scope)
			if dType != 0 && dType != node.Right.DataType {
				errors = append(errors, &sErrors.Error{Message: "trying to assign a different type", Type: sErrors.TypeError, Token: node.Token})
			}
			node.Left.DataType = node.Right.DataType
		}
	}

	for _, t := range trees {
		for len(types) <= t.Scope {
			types = append(types, map[string]ast.DataType{})
		}
		if t.Scope < len(types)-1 {
			types = types[:t.Scope+1]
		}
		setType(t.Root, t.Scope)
	}
	return errors
}
