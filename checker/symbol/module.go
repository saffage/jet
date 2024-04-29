package symbol

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/internal/log"
)

// Module is a file.
type Module struct {
	base
	symbols   []Symbol
	deferred  []deferred
	completed bool
}

func NewModule(id ID, owner *Module, name *ast.Ident, node ast.Node) *Module {
	list, ok := node.(*ast.List)
	assert.Ok(ok)

	m := &Module{
		base: base{
			id:    id,
			owner: owner,
			name:  name,
			node:  node,
		},
		symbols:   []Symbol{},
		deferred:  []deferred{},
		completed: false,
	}
	appendBuiltins(m)

	// Pass 1:
	//  * [x] define members
	//  * [ ] resolve usings
	//  * [ ] find cyclic usings

	for _, node := range list.Nodes {
		ast.WalkTopDown(m.Visit, node)
	}

	// Pass 2:
	//  * [ ] determine symbol types
	//  * [ ] find cyclic symbol definitions

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
	return m
}

func (m *Module) Parent() Scope {
	return m.owner
}

func (m *Module) Define(symbol Symbol) Symbol {
	if symbol == nil {
		panic("attempt to define nil symbol")
	}

	if sym := m.ResolveMember(symbol.Name()); sym != nil {
		return sym
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

func (m *Module) Visit(node ast.Node) ast.Visitor {
	if _, isEmpty := node.(*ast.Empty); isEmpty {
		return nil
	}

	decl, isDecl := node.(ast.Decl)

	if !isDecl {
		// NOTE parser should prevent this in future
		panic(NewError(node, "expected declaration"))
	}

	switch d := decl.(type) {
	case *ast.ModuleDecl:
		panic("todo")

	case *ast.GenericDecl:
		switch d.Kind {
		case ast.VarDecl:
			for _, name := range d.Field.Names {
				sym := NewVar(0, name, d, m)

				if definedSym := m.Define(sym); definedSym != nil {
					panic(NewErrorf(d.Ident(), "variable `%s` is already defined", d.Ident().Name))
				}

				fmt.Printf(">>> def var `%s`\n", name.Name)
			}

		case ast.ValDecl:
			panic(NewError(d, "`val` declarations are supported for now"))

		case ast.ConstDecl:
			for _, name := range d.Field.Names {
				sym := NewConst(0, name, d, m)

				if definedSym := m.Define(sym); definedSym != nil {
					err := NewErrorf(name, "constant `%s` is already defined", name)
					err.Notes = append(err.Notes, NewError(definedSym.Ident(), "previous name was defined here"))
					panic(err)
				}

				fmt.Printf(">>> def const `%s`\n", name.Name)
			}

		default:
			panic("unreachable")
		}

	case *ast.FuncDecl:
		sym := NewFunc(0, d, m)

		if definedSym := m.Define(sym); definedSym != nil {
			panic(NewErrorf(d.Ident(), "function `%s` is already defined", d.Ident().Name))
		}

		fmt.Printf(">>> def func `%s`\n", d.Name.Name)

	case *ast.TypeAliasDecl:
		sym := NewTypeAlias(0, nil, d, m)

		if definedSym := m.Define(sym); definedSym != nil {
			panic(NewErrorf(d.Ident(), "name `%s` is already bound", d.Ident().Name))
		}

		fmt.Printf(">>> def alias `%s`\n", sym.Name())

	default:
		panic(fmt.Sprintf("unhandled declaration kind (%T)", decl))
	}

	return nil
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

func (m *Module) resolveSymbolType(sym Symbol) {
	if sym.Type() != nil {
		return
	}

	switch node := sym.Node().(type) {
	case *ast.GenericDecl:
		resolveGenericDeclType(m, sym, node)

	case *ast.TypeAliasDecl:
		resolveTypeAliasDeclType(m, sym, node)

	case *ast.FuncDecl:
		// panic("not implemented")

	case *ast.BuiltInCall:
		// Nothing to do.

	default:
		if sym.Node() == nil {
			log.Hint("symbol `%s` doesn't have a node to resolve their type, skipped", sym.Name())
			return
		}

		panic("todo")
	}
}

func resolveGenericDeclType(m *Module, sym Symbol, node *ast.GenericDecl) {
	explicitType, valueType := types.Type(nil), types.Type(nil)
	nilType := (*nilTypeError)(nil)

	if node.Field.Type != nil {
		type_, err := TypeOf(m, node.Field.Type)
		if err != nil {
			panic(err)
		}

		explicitType = types.UnwrapTypeDesc(type_)
	}

	if node.Field.Value != nil {
		type_, err := TypeOf(m, node.Field.Value)

		if nilTypeErr, ok := err.(*nilTypeError); ok {
			nilType = nilTypeErr
		} else if err != nil {
			panic(err)
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
				panic(err)
			}

			m.resolveSymbolType(deferred.required)
			valueType = deferred.required.Type()
		} else {
			fmt.Printf(">>> defer `%s` for `%s`\n", sym.Name(), nilType.sym.Name())
			m.deferred = append(m.deferred, newDeferredSym(sym, nilType.sym, nilType.use))

			// symbol was deferred
			return
		}
	}

	if valueType == nil {
		panic(NewErrorf(node, "cannot infer type of the symbol"))
	}

	if node.Kind != ast.ConstDecl {
		valueType = types.TypedFromUntyped(valueType)
	}

	if explicitType != nil && !types.IsUnknown(explicitType) && !explicitType.Equals(valueType) {
		panic(NewErrorf(
			node.Ident(),
			"invalid type for `%s`, expected %s, got %s",
			sym.Name(),
			explicitType.String(),
			valueType.String(),
		))
	}

	sym.setType(valueType)
	fmt.Printf(">>> set `%s` type `%s`\n", sym.Name(), valueType.String())
}

func resolveTypeAliasDeclType(m *Module, sym Symbol, node *ast.TypeAliasDecl) {
	type_, err := TypeOf(m, node.Expr)

	if nilTypeErr, ok := err.(*nilTypeError); ok {
		fmt.Printf(">>> defer `%s` for `%s`\n", sym.Name(), nilTypeErr.sym.Name())
		m.deferred = append(m.deferred, newDeferredSym(sym, nilTypeErr.sym, nilTypeErr.use))
		return
	} else if err != nil {
		panic(err)
	}

	if !types.IsTypeDesc(type_) {
		panic(NewErrorf(
			node.Expr,
			"expected type descriptor for type alias, but expression is of type '%s'",
			type_.String(),
		))
	}

	sym.setType(type_)
	fmt.Printf(">>> set `%s` type `%s`\n", sym.Name(), type_)
}

func appendBuiltins(m *Module) {
	builtins := []*Builtin{
		{
			base: base{
				name: &ast.Ident{Name: "@magic"},
			},
			params: []BuiltinParam{
				{
					name:  &ast.Ident{Name: "name"},
					type_: types.UntypedString{},
				},
			},
			fn: builtinMagic,
		},
		{
			base: base{
				name: &ast.Ident{Name: "@type_of"},
			},
			params: []BuiltinParam{
				{
					name:  &ast.Ident{Name: "expr"},
					type_: types.Any{},
				},
			},
			fn: builtinTypeOf,
		},
	}

	for _, builtin := range builtins {
		m.Define(builtin)
	}
}
