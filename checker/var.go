package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Var struct {
	owner *Scope
	t     types.Type
	node  *ast.Binding
	name  *ast.Ident
}

func NewVar(owner *Scope, t types.Type, node *ast.Binding, name *ast.Ident) *Var {
	return &Var{
		owner: owner,
		t:     t,
		node:  node,
		name:  name,
	}
}

func (v *Var) Owner() *Scope     { return v.owner }
func (v *Var) Type() types.Type  { return v.t }
func (v *Var) Name() string      { return v.name.Name }
func (v *Var) Ident() *ast.Ident { return v.name }
func (v *Var) Node() ast.Node    { return v.node }
