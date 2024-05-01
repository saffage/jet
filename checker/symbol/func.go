package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
)

type Func struct {
	id    ID
	owner Scope

	type_ types.Type
	node  *ast.FuncDecl
}

func NewFunc(owner Scope, t types.Type, node *ast.FuncDecl) *Func {
	return &Func{
		id:    nextID(),
		owner: owner,
		type_: t,
		node:  node,
	}
}

func (v *Func) ID() ID            { return v.id }
func (v *Func) Owner() Scope      { return v.owner }
func (v *Func) Type() types.Type  { return v.type_ }
func (v *Func) Name() string      { return v.node.Name.Name }
func (v *Func) Ident() *ast.Ident { return v.node.Name }
func (v *Func) Node() ast.Node    { return v.node }

func (v *Func) setType(t types.Type) { v.type_ = t }
