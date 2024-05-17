package checker

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type TypeInfo struct {
	// Constant data.
	Data *orderedmap.OrderedMap[ast.Node, *TypedValue]

	// Every definition in the entire module.
	Defs *orderedmap.OrderedMap[*ast.Ident, Symbol]

	// Type symbols for every type defined in the module.
	TypeSyms map[types.Type]Symbol

	// Type of every AST node.
	Types map[ast.Node]*TypedValue

	// Every usage in the entire module.
	Uses map[*ast.Ident]Symbol
}

func NewTypeInfo() *TypeInfo {
	return &TypeInfo{
		Data:     orderedmap.NewOrderedMap[ast.Node, *TypedValue](),
		Defs:     orderedmap.NewOrderedMap[*ast.Ident, Symbol](),
		TypeSyms: make(map[types.Type]Symbol),
		Types:    make(map[ast.Node]*TypedValue),
		Uses:     make(map[*ast.Ident]Symbol),
	}
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
		if tv, ok := ti.Types[expr]; ok && tv != nil {
			return tv
		}
	}
	return nil
}

func (ti *TypeInfo) SymbolOf(ident *ast.Ident) Symbol {
	if ident != nil {
		if sym, _ := ti.Defs.Get(ident); sym != nil {
			return sym
		}
		if sym := ti.Uses[ident]; sym != nil {
			return sym
		}
	}
	return nil
}
