package types

import (
	"github.com/saffage/jet/ast"
)

type TypeDef struct {
	owner    *Env
	t        *Alias
	typedesc *TypeDesc
	decl     *ast.TypeDecl
	opaque   bool
	unique   bool
}

func NewTypeDef(owner *Env, t *TypeDesc, decl *ast.TypeDecl) *TypeDef {
	assert(!IsUntyped(SkipTypeDesc(t)), "cannot use untyped typedesc for type definition")

	return &TypeDef{
		owner:    owner,
		t:        NewAlias(t.base, decl.Name.String()),
		typedesc: t,
		decl:     decl,
	}
}

func (sym *TypeDef) Name() string     { return sym.Ident().String() }
func (sym *TypeDef) Type() Type       { return sym.typedesc }
func (sym *TypeDef) Node() ast.Node   { return sym.decl }
func (sym *TypeDef) Ident() ast.Ident { return sym.decl.Name }
func (sym *TypeDef) Owner() *Env      { return sym.owner }

// func (check *checker) resolveTypeAliasDecl(decl *ast.TypeDecl) {
// 	t := check.typeOf(decl.Expr)
// 	if t == nil {
// 		return
// 	}

// 	typedesc := AsTypeDesc(t)

// 	if typedesc == nil {
// 		check.errorf(decl.Expr, "expression is not a type")
// 		return
// 	}

// 	sym := NewTypeAlias(check.env, typedesc, decl)

// 	if defined := check.env.Define(sym); defined != nil {
// 		check.problem(errorAlreadyDefined(sym.Ident(), defined.Ident()))
// 		return
// 	}

// 	check.newDef(decl.Name, sym)
// 	check.setType(decl, typedesc)
// 	report.TaggedDebugf("checker", "alias: set type: %s", typedesc)
// }
