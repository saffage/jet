package checker

import "github.com/saffage/jet/ast"

func getAttribute(sym Symbol, name string) ast.Node {
	if decl, _ := sym.Node().(ast.Decl); decl != nil {
		attrList := decl.Attributes()
		if attrList == nil {
			return nil
		}

		for _, expr := range attrList.List.Exprs {
			switch expr := expr.(type) {
			case *ast.Ident:
				if expr.Name == name {
					return expr
				}
			}
		}
	}

	return nil
}
