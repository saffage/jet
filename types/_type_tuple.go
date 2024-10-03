//go:build ignore

package types

import (
	"slices"
	"strings"
)

type Tuple struct {
	types []Type
}

func NewTuple(types ...Type) *Tuple {
	return &Tuple{types}
}

// 2 tuples have the same type only when all their elements have equal
func (t *Tuple) Equal(other Type) bool {
	if t2 := As[*Tuple](other); t2 != nil {
		return slices.EqualFunc(
			t.types,
			t2.types,
			func(a, b Type) bool { return a.Equal(b) },
		)
	}

	return false
}

func (t *Tuple) Underlying() Type {
	return t
}

func (t *Tuple) String() string {
	buf := strings.Builder{}
	buf.WriteByte('(')

	for i := range t.types {
		if i > 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(t.types[i].String())
	}

	buf.WriteByte(')')
	return buf.String()
}

func (t *Tuple) Types() []Type {
	return t.types
}

func (t *Tuple) Len() int {
	return len(t.types)
}

func WrapInTuple(t Type) *Tuple {
	if tuple, _ := t.(*Tuple); tuple != nil {
		return tuple
	}

	return NewTuple(t)
}
