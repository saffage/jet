package types

type Type interface {
	Equals(Type) bool
	Underlying() Type

	// A human readable representation. For more correct output context is required.
	String() string
}

// func (t *Type) Equals(other *Type) bool {
// 	if t == other || (!IsUntyped(t) && t == SkipUntyped(other)) {
// 		return true
// 	}

// 	if t == nil || other == nil {
// 		return false
// 	}

// 	if t.info != nil && other.info != nil {
// 		return t.info == other.info || t.info.Equals(other)
// 	}

// 	return false
// }
