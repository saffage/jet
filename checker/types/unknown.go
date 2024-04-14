package types

type Unknown struct{}

func (t Unknown) Underlying() Type {
	return t
}

func (t Unknown) String() string {
	return "unknown"
}

func (t Unknown) SameType(other Type) bool {
	return IsUnknown(other)
}
