package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/constant"
)

type Const struct {
	id    ID
	owner Scope

	type_ types.Type
	value constant.Value

	node *ast.GenericDecl
	name *ast.Ident
}

func NewConst(owner Scope, node *ast.GenericDecl, name *ast.Ident) (*Const, error) {
	if node.Kind != ast.ConstDecl {
		return nil, NewError(node, "expected constant declaration")
	}

	if node.Field.Value == nil {
		return nil, NewError(name, "value is required for constant")
	}

	value := constant.FromNode(node.Field.Value)
	sym := &Const{
		id:    nextID(),
		owner: owner,
		// type_: types.FromConstant(value.Kind()),
		value: value,
		node:  node,
		name:  name,
	}

	return sym, nil
}

func (v *Const) ID() ID            { return v.id }
func (v *Const) Owner() Scope      { return v.owner }
func (v *Const) Type() types.Type  { return v.type_ }
func (v *Const) Name() string      { return v.name.Name }
func (v *Const) Ident() *ast.Ident { return v.name }
func (v *Const) Node() ast.Node    { return v.node }

func (v *Const) setType(t types.Type) { v.type_ = t }
