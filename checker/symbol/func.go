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
			owner: owner,
			id:    id,
			name:  node.Name,
			node:  node,
		},
	}
}
