package types

import "github.com/saffage/jet/constant"

type Primitive struct {
	kind PrimitiveKind
}

func (t *Primitive) Equals(other Type) bool {
	return t.IsImplicitlyConvertibleTo(other)
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
	if t != nil {
		switch t := SkipAlias(t).(type) {
		case *Primitive:
			switch t.kind {
			case KindUntypedBool:
				return Bool

			case KindUntypedInt:
				return I32

			case KindUntypedString:
				return String
			}

		case *Tuple:
			elems := make([]Type, len(t.types))
			for i := range t.types {
				elems[i] = SkipUntyped(t.types[i])
			}
			return NewTuple(elems...)
		}
	}
	return t
}

func IsUntyped(t Type) bool {
	if t != nil {
		switch t := t.Underlying().(type) {
		case *Primitive:
			switch t.kind {
			case KindUntypedBool, KindUntypedInt, KindUntypedFloat, KindUntypedString:
				return true
			}

		case *Tuple:
			for _, elem := range t.types {
				if IsUntyped(elem) {
					return true
				}
			}
		}
	}
	return false
}

func (t *Primitive) IsImplicitlyConvertibleTo(target Type) bool {
	if target != nil {
		switch target := target.Underlying().(type) {
		case *Primitive:
			switch target.kind {
			case KindUntypedBool, KindUntypedInt, KindUntypedFloat, KindUntypedString:
				return t.kind == target.kind

			case KindBool:
				return t.kind == KindUntypedBool || t.kind == KindBool

			case KindI32:
				return t.kind == KindUntypedInt || t.kind == KindI32 || t.kind == KindU8

			case KindU8:
				return t.kind == KindUntypedInt || t.kind == KindU8

			case KindAnyTypeDesc:
				return false

			case KindAny:
				return true

			default:
				panic("unreachable")
			}

		case *Struct:
			if target == String {
				return t.kind == KindUntypedString
			}
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
	KindU8   // u8

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
	U8   = &Primitive{KindU8}

	Any         = &Primitive{KindAny}
	AnyTypeDesc = &Primitive{KindAnyTypeDesc}

	String = &Struct{fields: map[string]Type{
		"len": I32,
		"ptr": &Ref{base: U8},
	}}
)
