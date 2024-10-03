package types

type Alias struct {
	base   Type // Type of the expression that was used in alias declaration (can be an alias).
	actual Type // Actual type of the type alias (can't be an alias).
	name   string
}

func NewAlias(t Type, name string) *Alias {
	if t == nil {
		panic("alias to unknown must be used in built-in types only")
	}
	return &Alias{
		base:   t,
		actual: removeAlias(t),
		name:   name,
	}
}

func (t *Alias) Equal(other Type) bool {
	return t.actual.Equal(SkipAlias(other))
}

func (t *Alias) Underlying() Type {
	return t.actual.Underlying()
}

func (t *Alias) String() string {
	if IsTypedPrimitive(t.base) {
		return t.name
	}
	return t.name + " aka " + t.base.String()
}

func SkipAlias(t Type) Type {
	if a, _ := t.(*Alias); a != nil {
		// What if `a.actual` is nil?
		return a.actual
	}
	return t
}

func removeAlias(t0 Type) Type {
	if t0 == nil {
		return nil
	}

	t := t0
	a, ok := t.(*Alias)

	for ok && a != nil {
		t = a.actual
		a, ok = t.(*Alias)
	}

	if t == nil {
		panic("broken alias")
	}

	return t
}
