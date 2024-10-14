package types

import (
	"github.com/saffage/jet/ast"
)

func HasAttribute(sym Symbol, name string) bool {
	return GetAttribute(sym, name) != nil
}

func GetAttribute(sym Symbol, name string) ast.Node {
	if decl, _ := sym.Node().(*ast.LetDecl); decl != nil {
		return FindAttr(decl.Attrs, name)
	}
	return nil
}

func FindAttr(attrList *ast.AttributeList, attr string) ast.Node {
	if attrList != nil {
		for _, expr := range attrList.List.Nodes {
			switch expr := expr.(type) {
			case *ast.Lower:
				if expr != nil && expr.Data == attr {
					return expr
				}

			case *ast.Call:
				ident, _ := expr.X.(*ast.Lower)

				if ident != nil && ident.Data == attr {
					return expr
				}
			}
		}
	}
	return nil
}
