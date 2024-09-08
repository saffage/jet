package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Symbol interface {
	Type() types.Type // Type of the symbol.
	Name() string     // Name of the symbol.
	Node() ast.Node   // Related AST node.
	Ident() ast.Ident // Identifier node.
	Owner() *Scope    // Scope where the symbol was defined.
}
