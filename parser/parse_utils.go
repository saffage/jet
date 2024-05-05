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

	case *ast.VarDecl:
		n.Attrs = attrs
	}
}

// func setDoc(node ast.Decl, commentGroup *ast.CommentGroup) {
// 	switch n := node.(type) {
// 	case *ast.ModuleDecl:
// 		n.CommentGroup = commentGroup

// 	case *ast.TypeAliasDecl:
// 		n.CommentGroup = commentGroup

// 	case *ast.FuncDecl:
// 		n.CommentGroup = commentGroup

// 	case *ast.VarDecl:
// 		n.CommentGroup = commentGroup
// 	}
// }
