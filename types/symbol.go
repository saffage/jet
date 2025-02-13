package types

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/text"
)

type Symbol interface {
	Type() Type       // Type of the symbol.
	Name() string     // Name of the symbol.
	Node() ast.Node   // Related AST node.
	Ident() ast.Ident // Related identifier AST node.
	Value() *Value    // Value of the symbol (if it's known at compile time)
	Owner() *Env      // Scope where the symbol was defined.
}

func PathOf(symbol Symbol) string {
	return symbol.Owner().Path() + "." + symbol.Name()
}

//------------------------------------------------
// Binding
//------------------------------------------------

type Binding struct {
	owner   *Env
	local   *Env
	value   *Value
	field   *Field
	variant *Variant
	params  []*Binding

	node        *ast.Decl
	declNode    *ast.LetDecl // May be nil.
	labelNode   *ast.Lower   // May be nil.
	variantNode *ast.Variant // May be nil.

	externName string
	isParam    bool
}

func NewBinding(
	owner *Env,
	local *Env,
	value *Value,
	decl *ast.Decl,
	letNode *ast.LetDecl,
) *Binding {
	return &Binding{
		owner:    owner,
		local:    local,
		value:    value,
		node:     decl,
		declNode: letNode,
	}
}

func NewVariant(
	owner *Env,
	local *Env,
	variant *Variant,
	params []*Binding,
	node *ast.Variant,
) *Binding {
	tParams := make(TypeList, len(params))

	for i, param := range params {
		tParams[i] = param.Type()
	}

	var decl *ast.Decl

	if node != nil {
		decl = &ast.Decl{Ident: node.Name}
	}

	return &Binding{
		owner:       owner,
		local:       local,
		variant:     variant,
		params:      params,
		value:       &Value{T: NewFunction(tParams, variant.Type(), nil)},
		node:        decl,
		variantNode: node,
	}
}

func NewField(
	owner *Env,
	field *Field,
	node *ast.Decl,
	label *ast.Lower,
) *Binding {
	return &Binding{
		owner:     owner,
		field:     field,
		value:     &Value{T: field.T},
		node:      node,
		labelNode: label,
	}
}

func (sym *Binding) Type() Type         { return sym.value.T }
func (sym *Binding) Node() ast.Node     { return sym.node }
func (sym *Binding) Name() string       { return sym.Ident().Name() }
func (sym *Binding) Value() *Value      { return sym.value }
func (sym *Binding) Owner() *Env        { return sym.owner }
func (sym *Binding) Local() *Env        { return sym.local }
func (sym *Binding) IsParam() bool      { return sym.isParam }
func (sym *Binding) IsField() bool      { return sym.field != nil }
func (sym *Binding) IsVariant() bool    { return sym.variant != nil }
func (sym *Binding) IsLocal() bool      { return !sym.IsField() && !sym.isParam }
func (sym *Binding) IsExtern() bool     { return sym.node == nil }
func (sym *Binding) ExternName() string { return sym.externName }
func (sym *Binding) Params() []*Binding { return sym.params }

func (sym *Binding) Ident() ast.Ident {
	if sym.IsVariant() {
		if sym.node != nil {
			return sym.node.Ident
		}

		return &ast.Lower{Data: sym.variant.Name}
	}

	if sym.IsField() {
		if sym.node != nil {
			return sym.node.Ident
		}

		return &ast.Lower{Data: sym.field.Name}
	}

	if sym.IsExtern() {
		assert(sym.externName != "", "binding without node must have `externName`")
		return &ast.Lower{Data: sym.externName}
	}

	return sym.node.Ident
}

func (sym *Binding) ParamTypes() TypeList {
	params := make(TypeList, len(sym.params))

	for i, param := range sym.params {
		params[i] = param.Type()
	}

	return params
}

func (sym *Binding) Variadic() Type {
	if fn, _ := As[*Function](sym.value.T); fn != nil {
		return fn.Variadic()
	}
	return nil
}

func (v *Binding) ValueNode() ast.Node {
	if v.declNode != nil {
		return v.declNode.Value
	}
	return nil
}

//------------------------------------------------
// Module
//------------------------------------------------

type Module struct {
	*TypeInfo

	Env     *Env
	Imports []*Module

	stmts     *ast.Stmts
	file      *text.File
	name      string
	completed bool
}

func NewModule(env *Env, name string, file *text.File, stmts *ast.Stmts) *Module {
	return &Module{
		TypeInfo: newTypeInfo(),
		Env:      env,
		file:     file,
		stmts:    stmts,
		name:     name,
	}
}

func (m *Module) Type() Type       { return moduleType }
func (m *Module) Name() string     { return m.name }
func (m *Module) Node() ast.Node   { return m.stmts }
func (m *Module) Ident() ast.Ident { return nil } // TODO: just use *ast.Name with zero-initialized range
func (m *Module) Value() *Value    { return &Value{T: moduleType} }
func (m *Module) Owner() *Env      { return m.Env.parent }

