package types

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

func (check *checker) resolveLetDecl(node *ast.LetDecl) {
	assert(node.Decl.Name != nil, "let declaration should have a name")

	var t Type

	switch ty := node.Decl.Type.(type) {
	case nil:
		var err error
		t, err = check.typeOf(node.Value)
		check.problem(err)

	case *ast.Upper, *ast.TypeVar:
		tDecl, err := check.typeOf(node.Decl.Type)
		check.problem(err)
		report.Debug("%T - %s", tDecl, tDecl)

		t, err = check.typeOf(node.Value, SkipTypeDesc(tDecl))
		report.Debug("%T - %s", t, t)

		if err, ok := err.(*errorTypeMismatch); ok {
			err.dest = node.Decl.Type
		}

		check.problem(err)

	case *ast.Signature:
		check.env = NewNamedEnv(check.env, node.Decl.Name.String()+" parameters")

		tDecl := check.resolveSignature(ty)
		_, err := check.typeOf(node.Value, tDecl.result)
		check.env = check.env.parent
		check.problem(err)

		t = tDecl

	default:
		panic(&errorIllFormedAst{node.Decl.Type})
	}

	t = IntoTyped(t)

	if t == nil {
		check.errorf(node.Value, "cannot get a type of the expression")
		return
	}

	_, discarded := node.Decl.Name.(*ast.Underscore)
	if discarded && Is[*Func](t) {
		check.problem(&warnDiscardedFuncDef{node.Decl.Name})
	}

	sym := NewBinding(check.env, nil, &Value{t, nil}, node.Decl, node)
	sym.isGlobal = check.env.parent == nil

	check.env.Define(sym)
	check.newDef(node.Decl.Name, sym)
}
