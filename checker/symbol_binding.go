package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

type Binding struct {
	owner    *Scope
	t        types.Type
	node     *ast.Decl
	letNode  *ast.LetDecl // May be nil.
	label    *ast.Name    // May be nil.
	isParam  bool
	isField  bool
	isGlobal bool
}

func NewBinding(owner *Scope, t types.Type, decl *ast.Decl, letNode *ast.LetDecl) *Binding {
	assert(!types.IsUntyped(t))
	return &Binding{owner: owner, t: t, node: decl, letNode: letNode}
}

func (v *Binding) Owner() *Scope    { return v.owner }
func (v *Binding) Type() types.Type { return v.t }
func (v *Binding) Name() string     { return v.Ident().String() }
func (v *Binding) Ident() ast.Ident { return v.node.Name }
func (v *Binding) Node() ast.Node   { return v.node }
func (v *Binding) IsLocal() bool    { return !v.isParam && !v.isField && !v.isGlobal }
func (v *Binding) IsParam() bool    { return v.isParam }
func (v *Binding) IsField() bool    { return v.isField }
func (v *Binding) IsGlobal() bool   { return v.isGlobal }

func (v *Binding) Value() ast.Node {
	if v.letNode != nil {
		return v.letNode.Value
	}
	return nil
}

func (check *checker) resolveVarDecl(node *ast.LetDecl) {
	// 'tValue' can be nil.
	tValue, ok := check.resolveVarValue(node.Value)
	if !ok {
		return
	}

	// 'tType' cannot be nil.
	tType := check.resolveVarType(node.Decl.Type, tValue)
	if tType == nil {
		return
	}

	if tValue != nil {
		report.TaggedDebugf("checker", "var value type: %s", tValue)
	}

	report.TaggedDebugf("checker", "var specified type: %s", tType)

	if tValue != nil && !tValue.Equals(tType) {
		check.errorf(
			node.Value,
			"type mismatch, expected '%s', got '%s'",
			tType,
			tValue,
		)
		return
	}

	tType = types.SkipUntyped(tType)

	// Set a correct type to the value.
	if tValue := types.AsArray(tValue); tValue != nil && types.IsUntyped(tValue.ElemType()) {
		// TODO this causes codegen to generate two similar typedefs.
		check.setType(node.Value, tType)
		report.TaggedDebugf("checker", "var set value type: %s", tType)
	}

	report.TaggedDebugf("checker", "var type: %s", tType)
	sym := NewBinding(check.scope, tType, node.Decl, node)
	sym.isGlobal = sym.owner == check.module.Scope

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(node.Decl.Name, sym)
}

func (check *checker) resolveVarValue(value ast.Node) (types.Type, bool) {
	if value != nil {
		t := check.typeOf(value)
		if t == nil {
			return nil, false
		}

		if types.IsTypeDesc(t) {
			check.errorf(value, "expected value, got type '%s' instead", t)
			return nil, false
		}

		return t, true
	}

	return nil, true
}

func (check *checker) resolveVarType(typeExpr ast.Node, value types.Type) types.Type {
	if typeExpr == nil {
		return value
	}

	t := check.typeOf(typeExpr)
	if t == nil {
		return value
	}

	typedesc := types.AsTypeDesc(t)

	// Unit can be either value and type.
	if t.Equals(types.Unit) {
		typedesc = types.NewTypeDesc(types.Unit)
	}

	if typedesc == nil {
		check.errorf(typeExpr, "expression is not a type")
		return nil
	}

	return typedesc.Base()
}
