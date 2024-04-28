package types

type Any struct{}

func (t Any) Underlying() Type       { return t }
func (t Any) String() string         { return "any" }
func (t Any) Equals(other Type) bool { return true }
