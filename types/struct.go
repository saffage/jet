package types

import (
	"strings"
)

// TODO add named types.

type Struct struct {
	// TODO use ordered map.
	fields map[string]Type
}

func NewStruct(fields map[string]Type) *Struct {
	return &Struct{fields}
}

func (t *Struct) Equals(other Type) bool {
	if otherStruct, _ := other.Underlying().(*Struct); otherStruct != nil {
		for name, tField := range t.fields {
			tOtherField, ok := otherStruct.fields[name]

			if !ok || !tField.Equals(tOtherField) {
				return false
			}
		}

		return true
	}

	return false
}

func (t *Struct) Underlying() Type { return t }

func (t *Struct) String() string {
	buf := strings.Builder{}
	buf.WriteString("struct{")

	first := true
	for name, tField := range t.fields {
		if !first {
			buf.WriteString("; ")
		}

		buf.WriteString(name)
		buf.WriteByte(' ')
		buf.WriteString(tField.String())
		first = false
	}

	buf.WriteByte('}')
	return buf.String()
}

func (t *Struct) Fields() map[string]Type { return t.fields }

func IsStruct(t Type) bool { return AsStruct(t) != nil }

func AsStruct(t Type) *Struct {
	if t != nil {
		if s, _ := t.Underlying().(*Struct); t != nil {
			return s
		}
	}

	return nil
}
