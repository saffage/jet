package checker

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Module struct {
	*TypeInfo
	Scope   *Scope
	Imports []*Module

	name      string
	stmts     *ast.Stmts
	completed bool
}

func NewModule(scope *Scope, name string, stmts *ast.Stmts) *Module {
	return &Module{
		TypeInfo: &TypeInfo{
			Defs:     orderedmap.NewOrderedMap[ast.Ident, Symbol](),
			TypeSyms: make(map[types.Type]Symbol),
			Types:    make(map[ast.Node]*TypedValue),
			Uses:     make(map[ast.Ident]Symbol),
		},
		Scope:     scope,
		name:      name,
		stmts:     stmts,
		completed: false,
	}
}

func (m *Module) Owner() *Scope    { return m.Scope.parent }
func (m *Module) Type() types.Type { return nil }
func (m *Module) Name() string     { return m.name }
func (m *Module) Ident() ast.Ident { return nil }
func (m *Module) Node() ast.Node   { return m.stmts }

func (m *Module) TypeOf(expr ast.Node) types.Type {
	if expr != nil {
		if t := m.TypeInfo.TypeOf(expr); t != nil {
			return t
		}
		if ident, _ := expr.(*ast.Name); ident != nil {
			if sym := m.SymbolOf(ident); sym != nil {
				return sym.Type()
			}
		}
	}
	return nil
}

func (m *Module) ValueOf(expr ast.Node) *TypedValue {
	if expr != nil {
		if t := m.TypeInfo.ValueOf(expr); t != nil {
			return t
		}
		if ident, _ := expr.(*ast.Name); ident != nil {
			if _const, _ := m.SymbolOf(ident).(*Const); _const != nil {
				return _const.value
			}
		}
	}
	return nil
}

func (m *Module) SymbolOf(ident ast.Ident) Symbol {
	if sym := m.TypeInfo.SymbolOf(ident); sym != nil {
		return sym
	}
	if sym, _ := m.Scope.Lookup(ast.Data(ident)); sym != nil {
		return sym
	}
	return nil
}

func (check *checker) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.LetDecl:
		check.resolveLetDecl(node)

	case *ast.TypeDecl:
		check.resolveTypeDecl(node)

	default:
		panic("ill-formed AST")
	}

	return nil
}
