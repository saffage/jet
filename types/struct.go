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
	for _, t := range fields {
		if IsUntyped(t) {
			panic("struct fields cannot be untyped")
		}
	}
	return &Struct{fields}
}

func (t *Struct) Equals(other Type) bool {
	if t2 := AsStruct(other); t2 != nil {
		for name, tField := range t.fields {
			t2Field, ok := t2.fields[name]
			if !ok || !tField.Equals(t2Field) {
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
