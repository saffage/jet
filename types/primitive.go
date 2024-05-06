package types

import "github.com/saffage/jet/constant"

type Primitive struct {
	kind PrimitiveKind
}

func (p *Primitive) Equals(other Type) bool {
	p2, _ := other.Underlying().(*Primitive)
	return p.kind == Any || (p2 != nil && (p.kind == p2.kind || p.kind == SkipUntyped(p2).(*Primitive).kind))
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
	if p, _ := t.(*Primitive); p != nil {
		switch p.kind {
		case UntypedBool, UntypedInt, UntypedFloat, UntypedString:
			return true
		}
	}

	return false
}

func FromConstant(value constant.Value) Type {
	switch value.Kind() {
	case constant.Bool:
		return Primitives[UntypedBool]

	case constant.Int:
		return Primitives[UntypedInt]

	case constant.Float, constant.String:
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
