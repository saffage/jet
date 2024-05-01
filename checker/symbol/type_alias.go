package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/internal/assert"
)

type TypeAlias struct {
	id    ID
	owner Scope
	type_ types.Type
	node  *ast.TypeAliasDecl
	name  *ast.Ident
}

func NewTypeAlias(owner Scope, t types.Type, node *ast.TypeAliasDecl) *TypeAlias {
	assert.Ok(t == nil || types.IsTypeDesc(t))

	return &TypeAlias{
		id:    nextID(),
		owner: owner,
		type_: t,
		node:  node,
		name:  node.Name,
	}
}

func (v *TypeAlias) ID() ID            { return v.id }
func (v *TypeAlias) Owner() Scope      { return v.owner }
func (v *TypeAlias) Type() types.Type  { return v.type_ }
func (v *TypeAlias) Name() string      { return v.name.Name }
func (v *TypeAlias) Ident() *ast.Ident { return v.name }
func (v *TypeAlias) Node() ast.Node    { return v.node }

func (v *TypeAlias) setType(t types.Type) { v.type_ = t }
