package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Symbol interface {
	Owner() *Scope    // Scope where a symbol was defined.
	Type() types.Type // Type of a symbol.
	Name() string     // Identifier or name of a symbol.
	Ident() ast.Ident // Identifier node.
	Node() ast.Node   // Related AST node.
}
