package types

import (
	"iter"
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/debug"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

// Type represents an interface for types that can be compared for equality
// and rendered into a string buffer.
type Type interface {
	Equal(target Type) bool
	String() string

	text.Renderer
}

// Named type represents named type from prelude or user-defined type.
//
// For example `Int` or `Float` it's a nominal type.
//
// This type may have type parameters.
type Named struct {
	// The module where the type is defined. Environment of the module may
	// differ from the owner, but owner should be somewhere in
	// the module hierarchy.
	module *Module

	// The environment where the type is defined.
	owner *Env

	// The environment where type contains their fields and constructors.
	//
	// Can be nil if the type does not defined any fields or constructors.
	body *Env

	// An AST node of the type definition.
	node *ast.TypeDef

	// Type parameters (aka "generics" or "parametric polymorphism").
	//
	// Empty for non-generic types.
	args []Type

	// TODO docs
	fields []*Field

	// TODO docs
	variants []*Variant

	// Type symbol's publicity.
	publicity Publicity
}

func (t *Named) Type() Type {
	return t
}

func (t *Named) Name() string {
	if ident := t.Ident(); ident != nil {
		return ident.Name()
	}
	return ""
}

func (t *Named) Node() ast.Node {
	return t.node
}

func (t *Named) Ident() ast.Ident {
	if t.node == nil {
		return nil
	}
	return t.node.Ident
}

func (t *Named) Value() *Value {
	// TODO type variant inference.
	return &Value{T: t}
}

func (t *Named) Owner() *Env {
	return t.owner
}

func (t *Named) Field(i int) *Field { return t.fields[i] }
func (t *Named) Fields() []*Field   { return t.fields }
func (t *Named) FieldsLen() int     { return len(t.fields) }

func (t *Named) Variant(i int) *Variant { return t.variants[i] }
func (t *Named) Variants() []*Variant   { return t.variants }
func (t *Named) VariantsLen() int       { return len(t.variants) }

func (t *Named) OnFields() iter.Seq2[int, *Field]     { return slices.All(t.fields) }
func (t *Named) OnVariants() iter.Seq2[int, *Variant] { return slices.All(t.variants) }

//------------------------------------------------
// Nominal Type Alias
//------------------------------------------------

// NOTE: this is not a real type.
type Alias struct {
	// The module where the type is defined. Environment of the module may
	// differ from the owner, but owner should be somewhere in
	// the module hierarchy.
	module *Module

	// The environment where the type is defined.
	owner *Env

	// An AST node of the type definition.
	node *ast.TypeAlias

	// Type parameters (aka "generics" or "parametric polymorphism").
	//
	// Empty for non-generic types.
	args []Type

	// Type symbol's publicity.
	publicity Publicity
}

// func NewNominalAlias(t Type, name string) *NominalAlias {
// 	if t == nil {
// 		panic("can't create unknown type alias")
// 	}
//
// 	// return &NominalAlias{
// 	// 	base:   t,
// 	// 	actual: removeAlias(t),
// 	// 	name:   name,
// 	// }
//
// 	return nil
// }

func (t *Alias) Type() Type {
	return t
}

func (t *Alias) Name() string {
	if ident := t.Ident(); ident != nil {
		return ident.Name()
	}
	return ""
}

func (t *Alias) Node() ast.Node {
	return t.node
}

func (t *Alias) Ident() ast.Ident {
	if t.node == nil {
		return nil
	}
	return t.node.Ident
}

func (t *Alias) Value() *Value {
	return &Value{T: t}
}

func (t *Alias) Owner() *Env {
	return t.owner
}

func (t *Alias) IsExternal() bool {
	if t.node == nil {
		return false
	}
	_, isExternal := t.node.Expr.(*ast.External)
	return isExternal
}

// Returned type is never a type variable.
func (t *Alias) RemoveAlias() Type {
	return nil
}

//------------------------------------------------
// Type Variable
//------------------------------------------------

type TypeVariable struct {
	// index of the variable in the inference context.
	// Not meant to be used outside the context itself.
	id uint64

	// Variable name, if it's defined explicitly in code.
	// Otherwise the type variable is "meta" (generated).
	name string
}

func (v *TypeVariable) IsMeta() bool {
	return v.name == ""
}

//------------------------------------------------
// Function Type
//------------------------------------------------

type Fn struct {
	params   TypeList
	labels   map[string]Type
	result   Type
	variadic Type
}

func (t *Fn) Result() Type {
	return t.result
}

func (t *Fn) Params() TypeList {
	return t.params
}

func (t *Fn) Variadic() Type {
	return t.variadic
}

func (t *Fn) CheckArgValues(values []*Value) (idx int, err error) {
	args := make(TypeList, len(values))

	for i := range args {
		args[i] = values[i].T
	}

	return -1, t.CheckArgs(args)
}

func (t *Fn) CheckArgs(args TypeList, argsNode ...*ast.Parens) error {
	debug.Assert(args != nil)
	debug.Assert(len(argsNode) < 2, "`argsNode` it's an optional parameter, not variadic")

	var node *ast.Parens

	if len(argsNode) == 1 {
		debug.Assert(argsNode[0] != nil)
		node = argsNode[0]
	}

	// Example:
	//
	//	params  args    diff    idx
	//	1       2       -1      1
	//	2       1        1      1
	//	0       3       -3      0
	//	3       0        3      0
	var diff = len(t.params) - len(args)

	if diff > 0 || diff < 0 && t.variadic == nil {
		return errIncorrectArity(node, len(t.params), len(args))
	}

	var errs []error
	var arg = func(i int) ast.Node {
		if i < len(argsNode) {
			return argsNode[i]
		}
		return nil
	}

	for i := range len(t.params) {
		var expected = t.params[i]
		var actual = args[i]

		if !actual.Equal(expected) {
			err := errArgTypeMismatch(arg(i), actual, expected, i, false)
			errs = append(errs, err)
		}
	}

	// Check variadic.
	for i, tArg := range args[len(t.params):] {
		if !tArg.Equal(t.variadic) {
			i += len(t.params)
			err := errArgTypeMismatch(arg(i), tArg, t.variadic, i, true)
			errs = append(errs, err)
		}
	}

	return report.Join(errs...)
}

//------------------------------------------------
// Type List
//------------------------------------------------

type TypeList []Type

func (list TypeList) Equal(target []Type) bool {
	if len(list) != len(target) {
		return false
	}

	for i, param := range list {
		if !param.Equal(SkipAlias(target[i])) {
			return false
		}
	}

	return true
}

func (list TypeList) Renderable() bool {
	return true
}

func (list TypeList) Render(buf text.Writer) (rendered bool) {
	for i, param := range list {
		if i > 0 {
			buf.WriteString(", ")
		}

		param.Render(buf)
	}

	return true
}
