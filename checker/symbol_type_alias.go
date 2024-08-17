package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

type TypeAlias struct {
	owner *Scope
	t     *types.Alias
	decl  *ast.LetDecl
}

func NewTypeAlias(owner *Scope, t *types.TypeDesc, decl *ast.LetDecl) *TypeAlias {
	assert(!types.IsUntyped(t))

	return &TypeAlias{
		owner: owner,
		t:     types.NewAlias(t.Base(), decl.Decl.Name.Ident()),
		decl:  decl,
	}
}

func (sym *TypeAlias) Owner() *Scope    { return sym.owner }
func (sym *TypeAlias) Type() types.Type { return types.NewTypeDesc(sym.t) }
func (sym *TypeAlias) Name() string     { return sym.Ident().Ident() }
func (sym *TypeAlias) Ident() ast.Ident { return sym.decl.Decl.Name }
func (sym *TypeAlias) Node() ast.Node   { return sym.decl }

func (check *Checker) resolveTypeAliasDecl(decl *ast.LetDecl) {
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

	check.newDef(decl.Decl.Name, sym)
	check.setType(decl, typedesc)
	report.TaggedDebugf("checker", "alias: set type: %s", typedesc)
}
