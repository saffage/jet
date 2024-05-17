package types

type Ref struct {
	base Type
}

func NewRef(t Type) *Ref {
	if IsTypeDesc(t) || IsUntyped(t) {
		panic("references to meta type is not allowed")
	}
	return &Ref{base: t}
}

func (t *Ref) Equals(target Type) bool {
	if t2 := AsPrimitive(target); t2 != nil {
		return t2.kind == KindPointer || t2.kind == KindAny
	}
	if t2 := AsRef(target); t2 != nil {
		return t.base.Equals(t2.base)
	}
	return false
}

func (t *Ref) Underlying() Type { return t }

func (t *Ref) String() string { return "*" + t.base.String() }

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
