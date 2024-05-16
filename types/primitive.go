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

func (t *Primitive) IsImplicitlyConvertibleTo(target Type) bool {
	if target != nil {
		switch target := target.Underlying().(type) {
		case *Primitive:
			switch target.kind {
			case KindUntypedBool, KindUntypedInt, KindUntypedFloat, KindUntypedString:
				return t.kind == target.kind

			case KindBool:
				return t.kind == KindUntypedBool ||
					t.kind == KindBool

			case KindI8:
				return t.kind == KindUntypedInt ||
					t.kind == KindI8

			case KindI16:
				return t.kind == KindUntypedInt ||
					t.kind == KindI16 ||
					t.kind == KindI8 ||
					t.kind == KindU8

			case KindI32:
				return t.kind == KindUntypedInt ||
					t.kind == KindI32 ||
					t.kind == KindI16 ||
					t.kind == KindI8 ||
					t.kind == KindU16 ||
					t.kind == KindU8

			case KindI64:
				return t.kind == KindUntypedInt ||
					t.kind == KindI64 ||
					t.kind == KindI32 ||
					t.kind == KindI16 ||
					t.kind == KindI8 ||
					t.kind == KindU32 ||
					t.kind == KindU16 ||
					t.kind == KindU8

			case KindU8:
				return t.kind == KindUntypedInt ||
					t.kind == KindU8

			case KindU16:
				return t.kind == KindUntypedInt ||
					t.kind == KindU16 ||
					t.kind == KindU8

			case KindU32:
				return t.kind == KindUntypedInt ||
					t.kind == KindU32 ||
					t.kind == KindU16 ||
					t.kind == KindU8

			case KindU64:
				return t.kind == KindUntypedInt ||
					t.kind == KindU64 ||
					t.kind == KindU32 ||
					t.kind == KindU16 ||
					t.kind == KindU8

			case KindF32:
				return t.kind == KindUntypedFloat ||
					t.kind == KindF32

			case KindF64:
				return t.kind == KindUntypedFloat ||
					t.kind == KindF64 ||
					t.kind == KindF32

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

			case KindUntypedFloat:
				return F64

			case KindUntypedString:
				return String
			}

		case *Tuple:
			elems := make([]Type, len(t.types))
			for i := range t.types {
				elems[i] = SkipUntyped(t.types[i])
			}
			return NewTuple(elems...)

		case *Array:
			return NewArray(t.size, SkipUntyped(t.elem))
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

		case *Array:
			return IsUntyped(t.elem)
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
	KindI8   // i8
	KindI16  // i16
	KindI32  // i32
	KindI64  // i64
	KindU8   // u8
	KindU16  // u16
	KindU32  // u32
	KindU64  // u64
	KindF32  // f32
	KindF64  // f64

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
	I8   = &Primitive{KindI8}
	I16  = &Primitive{KindI16}
	I32  = &Primitive{KindI32}
	I64  = &Primitive{KindI64}
	U8   = &Primitive{KindU8}
	U16  = &Primitive{KindU16}
	U32  = &Primitive{KindU32}
	U64  = &Primitive{KindU64}
	F32  = &Primitive{KindF32}
	F64  = &Primitive{KindF64}

	Any         = &Primitive{KindAny}
	AnyTypeDesc = &Primitive{KindAnyTypeDesc}

	String = &Struct{fields: []StructField{
		{"len", I32},
		{"ptr", &Ref{base: U8}},
	}}
)
