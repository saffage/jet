package checker

import (
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Enum struct {
	owner *Scope
	body  *Scope
	t     *types.TypeDesc
	decl  *ast.LetDecl
}

func NewEnum(owner *Scope, body *Scope, t *types.TypeDesc, decl *ast.LetDecl) *Enum {
	if !types.IsEnum(t.Base()) {
		panic("expected enum type")
	}
	if body.Parent() != owner {
		panic("invalid local scope parent")
	}
	return &Enum{owner, body, t, decl}
}

func (sym *Enum) Owner() *Scope    { return sym.owner }
func (sym *Enum) Type() types.Type { return sym.t }
func (sym *Enum) Name() string     { return sym.Ident().Ident() }
func (sym *Enum) Ident() ast.Ident { return sym.decl.Decl.Name }
func (sym *Enum) Node() ast.Node   { return sym.decl }

func (check *Checker) resolveEnumDecl(decl *ast.LetDecl, value *ast.EnumType) {
	bodyScope := NewScope(check.scope, "enum "+decl.Decl.Name.Ident())
	fields := make([]string, 0, len(value.Fields))

	for _, ident := range value.Fields {
		fields = append(fields, ident.Data)
	}

	ty := types.NewTypeDesc(types.NewEnum(fields...))
	sym := NewEnum(check.scope, bodyScope, ty, decl)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}
	check.newDef(decl.Decl.Name, sym)
}

func (check *Checker) enumMember(node *ast.Dot, t *types.Enum) types.Type {
	idx := slices.Index(t.Fields(), node.Y.Data)
	if idx == -1 {
		check.errorf(node.Y, "type has no member named '%s'", node.Y.Data)
	}
	return t
}
