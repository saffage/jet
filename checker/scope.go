package checker

import (
	"github.com/saffage/jet/ast"
)

var Global = NewScope(nil, "global")

type Scope struct {
	name     string
	parentID int
	parent   *Scope
	children []*Scope
	defers   []*ast.Defer
	symbols  map[string]Symbol
}

func NewScope(parent *Scope, name string) *Scope {
	scope := &Scope{name: name, parent: parent}

	if parent != nil && parent.parent != nil {
		scope.parentID = len(parent.children)
		parent.children = append(parent.children, scope)
	}

	return scope
}

// Returns a name of the scope. Used in code generator.
//
// Name can be next:
//   - module <name>
//   - func <name>
//   - struct <name>
//   - enum <name>
//   - block
//   - global
func (scope *Scope) Name() string {
	return scope.name
}

func (scope *Scope) Parent() *Scope {
	return scope.parent
}

func (scope *Scope) Children() []*Scope {
	return scope.children
}

func (scope *Scope) Defers() []*ast.Defer {
	return scope.defers
}

// Defines a new symbol in the scope. If a symbol with the same
// name is already defined in this scope, it will return error.
func (scope *Scope) Define(symbol Symbol) Symbol {
	if symbol == nil {
		// Scope should not contain a nil symbols.
		panic("attempt to define nil symbol")
	}

	if defined := scope.LookupLocal(symbol.Name()); defined != nil {
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
	if member := scope.LookupLocal(name); member != nil {
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
func (scope *Scope) LookupLocal(name string) Symbol {
	if scope.symbols == nil {
		return nil
	}

	return scope.symbols[name]
}

func errorAlreadyDefined(ident, previous *ast.Ident) *Error {
	err := newErrorf(ident, "name '%s' is already defined in this scope", ident.Name)

	if previous != nil && previous.Start.Line > 0 {
		err.Notes = append(err.Notes, &Error{
			Message: "previous declaration was here",
			Node:    previous,
		})
	}

	return err
}
