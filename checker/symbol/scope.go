package symbol

type Scope interface {
	// Returns the scope where the current scope was defined,
	// or nil if the current scope has no parent (the parent is `global`).
	// Usually the lack of a parent means that the module is the root
	// of the project.
	Parent() Scope

	// If a symbol with the specified name is already defined,
	// returns a reference to it, otherwise returns nil.
	Define(symbol Symbol) Symbol

	// Look up the specified symbol in the scope and return a reference to it,
	// or nil if such symbol is not defined or not accessible.
	// The lookup also involves all symbols used by this scope.
	Resolve(name string) Symbol

	// Look up the specified symbol in this scope and return a reference to it,
	// or nil if such a symbol is not defined or accessible.
	ResolveMember(name string) Symbol

	// Returns all symbols that was defined in this scope.
	Symbols() []Symbol
}

// Imports all symbols defined in another scope into the current scope.
// If the other scope is a module, this module must be completed.
//
// If any of the symbols from the other scope are already defined in the
// current scope, an error will occur.
func Use(current, other Scope) {
	if module, ok := other.(*Module); ok {
		if !module.IsCompleted() {
			panic(NewErrorf(nil, "module '%s' is not completed", module.Name()))
		}
	}

	errors := []Error{}

	for _, sym := range other.Symbols() {
		definedSym := current.Define(sym)

		if definedSym != nil {
			// TODO point to the `using` statement.
			err := NewErrorf(nil, "member '%s' is already defined in this scope", sym.Name())
			err.Notes = append(err.Notes, NewErrorf(definedSym.Ident(), "symbol defined here"))
			errors = append(errors, err)
		}
	}

	panic(errors)
}
