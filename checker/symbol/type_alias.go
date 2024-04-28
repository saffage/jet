package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/internal/assert"
)

type TypeAlias struct {
	base
}

func NewTypeAlias(id ID, t types.Type, node *ast.TypeAliasDecl, owner Scope) *TypeAlias {
	assert.Ok(t == nil || types.IsTypeDesc(t))

	return &TypeAlias{
		base: base{
			id:    id,
			owner: owner,
			type_: t,
			name:  node.Name,
			node:  node,
		},
	}
}
