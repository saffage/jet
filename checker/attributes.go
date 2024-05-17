package checker

import "github.com/saffage/jet/ast"

func GetAttribute(sym Symbol, name string) ast.Node {
	if decl, _ := sym.Node().(ast.Decl); decl != nil {
		return FindAttr(decl.Attributes(), name)
	} else if binding, _ := sym.Node().(*ast.Binding); binding != nil {
		return FindAttr(binding.Attrs, name)
	}
	return nil
}

func FindAttr(attrList *ast.AttributeList, attr string) ast.Node {
	if attrList != nil {
		for _, expr := range attrList.List.Exprs {
			if ident, _ := expr.(*ast.Ident); ident != nil && ident.Name == attr {
				return expr
			}
		}
	}
	return nil
}
