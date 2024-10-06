package types

import "github.com/saffage/jet/ast"

type block struct {
	*checker
	t Type
}

func (block *block) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.LetDecl, *ast.TypeDecl:
		panic("unimplemented")

	default:
		if t, err := block.typeOf(node); err == nil && t != nil {
			block.t = t
		} else {
			block.error(err)
		}
	}

	return nil
}
