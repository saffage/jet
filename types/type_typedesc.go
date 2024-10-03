package types

import "fmt"

type TypeDesc struct {
	base Type
}

func NewTypeDesc(t Type) *TypeDesc {
	if t == nil {
		panic("unreachable")
	}

	if typedesc, _ := t.(*TypeDesc); typedesc != nil {
		return typedesc
	}

	return &TypeDesc{base: t}
}

func (t *TypeDesc) Equal(other Type) bool {
	if t2 := As[*TypeDesc](other); t2 != nil {
		return t.base.Equal(t2.base)
	}

	return false
}

func (t *TypeDesc) Underlying() Type { return t }

func (t *TypeDesc) String() string { return fmt.Sprintf("type %s", t.base) }

func (t *TypeDesc) Base() Type { return t.base }

func SkipTypeDesc(t Type) Type {
	if typedesc, ok := t.(*TypeDesc); ok {
		return SkipTypeDesc(typedesc.base)
	}

	return t
}
