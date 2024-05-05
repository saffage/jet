package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/types"
)

func resolveVarDecl(node *ast.VarDecl, scope *Scope) (types.Type, error) {
	var t types.Type

	if node.Binding.Type != nil {
		tType, err := scope.TypeOf(node.Binding.Type)
		if err != nil {
			return tType, err
		}

		assert.Ok(tType != nil, fmt.Sprintf("expression '%s' should have type", node.Binding.Type))
		typedesc, _ := tType.Underlying().(*types.TypeDesc)

		if typedesc == nil {
			return t, NewErrorf(node.Binding.Type, "expression is not a type (%s)", tType)
		}

		t = typedesc.Base()
		fmt.Printf(">>> set `%s` type `%s`\n", node.Binding.Name, t)
	}

	isUnknownSizedArray := false

	if array := types.AsArray(t); array != nil && array.Size() == -1 {
		isUnknownSizedArray = true
	}

	if node.Value != nil {
		tValue, err := scope.TypeOf(node.Value)
		if err != nil {
			return nil, err
		}

		if types.IsTypeDesc(tValue) {
			return nil, NewErrorf(node.Value, "expected value, got type '%s' instead", tValue.Underlying())
		}

		tValue = types.SkipUntyped(tValue)

		if t != nil && !t.Equals(tValue) {
			return t, NewErrorf(node.Value, "type mismatch, expected '%s', got '%s'", t, tValue)
		}

		if array, arrayValue := types.AsArray(t), types.AsArray(tValue); isUnknownSizedArray {
			if array == nil || arrayValue == nil {
				panic("unreachable")
			}

			t = types.NewArray(arrayValue.Size(), array.ElemType())
			fmt.Printf(">>> set `%s` type `%s`\n", node.Binding.Name, t)
		}

		if t == nil {
			t = tValue
			fmt.Printf(">>> set `%s` type `%s`\n", node.Binding.Name, t)
		}
	} else if isUnknownSizedArray {
		return t, NewError(node.Binding.Name, "value is required to infer array size")
	}

	return t, nil
}

// func resolveVarValue(value ast.Node, scope *Scope) (types.Type, error) {
// 	if value == nil {
// 		panic("unreachable")
// 	}

// 	t, err := scope.TypeOf(value)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if types.IsTypeDesc(t) {
// 		return nil, NewErrorf(value, "expected value, got type '%s' instead", t.Underlying())
// 	}

// 	return types.SkipUntyped(t), nil
// }

func resolveFuncDecl(sym *Func) error {
	sig := sym.node.Signature
	tParams := []types.Type{}

	for _, param := range sig.Params.Exprs {
		switch param := param.(type) {
		case *ast.Binding:
			t, err := sym.owner.TypeOf(param.Type)
			if err != nil {
				return err
			}

			t = types.SkipTypeDesc(t)
			tParams = append(tParams, t)

			paramSym := NewVar(sym.scope, t, nil, param.Name)
			sym.scope.Define(paramSym)

			fmt.Printf(">>> set `%s` type `%s`\n", paramSym.Name(), t)
			fmt.Printf(">>> def param `%s`\n", paramSym.Name())

		case *ast.BindingWithValue:
			return NewError(param, "parameters can't have the default value")

		default:
			panic(fmt.Sprintf("ill-formed AST: unexpected node type '%T'", param))
		}
	}

	// Result.

	tResult := types.Unit

	if sig.Result != nil {
		t, err := sym.owner.TypeOf(sig.Result)
		if err != nil {
			return err
		}

		tResult = types.NewTuple(types.SkipTypeDesc(t))
	}

	// Produce function type.

	t := types.NewFunc(tResult, types.NewTuple(tParams...))

	sym.setType(t)
	fmt.Printf(">>> set `%s` type `%s`\n", sym.Name(), t.String())

	// Body.

	if sym.node.Body != nil {
		tBody, err := sym.scope.TypeOf(sym.node.Body)
		if err != nil {
			return err
		}

		if !tResult.Equals(tBody) {
			return NewErrorf(
				sym.node.Body.Nodes[len(sym.node.Body.Nodes)-1],
				"expected expression of type '%s' for function result, got '%s' instead",
				tResult,
				tBody,
			)
		}
	} else {
		return NewError(sym.Ident(), "functions without body is not allowed")
	}

	return nil
}
