package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/token"
)

type Type struct {
	base
}

func NewType(id ID, node *ast.AliasDecl, owner Scope) *Type {
	return &Type{
		base: base{
			owner: owner,
			id:    id,
			name:  node.Name,
			node:  node,
		},
	}
}

// `type_` returns:
//   - nil - expression has no type.
//   - [types.Unknown] - expression should have type but doesn't (symbol definition is written after or symbol not defined).

// Meaning:
//   - `type_` - type of the expression.
//   - `required` - symbol that is required for this expression but have no type.
//   - `where` - identifier that refers to the symbol (declared or not).
func TypeOf(owner Scope, expr ast.Node) (type_ types.Type, required Symbol, where *ast.Ident) {
	switch node := ast.UnwrapParen(expr).(type) {
	case *ast.ParenExpr:
		panic("unreachable")

	case *ast.Ident:
		if sym := owner.Resolve(node.Name); sym != nil {
			// symbol is defined
			if type_ := owner.Resolve(node.Name).Type(); type_ != nil {
				// symbol have a type (resolved)
				return type_, sym, node
			}
			// symbol have no type (not yet resolver)
			return &types.Unknown{}, sym, node
		}
		// identifier is undefined
		return &types.Unknown{}, nil, node

	case *ast.Literal:
		switch node.Kind {
		case token.Int:
			return &types.Primitive{Kind: types.UntypedInt}, nil, nil

		default:
			panic("todo")
		}

	case *ast.CurlyList:
		listScope := NewLocalScope(owner)
		walker := ast.NewWalker(listScope)
		walker.Walk(node.List)

		if listScope.Type != nil {
			return listScope.Type, nil, nil
		}

		return &types.Unknown{}, nil, nil
	}

	return nil, nil, nil
}
