package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Type struct {
	owner *Scope
	node  *ast.TypeDecl
	t     types.Type
}

func NewType(owner *Scope, t types.Type, node *ast.TypeDecl) *Type {
	return &Type{owner, node, t}
}

func (sym *Type) Type() types.Type { return sym.t }
func (sym *Type) Node() ast.Node   { return sym.node }
func (sym *Type) Name() string     { return sym.node.Name.Data }
func (sym *Type) Ident() ast.Ident { return sym.node.Name }
func (sym *Type) Owner() *Scope    { return sym.owner }
