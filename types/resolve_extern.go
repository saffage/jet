package types

import "github.com/saffage/jet/ast"

func (check *checker) resolveExternTypeDecl(extern *ast.Extern, node *ast.TypeDecl) {
	if extern.Args != nil {
		panic("unimplemented")
	}

	externName := node.Name.Data

	var typedesc *TypeDesc

	switch externName {
	case "Int":
		typedesc = NewTypeDesc(IntType)

	case "Float":
		typedesc = NewTypeDesc(FloatType)

	case "String":
		typedesc = NewTypeDesc(StringType)

	default:
		check.problem(&errorUnknownExtern{extern, externName})
		return
	}

	sym := NewTypeDef(check.env, typedesc, node)
	check.newDef(node.Name, sym)
}
