package types

type Type interface {
	// Returns the same type with the current type if
	// has no underlying type.
	Underlying() Type

	// A human readable representation.
	String() string

	SameType(Type) bool
}

func IsUnknown(t Type) bool {
	_, ok := t.(*Unknown)
	return ok
}
