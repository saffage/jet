package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type ModuleKind byte

const (
	ModuleKindRegular ModuleKind = iota
	ModuleKindTypes
	ModuleKindC
)

// Module is a file.
type Module struct {
	*TypeInfo
	Scope   *Scope
	Imports []*Module

	node      *ast.ModuleDecl
	kind      ModuleKind
	completed bool
}

func NewModule(scope *Scope, node *ast.ModuleDecl) *Module {
	return &Module{
		TypeInfo:  NewTypeInfo(),
		Scope:     scope,
		node:      node,
		kind:      ModuleKindRegular,
		completed: false,
	}
}

func (m *Module) Owner() *Scope     { return m.Scope.parent }
func (m *Module) Type() types.Type  { return nil }
func (m *Module) Name() string      { return m.node.Name.Name }
func (m *Module) Ident() *ast.Ident { return m.node.Name }
func (m *Module) Node() ast.Node    { return m.node }

func (m *Module) TypeOf(expr ast.Node) types.Type {
	if expr != nil {
		if t := m.TypeInfo.TypeOf(expr); t != nil {
			return t
		}
		if ident, _ := expr.(*ast.Ident); ident != nil {
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
		if ident, _ := expr.(*ast.Ident); ident != nil {
			if _const, _ := m.SymbolOf(ident).(*Const); _const != nil {
				return _const.value
			}
		}
	}
	return nil
}

func (m *Module) SymbolOf(ident *ast.Ident) Symbol {
	if ident != nil {
		if sym := m.TypeInfo.SymbolOf(ident); sym != nil {
			return sym
		}
		if sym, _ := m.Scope.Lookup(ident.Name); sym != nil {
			return sym
		}
	}
	return nil
}

func (check *Checker) visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case ast.Decl:
		switch decl := node.(type) {
		case *ast.ModuleDecl:
			panic("not implemented")

		case *ast.VarDecl:
			check.resolveVarDecl(decl)

		case *ast.ConstDecl:
			check.resolveConstDecl(decl)

		case *ast.FuncDecl:
			check.resolveFuncDecl(decl)

		case *ast.StructDecl:
			check.resolveStructDecl(decl)

		case *ast.EnumDecl:
			check.resolveEnumDecl(decl)

		case *ast.TypeAliasDecl:
			check.resolveTypeAliasDecl(decl)

		default:
			panic("unreachable")
		}

	case *ast.Import:
		check.resolveImport(node)

	default:
		// NOTE parser should prevent this in future.
		check.errorf(node, "expected declaration")
		return nil
	}

	return nil
}
