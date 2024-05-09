package types

import "fmt"

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

func (a *Alias) Equals(other Type) bool { return a.actual.Equals(removeAlias(other)) }

func (a *Alias) Underlying() Type { return a.actual.Underlying() }

func (a *Alias) String() string {
	if IsPrimitive(a.base) {
		return a.name
	}

	return fmt.Sprintf("%s aka %s", a.name, a.base)
}

func SkipAlias(t Type) Type {
	if a, _ := t.(*Alias); a != nil {
		return a.actual
	}

	return t
}

func removeAlias(t0 Type) Type {
	if t0 == nil {
		return nil
	}

	t := t0
	a, ok := t0.(*Alias)

	for ok && a != nil {
		t = a.actual
		a, ok = t.(*Alias)
	}

	if t == nil {
		panic("invalid alias")
	}

	return t
}
