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

	evalType      types.Type // Local scopes are also expression.
	typeFromIdent ast.Node
	typeFromSym   Symbol
}

func NewLocalScope(parent Scope) (*LocalScope, error) {
	if parent == nil {
		return nil, NewError(nil, "the local scope must have a parent")
	}

	scope := &LocalScope{
		symbols:  []Symbol{},
		parent:   parent,
		evalType: types.Unknown{},
	}

	fmt.Printf(">>> push local\n")
	return scope, nil
}

func (scope *LocalScope) Parent() Scope {
	return scope.parent
}

func (scope *LocalScope) Define(symbol Symbol) error {
	if symbol == nil {
		panic("attempt to define nil symbol")
	}

	if prev := scope.ResolveMember(symbol.Name()); prev != nil {
		err := NewError(symbol.Ident(), "declaration shadows previous declaration")
		err.Notes = []Error{
			NewError(prev.Ident(), "previous declaration was here"),
		}
		return err
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

func (scope *LocalScope) visit(node ast.Node) (ast.Visitor, error) {
	switch n := node.(type) {
	case ast.Decl:
		switch decl := n.(type) {
		case *ast.GenericDecl:
			// TODO handle all names
			assert.Ok(len(decl.Field.Names) == 1)

			variable := NewVar(scope, nil, decl, decl.Field.Names[0])
			err := scope.Define(variable)
			if err != nil {
				return nil, err
			}

			fmt.Printf(">>> def local var `%s`\n", variable.Name())

			t, err := TypeOf(scope, decl.Field.Value)
			if err != nil {
				return nil, err
			}

			variable.type_ = t
			scope.evalType = t

			fmt.Printf(">>> set `%s` type `%s`\n", variable.Name(), t)
			return nil, nil

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
			scope.typeFromSym = sym
		}
		// panic(NewErrorf(n, "identifier `%s` is undefined", n.Name))
		scope.typeFromIdent = n
		return nil, nil
	}

	return scope.visit, nil
}
