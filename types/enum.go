package types

import "strings"

type Enum struct {
	fields []string
}

func NewEnum(fields ...string) *Enum {
	return &Enum{fields}
}

func (t *Enum) Equals(other Type) bool {
	if t2 := AsPrimitive(other); t2 != nil {
		return t2.kind == KindAny
	}
	if t2 := AsEnum(other); t2 != nil {
		if len(t2.fields) != len(t.fields) {
			return false
		}

		for i := range t.fields {
			if t.fields[i] != t2.fields[i] {
				return false
			}
		}

		return true
	}
	return false
}

func (t *Enum) Underlying() Type { return t }

func (t *Enum) String() string {
	buf := strings.Builder{}
	buf.WriteString("enum{")

	first := true
	for _, field := range t.fields {
		if !first {
			buf.WriteString("; ")
		}

		buf.WriteString(field)
		first = false
	}

	buf.WriteByte('}')
	return buf.String()
}

func (t *Enum) Fields() []string { return t.fields }

func IsEnum(t Type) bool { return AsEnum(t) != nil }

func AsEnum(t Type) *Enum {
	if t != nil {
		if e, _ := t.Underlying().(*Enum); e != nil {
			return e
		}
	}

	return nil
}
