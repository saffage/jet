package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

// [checker.scope] must be set to the parameters scope.
func (check *checker) resolveSignature(sig *ast.Signature) (types.Type, bool) {
	var (
		tParams = make([]types.Type, 0, len(sig.Params.Nodes))
		tResult = types.Unit
	)

	resolved := true

	for _, node := range sig.Params.Nodes {
		var label *ast.Name

		if labelNode, ok := node.(*ast.Label); ok {
			label = labelNode.Label
			node = labelNode.X
		}

		decl, isDecl := node.(*ast.Decl)

		if !isDecl {
			panic(fmt.Sprintf(
				"ill-formed AST: unexpected node type '%T'",
				node,
			))
		}

		if sym, paramResolved := check.resolveParam(decl); paramResolved {
			sym.label = label

			if prev := check.scope.Define(sym); prev != nil {
				check.addError(
					errorAlreadyDefined(sym.Ident(), prev.Ident()),
				)
				resolved = false
			}

			tParams = append(tParams, sym.Type())
		} else {
			resolved = true
		}
	}

	if sig.Result != nil {
		if tResultActual := check.typeOf(sig.Result); tResultActual != nil {
			tResultActual = types.SkipTypeDesc(tResultActual)
			assert(!types.IsUntyped(tResultActual))
			tResult = types.WrapInTuple(types.SkipTypeDesc(tResultActual))
			// hasResult = true
		} else {
			resolved = false
		}
	}

	return types.NewFunc(types.NewTuple(tParams...), tResult, nil), resolved
}

// Result mey have no type.
func (check *checker) resolveParam(param *ast.Decl) (*Binding, bool) {
	if param.Type == nil {
		check.errorf(param, "generics are not implemented")
		return nil, false
	}

	var tParam types.Type

	if tParam = check.typeOf(param.Type); tParam == nil {
		// TODO: support for generics.
		check.errorf(param.Type, "cannot get a type")
		return nil, false
	} else {
		tParam = types.SkipTypeDesc(tParam)
	}

	paramSym := NewBinding(check.scope, tParam, param, nil)
	paramSym.isParam = true

	if defined := check.scope.Define(paramSym); defined != nil {
		check.errorf(
			param,
			"parameter with the same name was already defined",
		)
		return nil, false
	}

	check.newDef(param.Name, paramSym)
	// report.TaggedDebugf("checker", "func: def param: %s", paramSym.Name())
	// report.TaggedDebugf("checker", "func: set param type: %s", tParam)
	return paramSym, true
}
