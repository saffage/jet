package types

import (
	"iter"
	"slices"
)

var BoolType = NewCustom(
	"Bool",
	nil,
	[]Variant{{Name: "False"}, {Name: "True"}},
)

// Field of [Custom] type.
type Field struct {
	Name string
	T    Type

	parent Type
	index  int
}

func (f *Field) Type() Type       { return f.T }
func (f *Field) ParentType() Type { return f.parent }
func (f *Field) Index() int       { return f.index }

// Variant of [Custom] type.
type Variant struct {
	Name   string
	Params Params

	t     Type
	index int
}

func (v *Variant) Type() Type { return v.t }
func (v *Variant) Index() int { return v.index }

type Custom struct {
	fields   []Field
	variants []Variant
	name     string
}

func NewCustom(name string, fields []Field, variants []Variant) *Custom {
	t := &Custom{
		name:     name,
		fields:   make([]Field, len(fields)),
		variants: make([]Variant, len(variants)),
	}

	for i := range fields {
		assert(!IsUntyped(fields[i].T))

		t.fields[i] = Field{
			parent: t,
			T:      fields[i].T,
			Name:   fields[i].Name,
			index:  i,
		}
	}

	for i := range variants {
		assert(!slices.ContainsFunc(variants[i].Params, IsUntyped))

		t.variants[i] = Variant{
			t:      t,
			Params: variants[i].Params,
			Name:   variants[i].Name,
			index:  i,
		}
	}

	return t
}

func (t *Custom) Equal(expected Type) bool { return t == expected }
func (t *Custom) Underlying() Type         { return t }
func (t *Custom) String() string           { return t.name }

func (t *Custom) Field(i int) *Field { return &t.fields[i] }
func (t *Custom) Fields() []Field    { return t.fields }
func (t *Custom) FieldsLen() int     { return len(t.fields) }

func (t *Custom) Variant(i int) *Variant { return &t.variants[i] }
func (t *Custom) Variants() []Variant    { return t.variants }
func (t *Custom) VariantsLen() int       { return len(t.variants) }

func (t *Custom) OnFields() iter.Seq2[int, Field] {
	return slices.All(t.fields)
}

func (t *Custom) OnVariants() iter.Seq2[int, Variant] {
	return slices.All(t.variants)
}
