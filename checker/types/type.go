package types

type Type interface {
	// Returns the same type with the current type if
	// has no underlying type.
	Underlying() Type

	// A human readable representation.
	String() string

	Equals(Type) bool

	implType()
}

func (UntypedBool) implType()   {}
func (UntypedInt) implType()    {}
func (UntypedFloat) implType()  {}
func (UntypedString) implType() {}
func (Bool) implType()          {}
func (I32) implType()           {}
func (Unknown) implType()       {}
func (TypeDesc) implType()      {}
func (Any) implType()           {}
