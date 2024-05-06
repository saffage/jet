package checker

import (
	"fmt"

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

func (expr *Block) visit(node ast.Node) (ast.Visitor, error) {
	switch node := node.(type) {
	case ast.Decl:
		switch decl := node.(type) {
		case *ast.VarDecl:
			if err := resolveVar(decl, expr.scope); err != nil {
				return nil, err
			}

			fmt.Printf(">>> def local var `%s`\n", decl.Binding.Name)

			expr.t = types.Unit
			return nil, nil

		case *ast.TypeAliasDecl, *ast.FuncDecl, *ast.ModuleDecl:
			panic("not implemented")

		default:
			panic("unreachable")
		}

	default:
		t, err := expr.scope.TypeOf(node)
		if err != nil {
			return nil, err
		}

		expr.t = t
		return nil, nil

		// fmt.Printf("unchecked node: '%T'\n", node)
		// expr.t = types.Unit
		// return nil, nil
	}
}
