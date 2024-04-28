package types

type Unknown struct{}

func (t Unknown) Underlying() Type       { return t }
func (t Unknown) String() string         { return "unknown" }
func (t Unknown) Equals(other Type) bool { return IsUnknown(other) }

func IsUnknown(t Type) bool {
	return isOfType[Unknown](t)
}
