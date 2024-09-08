package checker

import (
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (check *checker) resolveLetDecl(node *ast.LetDecl) {
	var t types.Type
	_ = t

	if node.Decl.Type != nil {
		t = check.typeOf(node.Decl.Type)
	}

	name := node.Decl.Name.String()

	// For functions type is never nil, but for anonymous functions it can be.
	// if t != nil && types.IsFunc(t) {
	// 	// Type can be an alias.
	// 	if sig, ok := node.Decl.Type.(*ast.Signature); ok {
	// 		check.resolveSignature(sig)
	// 	} else {
	// 		panic("unimplemented: function type alias checking")
	// 	}
	// }

	if strings.HasPrefix(name, "_") {
		// TODO: don't emit a symbol
	}
}
