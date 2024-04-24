package symbol

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/internal/log"
)

type delayed struct {
	symbol   Symbol
	required Symbol     // Which symbol is required
	use      *ast.Ident // Where symbol is required
}

func newDelayedSym(sym Symbol, requires Symbol, use *ast.Ident) delayed {
	if requires == nil || use == nil {
		panic("can't use nil for delayed symbol definition")
	}

	return delayed{sym, requires, use}
}

// Module is a file.
type Module struct {
	base
	symbols   []Symbol
	delayed   []delayed
	completed bool
}

func NewModule(id ID, name *ast.Ident, owner *Module, node ast.Node) *Module {
	list, ok := node.(*ast.List)
	assert.Ok(ok)

	m := &Module{
		base: base{
			name:  name,
			owner: owner,
			id:    id,
			type_: nil,
			node:  node,
		},
		symbols:   []Symbol{},
		delayed:   []delayed{},
		completed: false,
	}

	// Pass 1:
	//  * [x] define members
	//  * [ ] resolve usings
	//  * [ ] find cyclic usings

	walker := ast.NewWalker(m)

	for _, node := range list.Nodes {
		walker.Walk(node)
	}

	// Pass 2:
	//  * [ ] determine symbol types
	//  * [ ] find cyclic symbol definitions

	log.Hint("resolve symbol types")

	for _, sym := range m.symbols {
		m.resolveSymbolType(sym)
	}

	log.Hint("resolve delayed symbol types")

	for _, delayed := range m.delayed {
		fmt.Printf("resolving delayed symbol `%s`\n", delayed.symbol.Name())
		m.resolveSymbolType(delayed.symbol)
	}

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

func (m *Module) Use(module *Module) {
	if !module.Completed() {
		panic(fmt.Sprintf("module '%s' not completed", module.Name()))
	}

	for _, sym := range module.Symbols() {
		if m.Define(sym) != nil {
			panic(fmt.Sprintf("member '%s' is already defined", sym.Name()))
		}
	}
}

func (m *Module) Completed() bool {
	return m.completed
}

func (m *Module) MarkCompleted() {
	m.completed = true
}

func (m *Module) TypeOf(expr ast.Node) (types.Type, error) {
	return nil, nil
}

func (m *Module) Visit(node ast.Node) ast.Visitor {
	if _, isEmpty := node.(*ast.Empty); isEmpty {
		return nil
	}

	decl, isDecl := node.(ast.Decl)

	if !isDecl {
		// NOTE parser should prevent this in future
		// panic(fmt.Sprintf("checker.(*ScopeBuilder).Visit: invalid top-level statement (expected declaration): got '%T' node", node))
		panic(NewError(node, "expected declaration"))
	}

	switch d := decl.(type) {
	case *ast.ModuleDecl:
		panic("todo")

	case *ast.GenericDecl:
		switch d.Kind {
		case ast.VarDecl:

		default:
			panic(NewError(d, "the only `var` declarations are supported for now"))
		}

		for _, name := range d.Field.Names {
			if definedSym := m.Define(NewVar(0, name, d, m)); definedSym != nil {
				panic(NewErrorf(d.Ident(), "variable `%s` is already defined", d.Ident().Name))
			}

			fmt.Printf(">>> def var `%s`\n", name.Name)
		}

	case *ast.FuncDecl:
		if definedSym := m.Define(NewFunc(0, d, m)); definedSym != nil {
			panic(NewErrorf(d.Ident(), "function `%s` is already defined", d.Ident().Name))
		}

		fmt.Printf(">>> def func `%s`\n", d.Name.Name)

	case *ast.StructDecl:
		panic("todo")

	case *ast.EnumDecl:
		panic("todo")

	case *ast.TypeAliasDecl:
		panic("todo")

	default:
		panic(fmt.Sprintf("unknown declaration type '%T'", decl))
	}

	return nil
}

func (m *Module) genRecursiveDeclNotes(first delayed) (notes []Error, note string) {
	delayed := &first
	note = fmt.Sprintf("`%s`", first.symbol.Name())

	for {
		note += fmt.Sprintf(" -> `%s`", delayed.required.Name())
		note := NewErrorf(delayed.use, "`%s` requires `%s`", delayed.symbol.Name(), delayed.required.Name())
		notes = append(notes, note)

		if delayed.required == first.symbol {
			if len(notes) == 0 {
				// panic("internal error: self-recursive symbol must be handled somewhere else")
				panic("todo")
			}

			// found cycle
			break
		}

		delayed = m.resolveDelayed(delayed.required)

		if delayed == nil {
			// not a cycle
			return nil, ""
		}
	}
	return
}

func (m *Module) resolveDelayed(sym Symbol) *delayed {
	for _, delayed := range m.delayed {
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
		explicitType, _, _ := types.Type(nil), Symbol(nil), (*ast.Ident)(nil)

		if node.Field.Type != nil {
			explicitType, _, _ = TypeOf(m, node.Field.Type)
		}

		valueType, untypedSym, use := TypeOf(m, node.Field.Value)

		if types.IsUnknown(valueType) && untypedSym == nil {
			assert.Ok(use != nil)
			panic(NewErrorf(use, "identifier `%s` is undefined", use.Name))
		}

		if valueType == nil {
			panic(NewError(node.Field.Value, "expression has no type"))
		}

		if types.IsUnknown(valueType) {
			if delayed := m.resolveDelayed(sym); delayed != nil {
				notes, note := m.genRecursiveDeclNotes(*delayed)
				err := NewErrorf(node.Ident(), "recursive symbol definition (%s)", note)
				err.Notes = notes
				panic(err)
			}

			fmt.Printf(">>> delay `%s` for `%s`\n", sym.Name(), untypedSym.Name())
			m.delayed = append(m.delayed, newDelayedSym(sym, untypedSym, use))
			return // delayed
		}

		if explicitType != nil && !explicitType.Equals(valueType) {
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

	default:
		panic("todo")
	}
}