func (m *Module) TypeOf(expr ast.Node) Type {
	if expr != nil {
		if t := m.TypeInfo.TypeOf(expr); t != nil {
			return t
		}
		if ident, _ := expr.(*ast.Lower); ident != nil {
			if sym := m.SymbolOf(ident); sym != nil {
				return sym.Type()
			}
		}
	}
	return nil
}

func (m *Module) ValueOf(expr ast.Node) *Value {
	if expr != nil {
		if t := m.TypeInfo.ValueOf(expr); t != nil {
			return t
		}
	}
	return nil
}

func (m *Module) SymbolOf(ident ast.Ident) Symbol {
	if sym := m.TypeInfo.SymbolOf(ident); sym != nil {
		return sym
	}
	if sym, _ := m.Env.Lookup(ident.Name()); sym != nil {
		return sym
	}
	return nil
}

//------------------------------------------------
// Type Def Symbol
//------------------------------------------------

type TypeDef struct {
	owner  *Env
	local  *Env
	custom *Custom

	// This field must be specified as it defines the name of the symbol.
	node *ast.TypeDef
}

func NewTypeDef(owner, local *Env, custom *Custom, node *ast.TypeDef) *TypeDef {
	return &TypeDef{
		owner:  owner,
		local:  local,
		custom: custom,
		node:   node,
	}
}

// NOTE: name of the type comes from [TypeDef.custom].
func NewExternTypeDef(owner, local *Env, custom *Custom) *TypeDef {
	return &TypeDef{
		owner:  owner,
		local:  local,
		custom: custom,
	}
}

func (sym *TypeDef) Type() Type     { return NewDescriptor(sym.custom) }
func (sym *TypeDef) Node() ast.Node { return sym.node }
func (sym *TypeDef) Value() *Value  { return &Value{T: sym.Type()} }
func (sym *TypeDef) Owner() *Env    { return sym.owner }
func (sym *TypeDef) Local() *Env    { return sym.local }
func (sym *TypeDef) IsExtern() bool { return sym.node == nil }

func (sym *TypeDef) Name() string {
	if sym.IsExtern() {
		return sym.custom.name
	}

	return sym.Ident().Name()
}

func (sym *TypeDef) Ident() ast.Ident {
	if sym.IsExtern() {
		return &ast.Upper{Data: sym.custom.name}
	}

	return sym.node.Ident
}

//------------------------------------------------
// Type Alias Symbol
//------------------------------------------------

type TypeAlias struct {
	owner *Env
	node  *ast.TypeAlias
	alias *Alias
}

func NewTypeAlias(owner *Env, alias *Alias, node *ast.TypeAlias) *TypeAlias {
	return &TypeAlias{
		owner: owner,
		alias: alias,
		node:  node,
	}
}

// NOTE: name of the type alias comes from [TypeAlias.alias].
func NewExternTypeAlias(owner *Env, extern Type) *TypeAlias {
	assert(IsAtom(extern))

	return &TypeAlias{
		owner: owner,
		alias: &Alias{base: extern, name: Render(extern)},
	}
}

func (sym *TypeAlias) Type() Type     { return NewDescriptor(sym.alias) }
func (sym *TypeAlias) Node() ast.Node { return sym.node }
func (sym *TypeAlias) Value() *Value  { return &Value{T: sym.Type()} }
func (sym *TypeAlias) Owner() *Env    { return sym.owner }

func (sym *TypeAlias) Name() string {
	if sym.node != nil {
		return sym.node.Ident.Name()
	}

	return sym.alias.name
}

func (sym *TypeAlias) Ident() ast.Ident {
	if sym.node != nil {
		return sym.node.Ident
	}

	return &ast.Upper{Data: sym.alias.name}
}

//------------------------------------------------
// Debug Symbol Printer
//------------------------------------------------

type debugPrinter interface {
	debug() string
}

func (sym *Binding) debug() string {
	result := symbolTypeNoQualifier(sym)
	mods := make([]string, 0, 5)
	if sym.isParam {
		mods = append(mods, "param")
	}
	if sym.IsField() {
		mods = append(mods, "field")
	}
	if sym.IsVariant() {
		mods = append(mods, "variant")
	}
	if sym.IsExtern() {
		mods = append(mods, "extern")
	}
	if len(mods) != 0 {
		result += "(" + strings.Join(mods, ", ") + ")"
	}
	return result
}

func symbolTypeNoQualifier(sym Symbol) string {
	return strings.TrimPrefix(
		strings.TrimPrefix(fmt.Sprintf("%T", sym), "*"),
		"types.",
	)
}
