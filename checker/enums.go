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
	node  *ast.EnumDecl
}

func NewEnum(owner *Scope, body *Scope, t *types.TypeDesc, node *ast.EnumDecl) *Enum {
	if !types.IsEnum(t.Base()) {
		panic("expected enum type")
	}
	if body.Parent() != owner {
		panic("invalid local scope parent")
	}
	return &Enum{owner, body, t, node}
}

func (sym *Enum) Owner() *Scope     { return sym.owner }
func (sym *Enum) Type() types.Type  { return sym.t }
func (sym *Enum) Name() string      { return sym.node.Name.Name }
func (sym *Enum) Ident() *ast.Ident { return sym.node.Name }
func (sym *Enum) Node() ast.Node    { return sym.node }

func (check *Checker) resolveEnumDecl(node *ast.EnumDecl) {
	local := NewScope(check.scope, "enum "+node.Name.Name)
	fields := make([]string, 0, len(node.Body.Nodes))

	// TODO field names as distinct symbols.
	for _, ident := range node.Body.Nodes {
		ident, _ := ident.(*ast.Ident)
		if ident == nil {
			check.errorf(ident, "expected field identifier for enum")
			continue
		}

		fields = append(fields, ident.Name)
	}

	t := types.NewTypeDesc(types.NewEnum(fields...))
	sym := NewEnum(check.scope, local, t, node)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}
	check.newDef(node.Name, sym)
}

func (check *Checker) enumMember(node *ast.MemberAccess, t *types.Enum) types.Type {
	fieldIdent, _ := node.Selector.(*ast.Ident)
	if fieldIdent == nil {
		check.errorf(node.Selector, "expected identifier for enum member")
		return t
	}

	idx := slices.Index(t.Fields(), fieldIdent.Name)
	if idx == -1 {
		check.errorf(fieldIdent, "type has no member named '%s'", fieldIdent.Name)
		return t
	}

	return t
}
