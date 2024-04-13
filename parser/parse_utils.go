package parser

import "github.com/saffage/jet/ast"

func setAnnotations(node ast.Decl, annotations []*ast.Annotation) {
	switch n := node.(type) {
	case *ast.ModuleDecl:
		n.Annots = annotations

	case *ast.AliasDecl:
		n.Annots = annotations

	case *ast.StructDecl:
		n.Annots = annotations

	case *ast.FuncDecl:
		n.Annots = annotations

	case *ast.GenericDecl:
		n.Annots = annotations
	}
}
