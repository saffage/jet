package types

import (
	"strings"
)

func As[T Type](t Type) (T, bool) {
	if t != nil {
		if t, ok := t.(T); ok {
			return t, true
		}
		if t, ok := SkipAlias(t).(T); ok {
			return t, true
		}
	}
	var zero T
	// if IsAtom(zero) {
	// 	if p, _ := t.(*Parameter); p != nil {
	// 		return zero, p.Equal(zero)
	// 	}
	// }
	return zero, false
}

func Is[T Type](t Type) bool {
	_, ok := As[T](t)
	return ok
}

func Render(t Type) string {
	buf := strings.Builder{}

	if t == nil || !t.Render(&buf) {
		buf.WriteString(`"Invalid Type"`)
	}

	return buf.String()
}

func SkipAlias(t Type) Type {
	// if a, _ := t.(*NominalAlias); a != nil {
	// 	// What if `a.actual` is nil?
	// 	debug.Assert(a.actual != nil)
	// 	return SkipAlias(a.actual)
	// }
	return t
}

// func SkipDescriptor(t Type) Type {
// 	if typedesc, ok := t.(Descriptor); ok {
// 		return SkipDescriptor(typedesc.base)
// 	}

// 	return t
// }

// func removeAlias(t0 Type) Descriptor {
// 	t := t0
// 	a, ok := t.(*Alias)

// 	for ok && a != nil {
// 		if !IsResolved(t) {
// 			panic("unresolved type in type alias")
// 		}
// 		t = a.actual
// 		a, ok = t.(*Alias)
// 	}

// 	if !IsResolved(t) {
// 		panic("unresolved type in type alias")
// 	}

// 	desc, ok := As[Descriptor](t)

// 	if !ok {
// 		panic("type is not a type descriptor")
// 	}

// 	return desc
// }

// func Underlying(t Type) Type {
// 	if a, _ := t.(*Distinct); a != nil {
// 		return Underlying(a.actual)
// 	}
// 	return t
// }

// func IsResolved(t Type) bool {
// 	switch t := SkipAlias(t).(type) {
// 	case nil:
// 		// It's not clear if nil is resolved here or not, it depends on the
// 		// context.
// 		return true

// 	case Unit, Never, Int, Float, String:
// 		// Atom types are always resolved.
// 		return true

// 	// case *Parameter:
// 	// 	// Not sure about it.
// 	// 	return true

// 	case module:
// 		// Not sure about it.
// 		return true

// 	case Descriptor:
// 		return IsResolved(t.base)

// 	case *Custom:
// 		for _, field := range t.fields {
// 			if !IsResolved(field.T) {
// 				return false
// 			}
// 		}

// 		for _, variant := range t.variants {
// 			if slices.IndexFunc(variant.Params, IsResolved) >= 0 {
// 				return false
// 			}
// 		}

// 		return true

// 	case *Fn:
// 		for _, param := range t.params {
// 			if !IsResolved(param) {
// 				return false
// 			}
// 		}

// 		if !IsResolved(t.result) {
// 			return false
// 		}

// 		if t.variadic != nil {
// 			return IsResolved(t.variadic)
// 		}

// 		return true

// 	default:
// 		panic("unreachable")
// 	}
// }

// func IsAtom(t Type) bool {
// 	switch t.(type) {
// 	case Unit, Never, Int, Float, String:
// 		return true

// 	default:
// 		return false
// 	}
// }

/*
// Trying to turn a type to the typed analog.
//
// If target type is not provided, then result is never nil.
//
// If target type is provided, then type will be checked for equality with
// target type, and if its not, then nil will be returned.
func IntoTyped(v *Value, target ...Type) Type {
	expected := Type(nil)

	if len(target) != 0 {
		expected = SkipAlias(target[0])
		debug.Assert(expected != nil, "argument is nil")
	}

	t := SkipAlias(v.T)

	if t == nil || expected != nil && !t.Equal(expected) {
		return nil
	}

	return t
}
*/

/* func FromConstant(value ConstantValue) Type {
	switch value.Kind() {
	case ConstantInt:
		return IntType

	case ConstantFloat:
		return FloatType

	case ConstantString:
		return StringType

	default:
		panic("unreachable")
	}
} */
