package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type TypeAlias struct {
	owner *Scope
	t     *types.Alias
	node  *ast.TypeAliasDecl
	name  *ast.Ident
}

func NewTypeAlias(owner *Scope, t *types.TypeDesc, node *ast.TypeAliasDecl) *TypeAlias {
	return &TypeAlias{
		owner: owner,
		t:     types.NewAlias(t.Base(), node.Name.Name),
		node:  node,
		name:  node.Name,
	}
}

func (sym *TypeAlias) Owner() *Scope     { return sym.owner }
func (sym *TypeAlias) Type() types.Type  { return types.NewTypeDesc(sym.t) }
func (sym *TypeAlias) Name() string      { return sym.name.Name }
func (sym *TypeAlias) Ident() *ast.Ident { return sym.name }
func (sym *TypeAlias) Node() ast.Node    { return sym.node }

func (check *Checker) resolveTypeAliasDecl(node *ast.TypeAliasDecl) {
	t := check.typeOf(node.Expr)
	if t == nil {
		return
	}

	typedesc := types.AsTypeDesc(t)

	if typedesc == nil {
		check.errorf(node.Expr, "expression is not a type")
		return
	}

	sym := NewTypeAlias(check.scope, typedesc, node)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(node.Name, sym)
	check.setType(node, typedesc)
}
