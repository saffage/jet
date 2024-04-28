package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
)

type Const struct {
	base
	value constant.Value
}

func NewConst(id ID, name *ast.Ident, node *ast.GenericDecl, owner Scope) *Const {
	if node.Kind != ast.ConstDecl {
		panic(NewError(node, "expected constant declaration"))
	}

	if node.Field.Value == nil {
		panic(NewError(name, "value is required for constant"))
	}

	return &Const{
		base: base{
			id:    id,
			owner: owner,
			name:  name,
			node:  node,
		},
		value: constant.FromNode(node.Field.Value),
	}
}
