package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

type Const struct {
	owner *Scope
	name  *ast.Ident
	value *TypedValue
}

func NewConst(owner *Scope, value *TypedValue, name *ast.Ident) *Const {
	return &Const{
		owner: owner,
		name:  name,
		value: value,
	}
}

func (v *Const) Owner() *Scope         { return v.owner }
func (v *Const) Type() types.Type      { return v.value.Type }
func (v *Const) Value() constant.Value { return v.value.Value }
func (v *Const) Name() string          { return v.name.Name }
func (v *Const) Ident() *ast.Ident     { return v.name }
func (v *Const) Node() ast.Node        { return nil }

func (check *Checker) resolveConstDecl(node *ast.ConstDecl) {
	value := check.valueOf(node.Binding.Value)
	if value == nil {
		check.errorf(node.Binding.Value, "value is not a constant expression")
		return
	}

	tType := check.resolveVarType(node.Binding.Type, value.Type)
	if tType == nil {
		panic("unreachable")
	}

	if value.Type != nil {
		report.TaggedDebugf("checker", "const: value type: %s", value.Type)
	}

	report.TaggedDebugf("checker", "const: specified type: %s", tType)

	if value.Type != nil && !value.Type.Equals(tType) {
		check.errorf(
			node.Binding.Name,
			"type mismatch, expected '%s', got '%s'",
			tType,
			value.Type,
		)
		return
	}

	value.Type = tType

	report.TaggedDebugf("checker", "const type: %s", tType)
	sym := NewConst(check.scope, value, node.Binding.Name)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(node.Binding.Name, sym)
}
