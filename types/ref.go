package types

type Ref struct {
	base Type
}

func NewRef(t Type) *Ref {
	if IsTypeDesc(t) {
		panic("references to meta type is not allowed")
	}

	return &Ref{base: t}
}

func (t *Ref) Equals(other Type) bool {
	if otherRef, ok := other.Underlying().(*Ref); ok {
		return t.base.Equals(otherRef.base)
	}

	return false
}

func (t *Ref) Underlying() Type { return t }

func (t *Ref) String() string { return "&" + t.base.String() }

func (t *Ref) Base() Type { return t.base }

func IsRef(t Type) bool { return AsRef(t) != nil }

func AsRef(t Type) *Ref {
	if t != nil {
		if ref, _ := t.Underlying().(*Ref); ref != nil {
			return ref
		}
	}

	return nil
}
