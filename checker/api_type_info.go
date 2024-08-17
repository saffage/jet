package checker

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type TypeInfo struct {
	// Every definition in the entire module.
	Defs *orderedmap.OrderedMap[ast.Ident, Symbol]

	// Type symbols for every type defined in the module.
	TypeSyms map[types.Type]Symbol

	// Type of every AST node.
	Types map[ast.Node]*TypedValue

	// Every usage in the entire module.
	Uses map[ast.Ident]Symbol
}

func (ti *TypeInfo) TypeOf(expr ast.Node) types.Type {
	if expr != nil {
		if tv, ok := ti.Types[expr]; ok && tv != nil {
			return tv.Type
		}
	}
	return nil
}

func (ti *TypeInfo) ValueOf(expr ast.Node) *TypedValue {
	if expr != nil {
		if tv, ok := ti.Types[expr]; ok && tv != nil && tv.Value != nil {
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
