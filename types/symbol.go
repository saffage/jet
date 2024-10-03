package types

import "github.com/saffage/jet/ast"

type Symbol interface {
	Type() Type       // Type of the symbol.
	Name() string     // Name of the symbol.
	Node() ast.Node   // Related AST node.
	Ident() ast.Ident // Related identifier AST node.
	Owner() *Env      // Scope where the symbol was defined.
}
