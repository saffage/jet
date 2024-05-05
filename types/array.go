package types

import "fmt"

type Array struct {
	size int
	elem Type
}

func NewArray(size int, t Type) *Array {
	if size < -1 {
		panic(fmt.Sprintf("invalid array size (%d)", size))
	}

	return &Array{size, t}
}

func (t *Array) Equals(other Type) bool {
	if otherArray, ok := other.Underlying().(*Array); ok {
		return (t.size == -1 || t.size == otherArray.size) && t.elem.Equals(otherArray.elem)
	}

	return false
}

func (t *Array) Underlying() Type { return t }

func (t *Array) String() string {
	if t.size == -1 {
		return "[_]" + t.elem.String()
	}

	return fmt.Sprintf("[%d]%s", t.size, t.elem)
}

func (t *Array) Size() int { return t.size }

func (t *Array) ElemType() Type { return t.elem }

func IsArray(t Type) bool { return AsRef(t) != nil }

func AsArray(t Type) *Array {
	if t != nil {
		if array, _ := t.Underlying().(*Array); array != nil {
			return array
		}
	}

	return nil
}
