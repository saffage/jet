package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

type TypeAlias struct {
	owner *Scope
	t     *types.Alias
	decl  *ast.Decl
}

func NewTypeAlias(owner *Scope, t *types.TypeDesc, decl *ast.Decl) *TypeAlias {
	assert(!types.IsUntyped(t))

	return &TypeAlias{
		owner: owner,
		t:     types.NewAlias(t.Base(), decl.Ident.Name),
		decl:  decl,
	}
}

func (sym *TypeAlias) Owner() *Scope     { return sym.owner }
func (sym *TypeAlias) Type() types.Type  { return types.NewTypeDesc(sym.t) }
func (sym *TypeAlias) Name() string      { return sym.decl.Ident.Name }
func (sym *TypeAlias) Ident() *ast.Ident { return sym.decl.Ident }
func (sym *TypeAlias) Node() ast.Node    { return sym.decl }

func (check *Checker) resolveTypeAliasDecl(decl *ast.Decl) {
	t := check.typeOf(decl.Value)
	if t == nil {
		return
	}

	typedesc := types.AsTypeDesc(t)

	if typedesc == nil {
		check.errorf(decl.Value, "expression is not a type")
		return
	}

	sym := NewTypeAlias(check.scope, typedesc, decl)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(decl.Ident, sym)
	check.setType(decl, typedesc)
	report.TaggedDebugf("checker", "alias: set type: %s", typedesc)
}
