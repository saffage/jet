package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
)

type Symbol interface {
	// Symbol ID.
	ID() ID

	// Scope where a symbol was defined.
	Owner() Scope

	// Type of a symbol.
	Type() types.Type

	// Identifier or name of a symbol.
	Name() string

	// Identifier node.
	Ident() *ast.Ident

	// Related AST node.
	Node() ast.Node

	setType(types.Type)
}

type ID uint64

var currentID = ID(0)

func nextID() ID {
	currentID++
	return currentID
}
