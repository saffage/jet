package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Func struct {
	owner *Scope
	local *Scope
	t     *types.Func
	node  *ast.FuncDecl
}

func NewFunc(owner *Scope, local *Scope, t *types.Func, node *ast.FuncDecl) *Func {
	return &Func{owner, local, t, node}
}

func (sym *Func) Owner() *Scope     { return sym.owner }
func (sym *Func) Type() types.Type  { return sym.t }
func (sym *Func) Name() string      { return sym.node.Name.Name }
func (sym *Func) Ident() *ast.Ident { return sym.node.Name }
func (sym *Func) Node() ast.Node    { return sym.node }
