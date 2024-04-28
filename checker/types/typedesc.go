package types

type TypeDesc struct{ Type Type }

func (t TypeDesc) Underlying() Type       { return t.Type }
func (t TypeDesc) String() string         { return "typedesc(" + t.Type.String() + ")" }
func (t TypeDesc) Equals(other Type) bool { return false }

func IsTypeDesc(t Type) bool {
	return isOfType[TypeDesc](t)
}

func UnwrapTypeDesc(t Type) Type {
	if IsTypeDesc(t) {
		return t.Underlying()
	}

	return t
}
