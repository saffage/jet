package types

import (
	ordered "github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"github.com/saffage/jet/ast"
)

type TypeInfo struct {
	// Every definition in the entire module.
	Defs *ordered.Map[ast.Ident, Symbol]

	// Every usage in the entire module.
	Uses map[ast.Ident]Symbol

	// Value of every AST node.
	Values map[ast.Node]*Value

	TypeOwners map[Type]Symbol
}

func newTypeInfo() *TypeInfo {
	return &TypeInfo{
		Defs:       ordered.New[ast.Ident, Symbol](),
		Uses:       make(map[ast.Ident]Symbol),
		Values:     make(map[ast.Node]*Value),
		TypeOwners: make(map[Type]Symbol),
	}
}

func (ti *TypeInfo) TypeOf(expr ast.Node) Type {
	if expr != nil {
		if tv, ok := ti.Values[expr]; ok && tv != nil {
			return tv.T
		}
	}
	return nil
}

func (ti *TypeInfo) ValueOf(expr ast.Node) *Value {
	if expr != nil {
		if tv, ok := ti.Values[expr]; ok && tv != nil && tv.V != nil {
			return tv
		}
	}
	return nil
}

func (ti *TypeInfo) SymbolOf(ident ast.Ident) Symbol {
	if ident != nil {
		if sym, ok := ti.Defs.Get(ident); ok && sym != nil {
			return sym
		}
		if sym, ok := ti.Uses[ident]; ok && sym != nil {
			return sym
		}
	}
	return nil
}
