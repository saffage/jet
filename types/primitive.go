package types

import "github.com/saffage/jet/constant"

type Primitive struct {
	kind PrimitiveKind
}

func (p *Primitive) Equals(other Type) bool {
	p2, ok := other.Underlying().(*Primitive)
	return p.kind == Any || (ok && p.kind == p2.kind || p.kind == SkipUntyped(p2).(*Primitive).kind)
}

func (p *Primitive) Underlying() Type { return p }

func (p *Primitive) String() string { return p.kind.String() }

func (p *Primitive) Kind() PrimitiveKind { return p.kind }

func IsPrimitive(t Type) bool {
	_, ok := t.(*Primitive)
	return ok
}

func SkipUntyped(t Type) Type {
	if p, ok := t.(*Primitive); ok {
		switch p.kind {
		case UntypedBool:
			return Primitives[Bool]

		case UntypedInt:
			return Primitives[I32]

		// case UntypedFloat, UntypedString:
		// 	panic("not implemented")
		}
	}

	return t
}

func IsUntyped(t Type) bool {
	if p, ok := t.(*Primitive); ok {
		switch p.kind {
		case UntypedBool, UntypedInt, UntypedFloat, UntypedString:
			return true
		}
	}

	return false
}

func NewTypeFromConstant(value constant.Value) Type {
	switch value.Kind() {
	case constant.Bool:
		return Primitives[UntypedBool]

	case constant.Int:
		return Primitives[UntypedInt]

	case constant.Float, constant.String, constant.Expression:
		panic("not implemented")

	default:
		panic("unreachable")
	}
}

//go:generate stringer -type=PrimitiveKind -linecomment -output=primitive_kind_string.go
type PrimitiveKind byte

const (
	UnknownPrimitive PrimitiveKind = iota

	UntypedBool   // untyped bool
	UntypedInt    // untyped int
	UntypedFloat  // untyped float
	UntypedString // untyped string

	Bool // bool
	I32  // i32

	// Meta types.

	Any         // any
	AnyTypeDesc // typedesc
)

var Primitives = [...]*Primitive{
	UnknownPrimitive: {UnknownPrimitive},
	UntypedBool:      {UntypedBool},
	UntypedInt:       {UntypedInt},
	UntypedFloat:     {UntypedFloat},
	UntypedString:    {UntypedString},

	Bool: {Bool},
	I32:  {I32},

	Any:         {Any},
	AnyTypeDesc: {AnyTypeDesc},
}
