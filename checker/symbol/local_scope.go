package symbol

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
)

type LocalScope struct {
	symbols []Symbol
	parent  Scope

	type_    types.Type // Used for scope type
	typeFrom ast.Node
	typeSym  Symbol
}

func NewLocalScope(parent Scope) *LocalScope {
	if parent == nil {
		panic("the local scope must have a parent")
	}

	fmt.Printf(">>> push local\n")

	return &LocalScope{
		symbols: []Symbol{},
		parent:  parent,
	}
}

func (scope *LocalScope) Free() {
	// Exists only for debug.
	fmt.Printf(">>> pop local\n")
}

func (scope *LocalScope) Parent() Scope {
	return scope.parent
}

func (scope *LocalScope) Define(symbol Symbol) Symbol {
	if symbol == nil {
		panic("attempt to define nil symbol")
	}

	if sym := scope.ResolveMember(symbol.Name()); sym != nil {
		return sym
	}

	scope.symbols = append(scope.symbols, symbol)
	return nil
}

func (scope *LocalScope) Resolve(name string) Symbol {
	if sym := scope.ResolveMember(name); sym != nil {
		return sym
	}

	return scope.parent.Resolve(name)
}

func (scope *LocalScope) ResolveMember(name string) Symbol {
	for _, sym := range scope.symbols {
		if sym.Name() == name {
			return sym
		}
	}

	return nil
}

func (scope *LocalScope) Symbols() []Symbol {
	return scope.symbols
}

func (scope *LocalScope) Use(module *Module) {
	if !module.Completed() {
		panic(NewErrorf(nil, "module '%s' is not completed", module.Name()))
	}

	for _, sym := range module.Symbols() {
		if scope.Define(sym) != nil {
			panic(NewErrorf(nil, "member '%s' is already defined", sym.Name()))
		}
	}
}

func (scope *LocalScope) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case ast.Decl:
		switch decl := n.(type) {
		case *ast.AliasDecl:
		case *ast.EnumDecl:
		case *ast.FuncDecl:

		case *ast.GenericDecl:
			// TODO handle all names
			variable := NewVar(0, decl.Field.Names[0], decl, scope)

			if declared := scope.Define(variable); declared != nil {
				err := NewErrorf(decl.Field.Names[0], "declaration shadows previous declaration")
				err.Notes = []Error{NewError(declared.NameNode(), "previous declaration was here")}
				panic(err)
			}

			fmt.Printf(">>> def local var `%s`\n", variable.Name())

			type_, _, where := TypeOf(scope, decl.Field.Value)

			if where != nil {
				panic(NewErrorf(where, "identifier `%s` is undefined", where.Name))
			}

			variable.setType(type_)
			scope.type_ = type_
			return nil

		case *ast.ModuleDecl:
		case *ast.StructDecl:

		default:
			panic("unreachable")
		}
		panic("todo")

	case *ast.Ident:
		if sym := scope.Resolve(n.Name); sym != nil {
			if sym.Type() != nil {
				scope.type_ = sym.Type()
			} else {
				// Symbol is defined but have no type yet. Defered sym?
				scope.type_ = types.Unknown{}
			}
			scope.typeSym = sym
		}
		// panic(NewErrorf(n, "identifier `%s` is undefined", n.Name))
		scope.typeFrom = n
		return nil
	}

	return scope
}
