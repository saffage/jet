package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

// Module is a file.
type Module struct {
	scope     *Scope
	node      *ast.ModuleDecl
	completed bool
}

func NewModule(node *ast.ModuleDecl) (*Module, error) {
	m := &Module{
		scope: NewScope(nil),
		node:  node,
	}

	nodes := []ast.Node(nil)

	switch body := node.Body.(type) {
	case *ast.List:
		nodes = body.Nodes

	case *ast.CurlyList:
		nodes = body.List.Nodes

	default:
		panic("ill-formed AST")
	}

	for _, node := range nodes {
		if err := ast.WalkTopDown(m.visit, node); err != nil {
			return nil, err
		}
	}

	m.completed = true
	return m, nil
}

func (m *Module) Owner() *Scope     { panic("modules have no owner") }
func (m *Module) Type() types.Type  { panic("modules have no type") }
func (m *Module) Name() string      { return m.node.Name.Name }
func (m *Module) Ident() *ast.Ident { return m.node.Name }
func (m *Module) Node() ast.Node    { return m.node }

func (m *Module) setType(t types.Type) { panic("modules have no type") }

func (m *Module) visit(node ast.Node) (ast.Visitor, error) {
	if _, isEmpty := node.(*ast.Empty); isEmpty {
		return nil, nil
	}

	decl, isDecl := node.(ast.Decl)

	if !isDecl {
		// NOTE parser should prevent this in future.
		return nil, NewError(node, "expected declaration")
	}

	switch decl := decl.(type) {
	case *ast.ModuleDecl:
		panic("not implemented")

	case *ast.VarDecl:
		sym := NewVar(m.scope, nil, decl.Binding, decl.Binding.Name)
		t, err := resolveVarDecl(decl, m.scope)
		if err != nil {
			return nil, err
		}

		sym.setType(t)

		if defined := m.scope.Define(sym); defined != nil {
			return nil, errorAlreadyDefined(sym.Ident(), defined.Ident())
		}

		fmt.Printf(">>> def var `%s`\n", decl.Binding.Name)

	case *ast.FuncDecl:
		sym := NewFunc(m.scope, nil, decl)

		if defined := m.scope.Define(sym); defined != nil {
			return nil, errorAlreadyDefined(sym.Ident(), defined.Ident())
		}

		fmt.Printf(">>> def func `%s`\n", decl.Name.Name)

		if err := resolveFuncDecl(sym); err != nil {
			return nil, err
		}

	case *ast.TypeAliasDecl:
		t, err := m.scope.TypeOf(decl.Expr)
		if err != nil {
			return nil, err
		}

		t = types.SkipAlias(t)

		typedesc, _ := t.(*types.TypeDesc)
		if typedesc == nil {
			return nil, NewErrorf(decl.Expr, "expression is not a type (%s)", t)
		}

		sym := NewTypeAlias(m.scope, typedesc, decl)

		if defined := m.scope.Define(sym); defined != nil {
			return nil, errorAlreadyDefined(sym.Ident(), defined.Ident())
		}

		fmt.Printf(">>> def alias `%s` for `%s`\n", sym.Name(), typedesc.Base())

	default:
		panic(fmt.Sprintf("unhandled declaration kind (%T)", decl))
	}

	return nil, nil
}
