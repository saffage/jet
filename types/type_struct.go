//go:build ignore

package types

import (
	"fmt"
	"slices"
	"strings"
)

type StructField struct {
	Type Type
	Name string
}

type Struct struct {
	fields []StructField
}

func NewStruct(fields ...StructField) *Struct {
	fieldNames := make([]string, 0, len(fields))
	for _, field := range fields {
		if IsUntyped(field.Type) {
			panic("struct fields cannot be untyped")
		}
		if slices.Index(fieldNames, field.Name) != -1 {
			panic(fmt.Sprintf("duplicate fields in %#v", fields))
		}
		fieldNames = append(fieldNames, field.Name)
	}
	return &Struct{fields}
}

func (t *Struct) Equal(other Type) bool {
	if t2 := As[*Primitive](other); t2 != nil {
		return t2.kind == KindAny
	}
	if t2 := As[*Struct](other); t2 != nil {
		return slices.EqualFunc(
			t.fields,
			t2.fields,
			func(f1, f2 StructField) bool {
				return f1.Name == f2.Name && f1.Type.Equal(f2.Type)
			},
		)
	} else if ref := As[*Ref](other); ref != nil && t == String {
		p, _ := ref.base.Underlying().(*Primitive)
		return p != nil && p.kind == KindChar
	}
	return false
}

func (t *Struct) Underlying() Type {
	return t
}

func (t *Struct) String() string {
	buf := strings.Builder{}
	buf.WriteString("struct{")

	for i, field := range t.fields {
		if i != 0 {
			buf.WriteString("; ")
		}

		buf.WriteString(field.Name)
		buf.WriteByte(' ')
		buf.WriteString(field.Type.String())
	}

	buf.WriteByte('}')
	return buf.String()
}

func (t *Struct) Fields() []StructField { return t.fields }
