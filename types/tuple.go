package types

import (
	"slices"
	"strings"
)

var Unit = &Tuple{}

type Tuple struct {
	types []Type
}

func NewTuple(types ...Type) *Tuple {
	return &Tuple{types}
}

// 2 tuples have the same type only when all their elements have equal types.
func (t *Tuple) Equals(other Type) bool {
	if t2 := AsPrimitive(other); t2 != nil && t2.kind == KindAny {
		return true
	}

	if t2 := AsTuple(other); t2 != nil {
		return slices.EqualFunc(
			t.types,
			t2.types,
			func(a, b Type) bool { return a.Equals(b) },
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

func IsTuple(t Type) bool {
	return AsTuple(t) != nil
}

func AsTuple(t Type) *Tuple {
	if t != nil {
		if tuple, _ := t.Underlying().(*Tuple); tuple != nil {
			return tuple
		}
	}

	return nil
}

func WrapInTuple(t Type) *Tuple {
	if tuple, _ := t.(*Tuple); tuple != nil {
		return tuple
	}

	return NewTuple(t)
}
