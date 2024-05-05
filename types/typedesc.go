package types

import (
	"fmt"

	"github.com/saffage/jet/internal/assert"
)

type TypeDesc struct {
	base Type
}

func NewTypeDesc(t Type) Type {
	assert.Ok(t != nil)

	if _, ok := t.(*TypeDesc); ok {
		return t
	}

	return &TypeDesc{base: t}
}

func (t *TypeDesc) Equals(other Type) bool {
	typedesc, _ := other.(*TypeDesc)
	return typedesc != nil && t.base.Equals(typedesc.base)
}

func (t *TypeDesc) Underlying() Type { return t }

func (t *TypeDesc) String() string { return fmt.Sprintf("typedesc(%s)", t.base) }

func (t *TypeDesc) Base() Type { return t.base }

func IsTypeDesc(t Type) bool {
	_, ok := t.(*TypeDesc)
	return ok
}

func SkipTypeDesc(t Type) Type {
	if typedesc, ok := t.(*TypeDesc); ok {
		return SkipTypeDesc(typedesc.base)
	}

	return t
}
