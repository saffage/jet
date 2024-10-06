package types

import (
	"fmt"

	"github.com/saffage/jet/ast"
)

func (check *checker) resolveSignature(sig *ast.Signature) *Function {
	tParams := make([]Type, len(sig.Params.Nodes))
	tResult := Type(nil)

	for i, param := range sig.Params.Nodes {
		label := (*ast.Lower)(nil)

		if labelNode, ok := param.(*ast.Label); ok {
			label = labelNode.Name
			param = labelNode.X
		}

		decl, ok := param.(*ast.Decl)
		if !ok {
			panic(&errorIllFormedAst{param})
		}

		sym, err := check.resolveParam(decl, label)
		check.error(err)
		tParams[i] = sym.Type()

		if prev := check.env.Define(sym); prev != nil {
			check.error(&errorParamAlreadyDefined{decl.Name, prev.Ident()})
			continue
		}

		check.newDef(decl.Name, sym)
	}

	if sig.Result != nil {
		t, err := check.typeOf(sig.Result)
		check.error(err)

		if err == nil {
			assert(t != nil)
			assert(Is[*TypeDesc](t), fmt.Sprintf("%T - %s", t, t))

			tResult = SkipTypeDesc(t)

			assert(!IsUntyped(tResult))
		}
	}

	return NewFunction(tParams, tResult, nil)
}

func (check *checker) resolveParam(
	param *ast.Decl,
	label *ast.Lower,
) (sym *Binding, err error) {
	assert(param != nil)

	sym = NewBinding(check.env, nil, &Value{nil, nil}, param, nil)
	sym.label = label
	sym.isParam = true

	if param.TypeTok.IsValid() {
		err = internalErrorf(param, "type parameters is not implemented")
		return
	}

	if param.Type == nil {
		err = internalErrorf(param, "parameter type inference is not implemented")
		return
	}

	sym.value.T, err = check.typeOf(param.Type)

	if err != nil {
		return
	}

	sym.value.T = SkipTypeDesc(sym.value.T)

	assert(sym.value.T != nil)
	return
}
