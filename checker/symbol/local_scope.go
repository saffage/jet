package symbol

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/internal/assert"
)

type LocalScope struct {
	symbols []Symbol
	parent  Scope

	evalType types.Type // Local scopes are also expression.
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

// Exists only for debug.
func (scope *LocalScope) Free() {
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

func (scope *LocalScope) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case ast.Decl:
		switch decl := n.(type) {
		case *ast.GenericDecl:
			// TODO handle all names
			assert.Ok(len(decl.Field.Names) == 1)

			variable := NewVar(0, decl.Field.Names[0], decl, scope)

			if declared := scope.Define(variable); declared != nil {
				err := NewErrorf(decl.Field.Names[0], "declaration shadows previous declaration")
				err.Notes = []Error{NewError(declared.Ident(), "previous declaration was here")}
				panic(err)
			}

			fmt.Printf(">>> def local var `%s`\n", variable.Name())

			type_, err := TypeOf(scope, decl.Field.Value)
			if err != nil {
				panic(err)
			}

			variable.setType(type_)
			scope.evalType = type_
			return nil

		case *ast.TypeAliasDecl, *ast.FuncDecl, *ast.ModuleDecl:
			panic("not implemented")

		default:
			panic("unreachable")
		}

	case *ast.Ident:
		if sym := scope.Resolve(n.Name); sym != nil {
			if sym.Type() != nil {
				scope.evalType = sym.Type()
			} else {
				// Symbol is defined but have no type yet. Defered sym?
				scope.evalType = types.Unknown{}
			}
			scope.typeSym = sym
		}
		// panic(NewErrorf(n, "identifier `%s` is undefined", n.Name))
		scope.typeFrom = n
		return nil
	}

	return scope.Visit
}
