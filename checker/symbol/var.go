package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
)

type Var struct {
	id    ID
	owner Scope

	type_ types.Type

	node *ast.GenericDecl
	name *ast.Ident
}

func NewVar(owner Scope, t types.Type, node *ast.GenericDecl, name *ast.Ident) *Var {
	return &Var{
		id:    nextID(),
		owner: owner,
		type_: t,
		node:  node,
		name:  name,
	}
}

func (v *Var) ID() ID            { return v.id }
func (v *Var) Owner() Scope      { return v.owner }
func (v *Var) Type() types.Type  { return v.type_ }
func (v *Var) Name() string      { return v.name.Name }
func (v *Var) Ident() *ast.Ident { return v.name }
func (v *Var) Node() ast.Node    { return v.node }

func (v *Var) setType(t types.Type) { v.type_ = t }
