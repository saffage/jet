package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

type Const struct {
	owner *Scope
	ident ast.Ident
	value *TypedValue
}

func NewConst(owner *Scope, value *TypedValue, ident ast.Ident) *Const {
	return &Const{owner, ident, value}
}

func (v *Const) Owner() *Scope         { return v.owner }
func (v *Const) Type() types.Type      { return v.value.Type }
func (v *Const) Value() constant.Value { return v.value.Value }
func (v *Const) Name() string          { return v.ident.String() }
func (v *Const) Ident() ast.Ident      { return v.ident }
func (v *Const) Node() ast.Node        { return nil }

func (check *checker) resolveConstDecl(decl *ast.LetDecl) {
	value := check.valueOf(decl.Value)
	if value == nil {
		check.errorf(decl.Value, "value is not a constant expression")
		return
	}

	tType := check.resolveVarType(decl.Decl.Type, value.Type)
	if tType == nil {
		panic("unreachable")
	}

	if value.Type != nil {
		report.TaggedDebugf("checker", "const: value type: %s", value.Type)
	}

	report.TaggedDebugf("checker", "const: specified type: %s", tType)

	if value.Type != nil && !value.Type.Equals(tType) {
		check.errorf(
			decl.Decl.Name,
			"type mismatch, expected '%s', got '%s'",
			tType,
			value.Type,
		)
		return
	}

	value.Type = tType

	report.TaggedDebugf("checker", "const type: %s", tType)
	sym := NewConst(check.scope, value, decl.Decl.Name)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(decl.Decl.Name, sym)
}
