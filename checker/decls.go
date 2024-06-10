package checker

import "github.com/saffage/jet/ast"

func (check *Checker) resolveDecl(decl *ast.Decl) {
	switch {
	case decl.Name.Name == "_":
		check.errorf(decl.Name, "attempt to declare an empty identifier")

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
		check.errorf(decl.Name, "mutable compile-time variables are not supported")

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
					check.errorf(expr, "value is not a constant expression")
					return
				}

				if value.Value == nil {
					// Its a typedesc
					check.resolveTypeAliasDecl(decl)
				} else {
					check.resolveConstDecl(decl)
				}
			}
		}
	}
}
