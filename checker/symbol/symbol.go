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

type base struct {
	id    ID
	owner Scope
	type_ types.Type
	name  *ast.Ident
	node  ast.Node
}

func (sym *base) ID() ID            { return sym.id }
func (sym *base) Owner() Scope      { return sym.owner }
func (sym *base) Type() types.Type  { return sym.type_ }
func (sym *base) Name() string      { return sym.name.Name }
func (sym *base) Ident() *ast.Ident { return sym.name }
func (sym *base) Node() ast.Node    { return sym.node }

func (sym *base) setType(type_ types.Type) { sym.type_ = type_ }
