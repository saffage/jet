package types

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/saffage/jet/ast"
	. "github.com/saffage/jet/internal/debug"
	"github.com/saffage/jet/report"
)

// Type represents an interface for types that can be compared for equality
// and rendered into a string buffer.
type Type interface {
	Equal(target Type) bool

	report.Renderer
}

//------------------------------------------------
// Primitive types
//------------------------------------------------

type (
	None   struct{}
	Never  struct{}
	Int    struct{}
	Float  struct{}
	String struct{}
)

var (
	NoneType   Type = None{}
	NeverType  Type = Never{}
	IntType    Type = Int{}
	FloatType  Type = Float{}
	StringType Type = String{}
)

func (None) Render(buf *strings.Builder) { buf.WriteString("None") }
func (None) Renderable() bool            { return true }
func (None) String() string              { return Render(NoneType) }
func (None) Equal(target Type) bool      { return Is[None](target) }

func (Never) Render(buf *strings.Builder) { buf.WriteString("Never") }
func (Never) Renderable() bool            { return true }
func (Never) String() string              { return Render(NeverType) }
func (Never) Equal(target Type) bool      { return true }

func (Int) Render(buf *strings.Builder) { buf.WriteString("Int") }
func (Int) Renderable() bool            { return true }
func (Int) String() string              { return Render(IntType) }
func (Int) Equal(target Type) bool      { return Is[Int](target) }

func (Float) Render(buf *strings.Builder) { buf.WriteString("Float") }
func (Float) Renderable() bool            { return true }
func (Float) String() string              { return Render(FloatType) }
func (Float) Equal(target Type) bool      { return Is[Float](target) }

func (String) Render(buf *strings.Builder) { buf.WriteString("String") }
func (String) Renderable() bool            { return true }
func (String) String() string              { return Render(StringType) }
func (String) Equal(target Type) bool      { return Is[String](target) }

//------------------------------------------------
// Module type (used as a placeholder in the checker)
//------------------------------------------------

type module struct{}

var moduleType Type = module{}

func (module) Render(buf *strings.Builder) { buf.WriteString("module") }
func (module) Renderable() bool            { return true }
func (module) String() string              { return "module" }
func (module) Equal(target Type) bool      { return false }

//------------------------------------------------
// Type Alias
//------------------------------------------------

type Alias struct {
	base   Type // Type of the expression that was used in alias declaration (can be an alias).
	actual Descriptor
	name   string
}

func NewAlias(t Type, name string) *Alias {
	if t == nil {
		panic("can't create unknown type alias")
	}

	if !Is[Descriptor](t) {
		panic("expected type descriptor")
	}

	return &Alias{
		base:   t,
		actual: removeAlias(t),
		name:   name,
	}
}

func (t *Alias) Equal(other Type) bool {
	return t.actual.Equal(SkipAlias(other))
}

func (t *Alias) String() string {
	return Render(t)
}

func (t *Alias) Renderable() bool {
	return true
}

func (t *Alias) Render(buf *strings.Builder) {
	buf.WriteString(t.name)

	// You can't ref non-alias primitive type other way than through compiler.
	if !IsAtom(t.base) {
		buf.WriteString(" aka ")

		SkipDescriptor(t.base).Render(buf)
	}
}

func SkipAlias(t Type) Type {
	if a, _ := t.(*Alias); a != nil {
		// What if `a.actual` is nil?
		return SkipAlias(a.actual)
	}
	return t
}

func removeAlias(t0 Type) Descriptor {
	t := t0
	a, ok := t.(*Alias)

	for ok && a != nil {
		if !IsResolved(t) {
			panic("unresolved type in type alias")
		}
		t = a.actual
		a, ok = t.(*Alias)
	}

	if !IsResolved(t) {
		panic("unresolved type in type alias")
	}

	desc, ok := As[Descriptor](t)

	if !ok {
		panic("type is not a type descriptor")
	}

	return desc
}

//------------------------------------------------
// Type Descriptor
//------------------------------------------------

type Descriptor struct {
	base Type
}

