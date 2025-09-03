package types

//------------------------------------------------
// Custom\User-defined Type
//------------------------------------------------
/*
var (
	BoolType = NewCustom(
		"Bool",
		nil,
		[]Variant{{Name: "False"}, {Name: "True"}},
	)

	False = BoolType.Variant(0)
	True  = BoolType.Variant(1)
)

// TODO support generics

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

func (t *Custom) String() string         { return Render(t) }
func (t *Custom) Equal(target Type) bool { return t == target }

func (t *Custom) Field(i int) *Field { return &t.fields[i] }
func (t *Custom) Fields() []Field    { return t.fields }
func (t *Custom) FieldsLen() int     { return len(t.fields) }

func (t *Custom) Variant(i int) *Variant { return &t.variants[i] }
func (t *Custom) Variants() []Variant    { return t.variants }
func (t *Custom) VariantsLen() int       { return len(t.variants) }

func (t *Custom) OnFields() iter.Seq2[int, Field]     { return slices.All(t.fields) }
func (t *Custom) OnVariants() iter.Seq2[int, Variant] { return slices.All(t.variants) }

func (t *Custom) Render(buf text.Writer) (rendered bool) {
	buf.WriteString(t.name)
	return true
}
*/

// Field of [Named] type.
type Field struct {
	Name string
	T    Type

	parent *Named
	index  int
}

func (f *Field) Type() Type         { return f.T }
func (f *Field) ParentType() *Named { return f.parent }
func (f *Field) Index() int         { return f.index }

// Variant of [Named] type.
type Variant struct {
	Name   string
	Params TypeList

	parent *Named
	index  int
}

func (v *Variant) Type() Type         { return v.parent }
func (v *Variant) ParentType() *Named { return v.parent }
func (v *Variant) Index() int         { return v.index }
