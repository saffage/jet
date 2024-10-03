package types

import "fmt"

type Array struct {
	elem Type
	size int
}

func NewFixedArray(size int, t Type) *Array {
	if size < 0 {
		panic(fmt.Sprintf("invalid array size `%d`", size))
	}

	return &Array{t, size}
}

func (t *Array) Equal(expected Type) bool {
	if expected := As[*Array](expected); expected != nil {
		return t.size == expected.size && t.elem.Equal(expected.elem)
	}

	return false
}

func (t *Array) Underlying() Type {
	return t
}

func (t *Array) String() string {
	return fmt.Sprintf("Array(%d, %s)", t.size, t.elem)
}

func (t *Array) Size() int {
	return t.size
}

func (t *Array) ElemType() Type {
	return t.elem
}