func NewDescriptor(t Type) Descriptor {
	if !IsResolved(t) {
		panic(fmt.Sprintf("type %s is not resolved", Render(t)))
	}

	if desc, _ := t.(Descriptor); desc.base != nil {
		return desc
	}

	return Descriptor{base: t}
}

func (t Descriptor) Equal(other Type) bool {
	if desc, _ := As[Descriptor](other); desc.base != nil {
		return t.base.Equal(desc.base)
	}

	return false
}

func (t Descriptor) String() string {
	return Render(t)
}

func (t Descriptor) Renderable() bool {
	return true
}

func (t Descriptor) Render(buf *strings.Builder) {
	buf.WriteString("type ")

	t.base.Render(buf)
}

func (t Descriptor) Base() Type {
	return t.base
}

func SkipDescriptor(t Type) Type {
	if typedesc, ok := t.(Descriptor); ok {
		return SkipDescriptor(typedesc.base)
	}

	return t
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

func NewFn(params TypeList, result, variadic Type) *Fn {
	return &Fn{
		params:   params,
		result:   result,
		variadic: variadic,
	}
}

func (t *Fn) Equal(expected Type) bool {
	if expected, _ := As[*Fn](expected); expected != nil {
		return (t.variadic != nil && t.variadic.Equal(expected.variadic) ||
			t.variadic == nil && expected.variadic == nil) &&
			t.result.Equal(expected.result) && t.params.Equal(expected.params)
	}

	return false
}

func (t *Fn) String() string {
	return Render(t)
}

func (t *Fn) Renderable() bool {
	return true
}

func (t *Fn) Render(buf *strings.Builder) {
	buf.WriteString("fn(")

	if t.params != nil {
		t.params.Render(buf)
	}

	buf.WriteString(")")

	if t.result != nil {
		buf.WriteByte(' ')

		t.result.Render(buf)
	}
}

func (t *Fn) Result() Type     { return t.result }
func (t *Fn) Params() TypeList { return t.params }
func (t *Fn) Variadic() Type   { return t.variadic }

func (t *Fn) CheckArgValues(values []*Value) (idx int, err error) {
	args := make(TypeList, len(values))

	for i := range args {
		args[i] = values[i].T
	}

	return -1, t.CheckArgs(args)
}

func (t *Fn) CheckArgs(args TypeList, argsNode ...*ast.Parens) error {
	Assert(args != nil)
	Assert(len(argsNode) < 2, "`argsNode` it's an optional parameter, not variadic")

	var node *ast.Parens

	if len(argsNode) == 1 {
		Assert(argsNode[0] != nil)
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

func (list TypeList) Render(buf *strings.Builder) {
	for i, param := range list {
		if i > 0 {
			buf.WriteString(", ")
		}

		param.Render(buf)
	}
}

//------------------------------------------------
// Custom\User-defined Type
//------------------------------------------------

var (
	BoolType = NewCustom(
		"Bool",
		nil,
		[]Variant{{Name: "False"}, {Name: "True"}},
	)

	False = BoolType.Variant(0)
	True  = BoolType.Variant(1)
)

type Custom struct {
	name     string
	fields   []Field
	variants []Variant
}

func NewCustom(name string, fields []Field, variants []Variant) *Custom {
	t := &Custom{
		name:     name,
		fields:   make([]Field, len(fields)),
		variants: make([]Variant, len(variants)),
	}

	// TODO validate input params

	for i, field := range fields {
		t.fields[i] = Field{
			Name:   field.Name,
			T:      field.T,
			parent: t,
			index:  i,
		}
	}

	for i, variant := range variants {
		t.variants[i] = Variant{
			Name:   variant.Name,
			Params: variant.Params,
			parent: t,
			index:  i,
		}
	}

	return t
}

func (t *Custom) Render(buf *strings.Builder) { buf.WriteString(t.name) }
func (t *Custom) Renderable() bool            { return true }
func (t *Custom) String() string              { return Render(t) }
func (t *Custom) Equal(target Type) bool      { return t == target }

func (t *Custom) Field(i int) *Field { return &t.fields[i] }
func (t *Custom) Fields() []Field    { return t.fields }
func (t *Custom) FieldsLen() int     { return len(t.fields) }

func (t *Custom) Variant(i int) *Variant { return &t.variants[i] }
func (t *Custom) Variants() []Variant    { return t.variants }
func (t *Custom) VariantsLen() int       { return len(t.variants) }

func (t *Custom) OnFields() iter.Seq2[int, Field]     { return slices.All(t.fields) }
func (t *Custom) OnVariants() iter.Seq2[int, Variant] { return slices.All(t.variants) }

// Field of [Custom] type.
type Field struct {
	Name string
	T    Type

	parent *Custom
	index  int
}

func (f *Field) Type() Type          { return f.T }
func (f *Field) ParentType() *Custom { return f.parent }
func (f *Field) Index() int          { return f.index }

// Variant of [Custom] type.
type Variant struct {
	Name   string
	Params TypeList

	parent *Custom
	index  int
}

func (v *Variant) Type() Type          { return v.parent }
func (v *Variant) ParentType() *Custom { return v.parent }
func (v *Variant) Index() int          { return v.index }

//------------------------------------------------
// Util Functions
//------------------------------------------------

// func Underlying(t Type) Type {
// 	if a, _ := t.(*Distinct); a != nil {
// 		return Underlying(a.actual)
// 	}
// 	return t
// }

func Render(t Type) string {
	if t == nil {
		return "<invalid-type>"
	}

	buf := strings.Builder{}

	t.Render(&buf)

	return buf.String()
}

func As[T Type](t Type) (T, bool) {
	if t != nil {
		if t, ok := t.(T); ok {
			return t, true
		}
		if t, ok := SkipAlias(t).(T); ok {
			return t, true
		}
	}
	var zero T
	// if IsAtom(zero) {
	// 	if p, _ := t.(*Parameter); p != nil {
	// 		return zero, p.Equal(zero)
	// 	}
	// }
	return zero, false
}

func Is[T Type](t Type) bool {
	_, ok := As[T](t)
	return ok
}

func IsResolved(t Type) bool {
	switch t := SkipAlias(t).(type) {
	case nil:
		// It's not clear if nil is resolved here or not, it depends on the
		// context.
		return false

	case None, Never, Int, Float, String:
		// Atom types are always resolved.
		return true

	// case *Parameter:
	// 	// Not sure about it.
	// 	return true

	case module:
		// Not sure about it.
		return true

	case Descriptor:
		return IsResolved(t.base)

	// case *Array:
	// 	return isTypeResolved(t.elem)

	case *Fn:
		for _, param := range t.params {
			if !IsResolved(param) {
				return false
			}
		}

		if !IsResolved(t.result) {
			return false
		}

		if t.variadic != nil {
			return IsResolved(t.variadic)
		}

		return true

	case *Custom:
		for _, field := range t.fields {
			if !IsResolved(field.T) {
				return false
			}
		}

		for _, variant := range t.variants {
			if slices.IndexFunc(variant.Params, IsResolved) >= 0 {
				return false
			}
		}

		return true

	default:
		panic("unreachable")
	}
}

func IsAtom(t Type) bool {
	switch t.(type) {
	case None, Never, Int, Float, String:
		return true

	default:
		return false
	}
}

// Trying to turn a type to the typed analog.
//
// If target type is not provided, then result is never nil.
//
// If target type is provided, then type will be checked for equality with
// target type, and if its not, then nil will be returned.
func IntoTyped(v *Value, target ...Type) Type {
	expected := Type(nil)

	if len(target) != 0 {
		expected = SkipAlias(target[0])
		Assert(expected != nil, "argument is nil")
	}

	t := SkipAlias(v.T)

	if t == nil || expected != nil && !t.Equal(expected) {
		return nil
	}

	return t
}

func FromConstant(value ConstantValue) Type {
	switch value.Kind() {
	case ConstantInt:
		return IntType

	case ConstantFloat:
		return FloatType

	case ConstantString:
		return StringType

	default:
		panic("unreachable")
	}
}
