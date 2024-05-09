package types

import "github.com/saffage/jet/constant"

type Primitive struct {
	kind PrimitiveKind
}

func (t *Primitive) Equals(other Type) bool {
	if t2 := AsPrimitive(other); t2 != nil {
		return t.kind == KindAny || (t2 != nil && (t.kind == t2.kind || t.kind == SkipUntyped(t2).(*Primitive).kind))
	}

	return false
}

func (p *Primitive) Underlying() Type { return p }

func (p *Primitive) String() string { return p.kind.String() }

func (p *Primitive) Kind() PrimitiveKind { return p.kind }

func IsPrimitive(t Type) bool {
	return AsPrimitive(t) != nil
}

func AsPrimitive(t Type) *Primitive {
	if primitive, _ := t.Underlying().(*Primitive); primitive != nil {
		return primitive
	}
	return nil
}

func SkipUntyped(t Type) Type {
	if p, _ := t.(*Primitive); p != nil {
		switch p.kind {
		case KindUntypedBool:
			return Bool

		case KindUntypedInt:
			return I32

		// case UntypedFloat, UntypedString:
		// 	panic("not implemented")
		}
	}
	return t
}

func IsUntyped(t Type) bool {
	if p, _ := t.(*Primitive); p != nil {
		switch p.kind {
		case KindUntypedBool, KindUntypedInt, KindUntypedFloat, KindUntypedString:
			return true
		}
	}

	return false
}

func FromConstant(value constant.Value) Type {
	switch value.Kind() {
	case constant.Bool:
		return UntypedBool

	case constant.Int:
		return UntypedInt

	case constant.Float:
		return UntypedFloat

	case constant.String:
		return UntypedString

	default:
		panic("unreachable")
	}
}

//go:generate stringer -type=PrimitiveKind -linecomment -output=primitive_kind_string.go
type PrimitiveKind byte

const (
	UnknownPrimitiveKind PrimitiveKind = iota

	KindUntypedBool   // untyped bool
	KindUntypedInt    // untyped int
	KindUntypedFloat  // untyped float
	KindUntypedString // untyped string

	KindBool // bool
	KindI32  // i32

	// Meta types.

	KindAny         // any
	KindAnyTypeDesc // typedesc
)

var (
	UntypedBool   = &Primitive{KindUntypedBool}
	UntypedInt    = &Primitive{KindUntypedInt}
	UntypedFloat  = &Primitive{KindUntypedFloat}
	UntypedString = &Primitive{KindUntypedString}

	Bool = &Primitive{KindBool}
	I32  = &Primitive{KindI32}

	Any         = &Primitive{KindAny}
	AnyTypeDesc = &Primitive{KindAnyTypeDesc}
)
