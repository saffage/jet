package types

func (t *Named) Equal(target Type) bool {
	if other, ok := target.(*Named); ok && other != nil {
		return t == other
	}
	return false
}

func (t *Alias) Equal(target Type) bool {
	if other, ok := target.(*Alias); ok && other != nil {
		return t == other
	}
	return false
}

func (t *Fn) Equal(expected Type) bool {
	if expected, _ := As[*Fn](expected); expected != nil {
		return (t.variadic != nil && t.variadic.Equal(expected.variadic) ||
			t.variadic == nil && expected.variadic == nil) &&
			t.result.Equal(expected.result) && t.params.Equal(expected.params)
	}
	return false
}

func (t *TypeVariable) Equal(target Type) bool {
	if other, ok := target.(*TypeVariable); ok {
		// We can't compare type variables by their ids (as they can be used in
		// different inference context), so we need to compare their
		// instances instead.
		return t == other
	}
	return false
}

func (m *Module) Equal(target Type) bool {
	return true
}
