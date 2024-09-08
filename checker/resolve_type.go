package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

func (check *checker) resolveTypeDecl(node *ast.TypeDecl) {
	if node.Args != nil {
		check.errorf(node.Args, "generics are not implemented")
		return
	}

	if node.Expr == nil {
		check.resolveExternalTypeDecl(node)
		return
	}

	panic("unimplemented")
}

// 'node.Expr == nil'
//
// Actual type cannot be resolved, so just create a symbol for it.
//
// Also this function resolves built-in types.
func (check *checker) resolveExternalTypeDecl(node *ast.TypeDecl) {
	// TODO: resolve built-in types.
	// TODO: create a special type for external types.
	sym := NewType(check.scope, nil, node)

	if prev := check.scope.Define(sym); prev != nil {
		check.addError(errorAlreadyDefined(node.Name, prev.Ident()))
	}

	report.TaggedDebugf(
		"checker",
		"defined type %s at %s: %+v",
		sym.Name(),
		sym.Ident().Pos(),
		sym.Type(),
	)
}
