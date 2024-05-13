package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type TypeInfo struct {
	// Constant data.
	Data map[ast.Node]*TypedValue

	// Type of every AST node.
	Types map[ast.Node]*TypedValue

	// Type symbols for every type defined in the module.
	TypeSyms map[types.Type]Symbol

	// Every definition in the entire module.
	Defs map[*ast.Ident]Symbol

	// Every definition in the entire module.
	Uses map[*ast.Ident]Symbol
}

func NewTypeInfo() *TypeInfo {
	return &TypeInfo{
		Data:     make(map[ast.Node]*TypedValue),
		Types:    make(map[ast.Node]*TypedValue),
		TypeSyms: make(map[types.Type]Symbol),
		Defs:     make(map[*ast.Ident]Symbol),
		Uses:     make(map[*ast.Ident]Symbol),
	}
}
