package symbol

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/internal/log"
)

// Module is a file.
type Module struct {
	id    ID
	owner Scope

	node      *ast.ModuleDecl
	symbols   []Symbol
	deferred  []deferred
	completed bool
}

func NewModule(node *ast.ModuleDecl, owner *Module) (*Module, error) {
	m := &Module{
		id:        nextID(),
		owner:     owner,
		node:      node,
		symbols:   []Symbol{},
		deferred:  []deferred{},
		completed: false,
	}

	// Pass 1:
	//  * [x] define members
	//  * [ ] resolve usings
	//  * [ ] find cyclic usings

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

	// Pass 2:
	//  * [ ] resolve symbol types
	//  * [x] find cyclic symbol definitions

	log.Hint("resolve symbol types")

	for _, sym := range m.symbols {
		m.resolveSymbolType(sym)
	}

	log.Hint("resolve deferred symbol types")

	for _, deferred := range m.deferred {
		fmt.Printf("resolving deferred symbol `%s`\n", deferred.symbol.Name())
		m.resolveSymbolType(deferred.symbol)
	}

	m.completed = true
	return m, nil
}

func (m *Module) ID() ID            { return m.id }
func (m *Module) Owner() Scope      { return m.owner }
func (m *Module) Type() types.Type  { return nil }
func (m *Module) Name() string      { return m.node.Name.Name }
func (m *Module) Ident() *ast.Ident { return m.node.Name }
func (m *Module) Node() ast.Node    { return m.node }

func (m *Module) setType(t types.Type) { panic("modules have no type") }

func (m *Module) Parent() Scope {
	return m.owner
}

func (m *Module) Define(symbol Symbol) error {
	if symbol == nil {
		panic("attempt to define nil symbol")
	}

	if prev := m.ResolveMember(symbol.Name()); prev != nil {
		err := NewErrorf(symbol.Ident(), "name '%s' is already declared in this scope", symbol.Name())
		err.Notes = []Error{
			NewError(prev.Ident(), "previous declaration was here"),
		}
		return err
	}

	m.symbols = append(m.symbols, symbol)
	return nil
}

func (m *Module) Resolve(name string) Symbol {
	return m.ResolveMember(name)
}

func (m *Module) ResolveMember(name string) Symbol {
	if name == "" {
		return nil
	}

	for _, sym := range m.symbols {
		if sym.Name() == name {
			return sym
		}
	}

	return nil
}

func (m *Module) Symbols() []Symbol {
	return m.symbols
}

func (m *Module) IsCompleted() bool {
	return m.completed
}

func (m *Module) visit(node ast.Node) (ast.Visitor, error) {
	if _, isEmpty := node.(*ast.Empty); isEmpty {
		return nil, nil
	}

	decl, isDecl := node.(ast.Decl)

	if !isDecl {
		// NOTE parser should prevent this in future
		return nil, NewError(node, "expected declaration")
	}

	switch d := decl.(type) {
	case *ast.ModuleDecl:
		panic("not implemented")

	case *ast.GenericDecl:
		switch d.Kind {
		case ast.VarDecl:
			for _, name := range d.Field.Names {
				variable := NewVar(m, nil, d, name)

				if err := m.Define(variable); err != nil {
					return nil, err
				}

				fmt.Printf(">>> def var `%s`\n", name.Name)
			}

		case ast.ValDecl:
			return nil, NewError(d, "`val` declarations are not supported for now")

		case ast.ConstDecl:
			for _, name := range d.Field.Names {
				sym, err := NewConst(m, d, name)
				if err != nil {
					return nil, err
				}

				if err := m.Define(sym); err != nil {
					return nil, err
				}

				fmt.Printf(">>> def const `%s`\n", name.Name)
			}

		default:
			panic("unreachable")
		}

	case *ast.FuncDecl:
		sym := NewFunc(m, nil, d)

		if err := m.Define(sym); err != nil {
			return nil, err
		}

		fmt.Printf(">>> def func `%s`\n", d.Name.Name)

	case *ast.TypeAliasDecl:
		sym := NewTypeAlias(m, nil, d)

		if err := m.Define(sym); err != nil {
			return nil, err
		}

		fmt.Printf(">>> def alias `%s`\n", sym.Name())

	default:
		panic(fmt.Sprintf("unhandled declaration kind (%T)", decl))
	}

	return nil, nil
}

type deferred struct {
	symbol   Symbol
	required Symbol     // Which symbol is required.
	use      *ast.Ident // Where symbol is required.
}

func newDeferredSym(sym Symbol, requires Symbol, use *ast.Ident) deferred {
	if requires == nil || use == nil {
		panic("can't use nil for deferred symbol definition")
	}

	return deferred{sym, requires, use}
}

