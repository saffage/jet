package types

type Type interface {
	Equals(Type) bool
	Underlying() Type

	// A human readable representation.
	// For more correct output context is required.
	String() string
}
