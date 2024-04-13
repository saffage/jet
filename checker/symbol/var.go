package symbol

import "github.com/saffage/jet/ast"

type Var struct {
	base
}

func NewVar(id ID, name *ast.Ident, node *ast.GenericDecl, owner Scope) *Var {
	return &Var{
		base: base{
			owner: owner,
			id:    id,
			name:  name,
			node:  node,
		},
	}
}
