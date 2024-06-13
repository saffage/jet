package types

import "fmt"

type Array struct {
	size int
	elem Type
}

func NewArray(size int, t Type) *Array {
	if size < 0 {
		panic(fmt.Sprintf("invalid array size (%d)", size))
	}
	return &Array{size, t}
}

func (t *Array) Equals(other Type) bool {
	if t2 := AsPrimitive(other); t2 != nil {
		return t2.kind == KindAny
	}
	if t2 := AsArray(other); t2 != nil {
		return t.size == t2.size && t.elem.Equals(t2.elem)
	}
	return false
}

func (t *Array) Underlying() Type { return t }

func (t *Array) String() string { return fmt.Sprintf("[%d]%s", t.size, t.elem) }

func (t *Array) Size() int { return t.size }

func (t *Array) ElemType() Type { return t.elem }

func IsArray(t Type) bool { return AsArray(t) != nil }

func AsArray(t Type) *Array {
	if t != nil {
		if array, _ := t.Underlying().(*Array); array != nil {
			return array
		}
	}

	return nil
}
