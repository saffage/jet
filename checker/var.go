package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

type Var struct {
	owner    *Scope
	t        types.Type
	node     *ast.Binding
	name     *ast.Ident
	value    ast.Node // TODO move somewhere else.
	isParam  bool
	isField  bool
	isGlobal bool
}

func NewVar(owner *Scope, t types.Type, node *ast.Binding, name *ast.Ident) *Var {
	assert.Ok(!types.IsUntyped(t))

	return &Var{
		owner: owner,
		t:     t,
		node:  node,
		name:  name,
	}
}

func (v *Var) Owner() *Scope     { return v.owner }
func (v *Var) Type() types.Type  { return v.t }
func (v *Var) Name() string      { return v.name.Name }
func (v *Var) Ident() *ast.Ident { return v.name }
func (v *Var) Node() ast.Node    { return v.node }
func (v *Var) Value() ast.Node   { return v.value }
func (v *Var) IsLocal() bool     { return !v.isParam && !v.isField && !v.isGlobal }
func (v *Var) IsParam() bool     { return v.isParam }
func (v *Var) IsField() bool     { return v.isField }
func (v *Var) IsGlobal() bool    { return v.isGlobal }

func (check *Checker) resolveVarDecl(node *ast.VarDecl) {
	// 'tValue' can be nil.
	tValue, ok := check.resolveVarValue(node.Value)
	if !ok {
		return
	}

	// 'tType' must be not nil.
	tType := check.resolveVarType(node.Binding.Type, tValue)
	if tType == nil {
		return
	}

	if tValue != nil {
		report.TaggedDebugf("checker", "var: value type: %s", tValue)
	}

	report.TaggedDebugf("checker", "var: specified type: %s", tType)

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

	report.TaggedDebugf("checker", "var type: %s", tType)
	sym := NewVar(check.scope, tType, node.Binding, node.Binding.Name)
	sym.value = node.Value
	sym.isGlobal = sym.owner == check.module.Scope

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(node.Binding.Name, sym)
}

func (check *Checker) resolveVarValue(value ast.Node) (types.Type, bool) {
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

func (check *Checker) resolveVarType(typeExpr ast.Node, value types.Type) types.Type {
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
