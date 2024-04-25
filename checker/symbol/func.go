package symbol

import (
	"github.com/saffage/jet/ast"
)

type Func struct {
	base
}

func NewFunc(id ID, node *ast.FuncDecl, owner Scope) *Func {
	return &Func{
		base: base{
			id:    id,
			owner: owner,
			name:  node.Name,
			node:  node,
		},
	}
}
