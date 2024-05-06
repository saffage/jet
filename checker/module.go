package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

// Module is a file.
type Module struct {
	scope     *Scope
	node      *ast.ModuleDecl
	completed bool
}

func NewModule(node *ast.ModuleDecl) *Module {
	return &Module{
		scope:     NewScope(nil),
		node:      node,
		completed: false,
	}
}

func (m *Module) Owner() *Scope     { panic("modules have no owner") }
func (m *Module) Type() types.Type  { panic("modules have no type") }
func (m *Module) Name() string      { return m.node.Name.Name }
func (m *Module) Ident() *ast.Ident { return m.node.Name }
func (m *Module) Node() ast.Node    { return m.node }

func (check *Checker) visit(node ast.Node) ast.Visitor {
	decl, isDecl := node.(ast.Decl)

	if !isDecl {
		// NOTE parser should prevent this in future.
		check.errorf(node, "expected declaration")
		return nil
	}

	switch decl := decl.(type) {
	case *ast.ModuleDecl:
		panic("not implemented")

	case *ast.VarDecl:
		check.resolveVarDecl(decl)

	case *ast.FuncDecl:
		check.resolveFuncDecl(decl)

	case *ast.TypeAliasDecl:
		check.resolveTypeAliasDecl(decl)

	default:
		panic("unreachable")
	}

	return nil
}
