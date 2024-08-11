package checker

import (
	"unicode"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (check *Checker) resolveDecl(decl *ast.Decl) {
	switch {
	case decl.Ident.Name == "_":
		// TODO do not introduce a new symbol and do
		// type checking instead of the error
		check.errorf(decl.Ident, "attempt to declare an empty identifier")

	case unicode.IsUpper([]rune(decl.Ident.Name)[0]), FindAttr(decl.Attrs, "comptime") != nil:
		switch expr := decl.Value.(type) {
		case nil:
			// TODO?

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

	case decl.Mut.IsValid():
		check.errorf(decl.Ident, "mutable compile-time variables are not supported")

	default:
		if fn, _ := decl.Value.(*ast.Function); fn != nil {
			// TODO check is we are in the module context
			check.resolveFuncDecl(decl, fn)
		} else if sig, _ := decl.Type.(*ast.Signature); sig != nil && decl.Value == nil {
			check.resolveFuncDecl(decl, &ast.Function{
				Signature: sig,
				Body:      nil,
			})
		} else {
			check.resolveVarDecl(decl)
		}
	}
}
