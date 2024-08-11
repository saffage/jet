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

func (t *TypeDesc) Equals(other Type) bool {
	if t2 := AsTypeDesc(other); t2 != nil {
		return t.base.Equals(t2.base)
	}
	if t2 := AsPrimitive(other); t2 != nil {
		return t2.kind == KindAnyTypeDesc
	}
	return false
}

func (t *TypeDesc) Underlying() Type { return t }

func (t *TypeDesc) String() string { return fmt.Sprintf("typedesc(%s)", t.base) }

func (t *TypeDesc) Base() Type { return t.base }

func IsTypeDesc(t Type) bool { return AsTypeDesc(t) != nil }

func AsTypeDesc(t Type) *TypeDesc {
	if t != nil {
		if typedesc, _ := t.Underlying().(*TypeDesc); typedesc != nil {
			return typedesc
		}
	}

	return nil
}

func SkipTypeDesc(t Type) Type {
	if typedesc, ok := t.(*TypeDesc); ok {
		return SkipTypeDesc(typedesc.base)
	}

	return t
}
