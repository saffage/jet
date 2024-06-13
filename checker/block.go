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

func (check *Checker) visitBlock(expr *Block) ast.Visitor {
	return func(node ast.Node) ast.Visitor {
		if decl, _ := node.(*ast.Decl); decl != nil {
			if !decl.IsVar {
				check.errorf(decl, "local constants are not supported")
				return nil
			}

			check.resolveVarDecl(decl)
			expr.t = types.Unit
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
