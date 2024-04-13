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

	// Сopies all symbols defined in another module to the current scope.
	// The module must be complete.
	Use(module *Module)
}
