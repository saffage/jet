package types

import (
	"github.com/saffage/jet/ast"
)

type Binding struct {
	owner      *Env
	local      *Env
	value      *Value
	parent     *TypeDef
	decl       *ast.Decl
	node       *ast.LetDecl // May be nil.
	label      *ast.Lower   // May be nil.
	variant    *ast.Variant // May be nil.
	externName string
	params     []*Binding
	isExtern   bool
	isParam    bool
	isField    bool
	isVariant  bool
	isGlobal   bool
}

func NewBinding(
	owner *Env,
	local *Env,
	value *Value,
	decl *ast.Decl,
	letNode *ast.LetDecl,
) *Binding {
	assert(!IsUntyped(value.T), "untyped binding is illegal")
	// assert(decl != nil)

	return &Binding{
		owner: owner,
		local: local,
		value: value,
		decl:  decl,
		node:  letNode,
	}
}

func NewVariant(
	owner *Env,
	local *Env,
	params []*Binding,
	parent *TypeDef,
	node *ast.Variant,
) *Binding {
	tParams := make(Params, len(params))

	for i, param := range params {
		tParams[i] = param.Type()
	}

	return &Binding{
		owner:     owner,
		local:     local,
		parent:    parent,
		value:     &Value{T: NewFunction(tParams, parent.typedesc.base, nil)},
		decl:      &ast.Decl{Name: node.Name},
		variant:   node,
		params:    params,
		isVariant: true,
	}
}

func NewField(
	owner *Env,
	parent *TypeDef,
	t Type,
	node *ast.Decl,
	label *ast.Lower,
) *Binding {
	assert(!IsUntyped(t), "field cannot be untyped")

	return &Binding{
		owner:   owner,
		parent:  parent,
		value:   &Value{T: t},
		decl:    node,
		label:   label,
		isField: true,
	}
}

func (sym *Binding) Type() Type { return sym.value.T }
func (sym *Binding) Name() string {
	if sym.decl.Name == nil {
		return "_"
	}
	return sym.Ident().String()
}
func (sym *Binding) Node() ast.Node     { return sym.decl }
func (sym *Binding) Ident() ast.Ident   { return sym.decl.Name }
func (sym *Binding) Owner() *Env        { return sym.owner }
func (sym *Binding) Local() *Env        { return sym.local }
func (sym *Binding) IsParam() bool      { return sym.isParam }
func (sym *Binding) IsField() bool      { return sym.isField }
func (sym *Binding) IsGlobal() bool     { return sym.isGlobal }
func (sym *Binding) IsLocal() bool      { return !sym.isParam && !sym.isField && !sym.isGlobal }
func (sym *Binding) IsExtern() bool     { return sym.isExtern }
func (sym *Binding) ExternName() string { return sym.externName }
func (sym *Binding) Params() []*Binding { return sym.params }

func (sym *Binding) ParamTypes() Params {
	params := make(Params, len(sym.params))

	for i, param := range sym.params {
		params[i] = param.Type()
	}

	return params
}

func (sym *Binding) Variadic() Type {
	if fn := As[*Function](sym.value.T); fn != nil {
		return fn.Variadic()
	}
	return nil
}

func (v *Binding) ValueNode() ast.Node {
	if v.node != nil {
		return v.node.Value
	}
	return nil
}
