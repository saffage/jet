package types

type Primitive struct {
	Kind PrimitiveKind
}

type PrimitiveKind byte

const (
	Invalid PrimitiveKind = iota
	UntypedInt
	UntypedFloat
	UntypedString

	I32
)

func (t *Primitive) Underlying() Type {
	return t
}

func (t *Primitive) SameType(other Type) bool {
	if primitive, ok := other.(*Primitive); ok {
		return t.Kind == primitive.Kind
	}
	return false
}

func (t *Primitive) String() string {
	switch t.Kind {
	case UntypedInt:
		return "untyped int"

	case UntypedFloat:
		return "untyped float"

	case UntypedString:
		return "untyped string"

	case I32:
		return "i32"

	default:
		panic("unknown type")
	}
}

func (t *Primitive) IsUntyped() bool {
	return UntypedInt <= t.Kind && t.Kind <= UntypedString
}
