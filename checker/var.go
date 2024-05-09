package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Var struct {
	owner    *Scope
	t        types.Type
	node     *ast.Binding
	name     *ast.Ident
	isParam  bool
	isField  bool
	isGlobal bool
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
func (v *Var) IsLocal() bool     { return !v.isParam && !v.isField && !v.isGlobal }
func (v *Var) IsParam() bool     { return v.isParam }
func (v *Var) IsField() bool     { return v.isField }
func (v *Var) IsGlobal() bool    { return v.isGlobal }
