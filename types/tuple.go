package types

import (
	"slices"
	"strings"
)

type Tuple struct {
	types []Type
}

var Unit = NewTuple()

func NewTuple(types ...Type) *Tuple {
	return &Tuple{types}
}

// 2 tuples have the same type only when all their elements have equal types.
//
// NOTE: name of the elements are not required to be the same.
func (t *Tuple) Equals(other Type) bool {
	other = other.Underlying()

	if otherTuple, ok := other.(*Tuple); ok {
		return slices.EqualFunc(
			t.types,
			otherTuple.types,
			func(a, b Type) bool { return a.Equals(b) },
		)
	} else if underlying := t.Underlying(); underlying != t {
		// The tuple has 1 element and can be equals to the non-tuple type.
		return underlying.Equals(other)
	}

	return false
}

// Unnamed tuple with 1 element is equals to the type of this element.
// Otherwise a tuple have the unique type.
func (t *Tuple) Underlying() Type {
	if len(t.types) == 1 {
		return t.types[0].Underlying()
	}

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

func (t *Tuple) Types() []Type { return t.types }

func (t *Tuple) Len() int { return len(t.types) }
