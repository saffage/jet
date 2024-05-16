package types

import (
	"fmt"
	"slices"
	"strings"
)

// TODO add named types.

type StructField struct {
	Name string
	Type Type
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

func (t *Struct) Equals(other Type) bool {
	if t2 := AsStruct(other); t2 != nil {
		return slices.EqualFunc(t.fields, t2.fields, func(f1, f2 StructField) bool {
			return f1.Name == f2.Name && f1.Type.Equals(f2.Type)
		})
	}
	return false
}

func (t *Struct) Underlying() Type { return t }

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

func IsStruct(t Type) bool { return AsStruct(t) != nil }

func AsStruct(t Type) *Struct {
	if t != nil {
		if s, _ := t.Underlying().(*Struct); s != nil {
			return s
		}
	}

	return nil
}
