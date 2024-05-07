package checker

import "github.com/saffage/jet/ast"

type Scope struct {
	parent  *Scope
	symbols map[string]Symbol
}

func NewScope(parent *Scope) *Scope {
	return &Scope{parent, nil}
}

// Returns the scope in which the current scope was defined,
// or nil if the current scope has no parent.
func (scope *Scope) Parent() *Scope {
	return scope.parent
}

// Defines a symbol in the scope. If a symbol with the same
// name is already defined in this scope, it will return it
// without overriding it.
//
// NOTE: the symbol will be defined even if a symbol with the
// same name is defined in the parent scope.
func (scope *Scope) Define(symbol Symbol) (defined Symbol) {
	// Scope should not contain a nil symbols.
	if symbol == nil {
		panic("attempt to define nil symbol")
	}

	if defined := scope.Member(symbol.Name()); defined != nil {
		return defined
	}

	if scope.symbols == nil {
		scope.symbols = make(map[string]Symbol)
	}

	scope.symbols[symbol.Name()] = symbol
	return nil
}

// Searches for the specified symbol by name in the context of
// the specified scope and returns it, or nil if such symbol
// is undefined or unavailable (private).
func (scope *Scope) Lookup(name string) (Symbol, *Scope) {
	if member := scope.Member(name); member != nil {
		return member, scope
	}

	if scope.parent != nil {
		return scope.parent.Lookup(name)
	}

	return nil, nil
}

// Searches for the specified symbol by name in the specified
// scope and returns it, or nil if no such symbol is defined
// in the current scope.
func (scope *Scope) Member(name string) Symbol {
	if symbol, ok := scope.symbols[name]; ok {
		return symbol
	}

	return nil
}

func errorAlreadyDefined(ident, previous *ast.Ident) *Error {
	err := NewErrorf(ident, "name '%s' is already defined in this scope", ident.Name)

	if previous != nil && previous.Start.Line > 0 {
		err.Notes = []*Error{
			NewError(previous, "previous declaration was here"),
		}
	}

	return err
}