func (m *Module) genRecursiveDeclNotes(first deferred) (notes []Error) {
	for deferred := &first; deferred != nil; deferred = m.resolveDeferred(deferred.required) {
		if deferred.symbol == deferred.required {
			notes = append(notes, NewErrorf(
				deferred.use,
				"`%s` requires itself",
				deferred.symbol.Name(),
			))
			return notes
		}

		notes = append(notes, NewErrorf(
			deferred.use,
			"`%s` requires `%s`",
			deferred.symbol.Name(),
			deferred.required.Name(),
		))

		if deferred.required == first.symbol {
			// found cycle
			return notes
		}
	}

	// not a cycle
	return nil
}

func (m *Module) resolveDeferred(sym Symbol) *deferred {
	for _, delayed := range m.deferred {
		if delayed.symbol == sym {
			return &delayed
		}
	}

	return nil
}

func (m *Module) resolveSymbolType(symbol Symbol) {
	if symbol.Type() != nil {
		return
	}

	switch sym := symbol.(type) {
	case *Const:
		resolveGenericDeclType(m, sym, sym.node)

	case *Var:
		resolveGenericDeclType(m, sym, sym.node)

	case *TypeAlias:
		resolveTypeAliasDeclType(m, symbol, sym.node)

	case *BuiltIn:
		// Nothing to do.

	default:
		if symbol.Node() == nil {
			log.Hint("symbol `%s` doesn't have a node to resolve their type, skipped", symbol.Name())
			return
		}

		panic("not implemented")
	}
}

func resolveGenericDeclType(m *Module, sym Symbol, node *ast.GenericDecl) error {
	explicitType, valueType := types.Type(nil), types.Type(nil)
	nilType := (*nilTypeError)(nil)

	if node.Field.Type != nil {
		type_, err := TypeOf(m, node.Field.Type)
		if err != nil {
			return err
		}

		explicitType = types.UnwrapTypeDesc(type_)
	}

	if node.Field.Value != nil {
		type_, err := TypeOf(m, node.Field.Value)

		if nilTypeErr, ok := err.(*nilTypeError); ok {
			nilType = nilTypeErr
		} else if err != nil {
			return err
		}

		valueType = type_
	} else if explicitType != nil && !types.IsUnknown(explicitType) {
		valueType = explicitType
	}

	if types.IsUnknown(valueType) {
		if deferred := m.resolveDeferred(sym); deferred != nil {
			if notes := m.genRecursiveDeclNotes(*deferred); notes != nil {
				err := NewError(node.Ident(), "recursive symbol definition")
				err.Notes = notes
				return err
			}

			m.resolveSymbolType(deferred.required)
			valueType = deferred.required.Type()
		} else if nilType != nil {
			fmt.Printf(">>> defer `%s` for `%s`\n", sym.Name(), nilType.sym.Name())
			m.deferred = append(m.deferred, newDeferredSym(sym, nilType.sym, nilType.use))

			// symbol was deferred
			return nil
		} else {
			panic("todo")
		}
	}

	if valueType == nil {
		return NewErrorf(node, "cannot infer type of the symbol")
	}

	if node.Kind != ast.ConstDecl {
		valueType = types.TypedFromUntyped(valueType)
	}

	if explicitType != nil && !types.IsUnknown(explicitType) && !explicitType.Equals(valueType) {
		return NewErrorf(
			node.Ident(),
			"invalid type for `%s`, expected %s, got %s",
			sym.Name(),
			explicitType.String(),
			valueType.String(),
		)
	}

	sym.setType(valueType)
	fmt.Printf(">>> set `%s` type `%s`\n", sym.Name(), valueType.String())
	return nil
}

func resolveTypeAliasDeclType(m *Module, sym Symbol, node *ast.TypeAliasDecl) error {
	t, err := TypeOf(m, node.Expr)

	if nilTypeErr, ok := err.(*nilTypeError); ok {
		fmt.Printf(">>> defer `%s` for `%s`\n", sym.Name(), nilTypeErr.sym.Name())
		m.deferred = append(m.deferred, newDeferredSym(sym, nilTypeErr.sym, nilTypeErr.use))
		return nil
	} else if err != nil {
		return err
	}

	if !types.IsTypeDesc(t) {
		return NewErrorf(
			node.Expr,
			"expected type descriptor for type alias, but expression is of type '%s'",
			t.String(),
		)
	}

	sym.setType(t)
	fmt.Printf(">>> set `%s` type `%s`\n", sym.Name(), t)
	return nil
}
