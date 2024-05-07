package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Block struct {
	scope *Scope
	t     types.Type
}

func NewBlock(scope *Scope) *Block {
	return &Block{scope, types.Unit}
}

func (check *Checker) blockVisitor(expr *Block) ast.Visitor {
	return func(node ast.Node) ast.Visitor {
		if decl, _ := node.(ast.Decl); decl != nil {
			switch decl := decl.(type) {
			case *ast.VarDecl:
				check.resolveVarDecl(decl)
				expr.t = types.Unit

			case *ast.TypeAliasDecl, *ast.FuncDecl, *ast.ModuleDecl:
				check.errorf(decl, "a local scope can contain only variable declarations")
				return nil

			default:
				panic("unreachable")
			}

			return nil
		}

		t := check.typeOf(node)
		if t == nil {
			return nil
		}

		expr.t = t
		return nil
	}
}
