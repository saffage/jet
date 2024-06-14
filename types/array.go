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

func (t *Array) Equals(expected Type) bool {
	if expected := AsPrimitive(expected); expected != nil {
		return expected.kind == KindAny
	}

	if expected := AsArray(expected); expected != nil {
		return t.size == expected.size && t.elem.Equals(expected.elem)
	}

	return false
}

func (t *Array) Underlying() Type {
	return t
}

func (t *Array) String() string {
	return fmt.Sprintf("[%d]%s", t.size, t.elem)
}

func (t *Array) Size() int {
	return t.size
}

func (t *Array) ElemType() Type {
	return t.elem
}

func IsArray(t Type) bool {
	return AsArray(t) != nil
}

func AsArray(t Type) *Array {
	if t != nil {
		if array, _ := t.Underlying().(*Array); array != nil {
			return array
		}
	}

	return nil
}
