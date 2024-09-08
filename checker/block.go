package checker

import (
	"unicode"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type block struct {
	check *checker
	scope *Scope
	t     types.Type
}

func newBlock(check *checker, scope *Scope) *block {
	return &block{check, scope, types.Unit}
}

func (block *block) Visit(node ast.Node) ast.Visitor {
	if decl, _ := node.(*ast.LetDecl); decl != nil {
		if unicode.IsUpper([]rune(decl.Decl.Name.String())[0]) || FindAttr(decl.Attrs, "comptime") != nil {
			block.check.errorf(decl, "local constants are not supported")
			return nil
		}

		block.check.resolveVarDecl(decl)
		block.t = types.Unit
		return nil
	}

	t := block.check.typeOf(node)
	if t == nil {
		return nil
	}

	block.t = t
	return nil
}
