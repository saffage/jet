package parser

import (
	"github.com/saffage/jet/ast"
)

func setAttributes(node ast.Decl, attrs *ast.AttributeList) {
	switch n := node.(type) {
	case *ast.ModuleDecl:
		n.Attrs = attrs

	case *ast.TypeAliasDecl:
		n.Attrs = attrs

	case *ast.FuncDecl:
		n.Attrs = attrs

	case *ast.GenericDecl:
		n.Attrs = attrs
	}
}
