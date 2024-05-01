package types

import "github.com/saffage/jet/constant"

type (
	UntypedBool   struct{}
	UntypedInt    struct{}
	UntypedFloat  struct{}
	UntypedString struct{}

	Bool struct{}
	I32  struct{}
)

func (t UntypedBool) Underlying() Type   { return t }
func (t UntypedBool) String() string     { return "untyped bool" }
func (t UntypedBool) Equals(x Type) bool { return isOfType[UntypedBool](x) }

func (t UntypedInt) Underlying() Type   { return t }
func (t UntypedInt) String() string     { return "untyped int" }
func (t UntypedInt) Equals(x Type) bool { return isOfType[UntypedInt](x) }

func (t UntypedFloat) Underlying() Type   { return t }
func (t UntypedFloat) String() string     { return "untyped float" }
func (t UntypedFloat) Equals(x Type) bool { return isOfType[UntypedFloat](x) }

func (t UntypedString) Underlying() Type   { return t }
func (t UntypedString) String() string     { return "untyped string" }
func (t UntypedString) Equals(x Type) bool { return isOfType[UntypedString](x) }

func (t Bool) Underlying() Type   { return t }
func (t Bool) String() string     { return "bool" }
func (t Bool) Equals(x Type) bool { return isOfType[Bool](x) || isOfType[UntypedBool](x) }

func (t I32) Underlying() Type   { return t }
func (t I32) String() string     { return "i32" }
func (t I32) Equals(x Type) bool { return isOfType[I32](x) || isOfType[UntypedInt](x) }

func FromConstant(kind constant.Kind) Type {
	switch kind {
	case constant.Bool:
		return UntypedBool{}

	case constant.Int:
		return UntypedInt{}

	case constant.Float:
		return UntypedFloat{}

	case constant.String:
		return UntypedString{}

	case constant.Expression:
		panic("not implemented")

	default:
		panic("unreachable")
	}
}

func IsUntyped(t Type) bool {
	switch t.(type) {
	case UntypedBool, UntypedInt, UntypedFloat, UntypedString:
		return true

	default:
		return false
	}
}

func TypedFromUntyped(t Type) Type {
	switch t.(type) {
	case UntypedBool:
		return Bool{}

	case UntypedInt:
		return I32{}

	case UntypedFloat, UntypedString:
		panic("not implemented")

	default:
		return t
	}
}

// NOTE use only for primitive types!
func isOfType[T Type](x Type) bool {
	_, ok := x.(T)
	return ok
}
