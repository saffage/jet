package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/types"
)

type Const struct {
	owner *Scope
	name  *ast.Ident
	value *TypedValue
}

func NewConst(owner *Scope, value *TypedValue, name *ast.Ident) *Const {
	return &Const{
		owner: owner,
		name:  name,
		value: value,
	}
}

func (v *Const) Owner() *Scope         { return v.owner }
func (v *Const) Type() types.Type      { return v.value.Type }
func (v *Const) Value() constant.Value { return v.value.Value }
func (v *Const) Name() string          { return v.name.Name }
func (v *Const) Ident() *ast.Ident     { return v.name }
func (v *Const) Node() ast.Node        { return nil }
