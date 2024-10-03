package types

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
)

type Type interface {
	Equal(expected Type) bool
	Underlying() Type
	String() string
}

type comparableType interface {
	Type
	comparable
}

func Is[T comparableType](t Type) bool {
	var zero T
	return As[T](t) != zero
}

func As[T comparableType](t Type) T {
	var zero T
	if t != nil {
		if t, ok := t.(T); ok && t != zero {
			return t
		}
		if t, ok := t.Underlying().(T); ok && t != zero {
			return t
		}
	}
	return zero
}

//-----------------------------------------------
// Types of untyped expressions

type (
	UntypedInt    struct{}
	UntypedFloat  struct{}
	UntypedString struct{}
)

var (
	UntypedIntType    Type = UntypedInt{}
	UntypedFloatType  Type = UntypedFloat{}
	UntypedStringType Type = UntypedString{}
)

func (UntypedInt) String() string         { return "untyped Int" }
func (UntypedInt) Underlying() Type       { return UntypedIntType }
func (UntypedInt) Equal(target Type) bool { return Is[UntypedInt](SkipAlias(target)) }

func (UntypedFloat) String() string         { return "untyped Float" }
func (UntypedFloat) Underlying() Type       { return UntypedFloatType }
func (UntypedFloat) Equal(target Type) bool { return Is[UntypedFloat](SkipAlias(target)) }

func (UntypedString) String() string         { return "untyped String" }
func (UntypedString) Underlying() Type       { return UntypedStringType }
func (UntypedString) Equal(target Type) bool { return Is[UntypedString](SkipAlias(target)) }

//-----------------------------------------------
// Primitive types

type (
	None   struct{}
	Never  struct{}
	Int    struct{}
	Float  struct{}
	String struct{}
)

var (
	NoneType   Type = None{}
	NeverType  Type = Never{}
	IntType    Type = Int{}
	FloatType  Type = Float{}
	StringType Type = String{}
)

func (None) String() string         { return "None" }
func (None) Underlying() Type       { return NoneType }
func (None) Equal(target Type) bool { return Is[None](SkipAlias(target)) }

func (Never) String() string         { return "Never" }
func (Never) Underlying() Type       { return NeverType }
func (Never) Equal(target Type) bool { return true }

func (Int) String() string   { return "Int" }
func (Int) Underlying() Type { return IntType }
func (Int) Equal(target Type) bool {
	return Is[Int](SkipAlias(target)) || Is[UntypedInt](SkipAlias(target))
}

func (Float) String() string   { return "Float" }
func (Float) Underlying() Type { return FloatType }
func (Float) Equal(target Type) bool {
	return Is[Float](SkipAlias(target)) || Is[UntypedFloat](SkipAlias(target))
}

func (String) String() string   { return "String" }
func (String) Underlying() Type { return StringType }
func (String) Equal(target Type) bool {
	return Is[String](SkipAlias(target)) || Is[UntypedString](SkipAlias(target))
}

//-----------------------------------------------
// Util functions

func FromConstant(value constant.Value) Type {
	switch value.Kind() {
	case constant.Int:
		return UntypedIntType

	case constant.Float:
		return UntypedFloatType

	case constant.String:
		return UntypedStringType

	default:
		panic("unreachable")
	}
}

func FromAst(node ast.Node) Type {
	literal, _ := node.(*ast.Literal)

	if literal == nil {
		panic("unreachable")
	}

	switch literal.Kind {
	case ast.IntLiteral:
		return UntypedIntType

	case ast.FloatLiteral:
		return UntypedFloatType

	case ast.StringLiteral:
		return UntypedStringType

	default:
		panic("unreachable")
	}
}

func IsPrimitive(t Type) bool {
	switch t.(type) {
	case None, Never, Int, Float, String, UntypedInt, UntypedFloat, UntypedString:
		return true

	default:
		return false
	}
}

func IsTypedPrimitive(t Type) bool {
	switch t.(type) {
	case None, Never, Int, Float, String:
		return true

	default:
		return false
	}
}

func IsUntyped(t Type) bool {
	return IntoTyped(t) != t
}

// Trying to turn a type to the typed analog.
//
// If target type if not provided, then result is never nil.
//
// If target type is provided, then type will be checked for equality with
// target type, and if its not, then nil will be returned.
func IntoTyped(t Type, target ...Type) Type {
	if len(target) > 1 {
		panic("invalid arguments count, expected target len < 2")
	}

	t = SkipAlias(t)

	if t == nil {
		return nil
	}

	expected := Type(nil)

	if len(target) > 0 {
		if target[0] == nil {
			panic("argument is nil")
		}
		if IsUntyped(target[0]) {
			panic("target type shouldn't be untyped")
		}
		expected = SkipAlias(target[0])
	}

	switch t := t.(type) {
	case UntypedInt:
		if expected != nil && !IntType.Equal(expected) {
			return nil
		}
		return IntType

	case UntypedFloat:
		if expected != nil && !FloatType.Equal(expected) {
			return nil
		}
		return FloatType

	case UntypedString:
		if expected != nil && !StringType.Equal(expected) {
			return nil
		}
		return StringType

	// case *Tuple:
	// 	var expected, _ = expected.(*Tuple)
	// 	var types = make([]Type, 0, len(t.types))

	// 	if expected != nil && len(t.types) != len(expected.types) {
	// 		return nil
	// 	}

	// 	for i := range t.types {
	// 		var typed Type

	// 		if expected != nil {
	// 			typed = AsTyped(t.types[i], expected.types[i])

	// 			if typed == nil {
	// 				return nil
	// 			}
	// 		}

	// 		types = append(types, typed)
	// 	}

	// 	t = &Tuple{types: types}
	// 	if expected != nil {
	// 		assert(t.Equal(expected))
	// 	}
	// 	return t

	case *Array:
		var elem Type
		var expected, _ = expected.(*Array)

		if expected != nil && (expected.size != t.size) {
			elem = IntoTyped(t.elem, expected.elem)

			if elem == nil {
				return nil
			}
		}

		t = &Array{elem: elem, size: t.size}
		if expected != nil {
			assert(t.Equal(expected))
		}
		return t

	default:
		if expected != nil && !t.Equal(expected) {
			return nil
		}
		return t
	}
}
