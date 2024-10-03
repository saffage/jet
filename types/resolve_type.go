package types

import (
	"github.com/saffage/jet/ast"
)

func (check *checker) typeSymOf(expr ast.Node) (t *TypeDef, err error) {
	switch node := expr.(type) {
	case *ast.Upper:
		if typedef := check.typeSymbolOf(node); typedef != nil {
			if typedef.Type() != nil {
				check.newUse(node, typedef)
				t = typedef
			} else {
				err = errorf(node, "expression has no type")
			}
		}

	case *ast.TypeVar:
		panic("unimplemented")

	default:
		panic(errorf(expr, "ill-formed AST: expected type, got %T instead", expr))
	}

	return
}

func (check *checker) resolveTypeDecl(node *ast.TypeDecl) {
	if node.Args != nil {
		check.errorf(node.Args, "type parameters are not implemented")
		return
	}

	// t := NewNamed(node.Name.String(), nil)

	switch expr := node.Expr.(type) {
	case *ast.Extern:
		check.resolveExternTypeDecl(expr, node)

	case *ast.Upper, *ast.TypeVar, *ast.Signature:
		panic("unimplemented")

	case *ast.Block:
		// for _, variant := range ty.Stmts.Nodes {
		// 	decl := variant.(*ast.Variant)
		// 	if decl.Params != nil {
		// 		panic("unimplemented")
		// 	}
		// 	check.newDef(decl.Name, newConstructor(check.env, t, decl))
		// }
		panic("unimplemented")

	default:
		panic(&errorIllFormedAst{node})
	}
}

func (check *checker) resolveTypeAlias(decl *ast.TypeDecl, t Type) {
	typedesc := As[*TypeDesc](t)

	if typedesc == nil {
		check.errorf(decl.Expr, "expression is not a type")
		return
	}

	sym := NewTypeDef(check.env, typedesc, decl)

	if defined := check.env.Define(sym); defined != nil {
		check.problem(&errorAlreadyDefined{sym.Ident(), defined.Ident()})
		return
	}

	check.newDef(decl.Name, sym)
	check.setType(decl, typedesc)
}
