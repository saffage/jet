package types

type moduleType struct{}

func (moduleType) Equal(expected Type) bool { return false }
func (moduleType) Underlying() Type         { return nil }
func (moduleType) String() string           { return "module" }
