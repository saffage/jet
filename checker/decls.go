package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (check *Checker) resolveDecl(decl *ast.Decl) {
	switch {
	case decl.Ident.Name == "_":
		check.errorf(decl.Ident, "attempt to declare an empty identifier")

	case decl.IsVar:
		if sig, _ := decl.Type.(*ast.Signature); sig != nil &&
			decl.Value == nil {
			check.resolveFuncDecl(decl, &ast.Function{
				Signature: sig,
				Body:      nil,
			})
		} else {
			check.resolveVarDecl(decl)
		}

	case decl.Mut.IsValid():
		check.errorf(decl.Ident, "mutable compile-time variables are not supported")

	default:
		if decl.Value != nil {
			switch expr := decl.Value.(type) {
			case *ast.Function:
				check.resolveFuncDecl(decl, expr)

			case *ast.StructType:
				check.resolveStructDecl(decl, expr)

			case *ast.EnumType:
				check.resolveEnumDecl(decl, expr)

			default:
				value := check.valueOf(expr)
				if value == nil {
					if ty := check.typeOf(expr); ty != nil &&
						(types.IsTypeDesc(ty) || ty.Equals(types.Unit)) {
						check.resolveTypeAliasDecl(decl)
					} else {
						check.errorf(expr, "value is not a constant expression")
					}
				} else if value.Value == nil {
					check.resolveTypeAliasDecl(decl)
				} else {
					check.resolveConstDecl(decl)
				}
			}
		}
	}
}
