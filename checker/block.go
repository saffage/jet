package checker

import (
	"unicode"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Block struct {
	check *Checker
	scope *Scope
	t     types.Type
}

func NewBlock(check *Checker, scope *Scope) *Block {
	return &Block{check, scope, types.Unit}
}

func (block *Block) Visit(node ast.Node) ast.Visitor {
	if decl, _ := node.(*ast.LetDecl); decl != nil {
		if unicode.IsUpper([]rune(decl.Decl.Name.Ident())[0]) || FindAttr(decl.Attrs, "comptime") != nil {
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
