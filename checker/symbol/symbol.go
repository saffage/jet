package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
)

type Symbol interface {
	// Scope where a symbol was defined.
	Owner() Scope

	// Symbol ID.
	ID() ID

	// Type of a symbol.
	Type() types.Type

	// Identifier or name of a symbol.
	Name() string

	// Identifier node.
	NameNode() *ast.Ident

	// Related AST node.
	Node() ast.Node

	setType(types.Type)
}

type TypeChecker interface {
	Symbol

	// Returns a type of the `expr` or error if `expr` can't have a type.
	TypeOf(expr ast.Node) (types.Type, error)
}

type ID uint64

type base struct {
	owner Scope
	id    ID
	type_ types.Type
	name  *ast.Ident
	node  ast.Node
}

func (sym *base) Owner() Scope         { return sym.owner }
func (sym *base) ID() ID               { return sym.id }
func (sym *base) Type() types.Type     { return sym.type_ }
func (sym *base) Name() string         { return sym.name.Name }
func (sym *base) NameNode() *ast.Ident { return sym.name }
func (sym *base) Node() ast.Node       { return sym.node }

func (sym *base) setType(type_ types.Type) { sym.type_ = type_ }
