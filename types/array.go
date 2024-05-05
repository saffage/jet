package types

import "fmt"

type Array struct {
	size uint
	elem Type
}

func NewArray(size uint, t Type) *Array {
	return &Array{size, t}
}

func (t *Array) Equals(other Type) bool {
	other = other.Underlying()

	if otherArray, ok := other.(*Array); ok {
		return t.size == otherArray.size && t.elem.Equals(otherArray.elem)
	}

	return false
}

func (t *Array) Underlying() Type { return t }

func (t *Array) String() string { return fmt.Sprintf("[%d]%s", t.size, t.elem) }

func (t *Array) Size() uint { return t.size }

func (t *Array) ElemType() Type { return t.elem }
