package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func resolveVar(node *ast.VarDecl, scope *Scope) error {
	// 'tValue' can be nil.
	tValue, err := resolveVarValue(node.Value, scope)
	if err != nil {
		return err
	}

	// 'tType' must be not nil.
	tType, err := resolveVarType(node.Binding.Type, tValue, scope)
	if err != nil {
		return err
	}

	fmt.Printf(">>> var value '%s'\n", tValue)
	fmt.Printf(">>> var type '%s'\n", tType)

	if tValue != nil && !tType.Equals(tValue) {
		return NewErrorf(
			node.Binding.Name,
			"type mismatch, expected '%s', got '%s'",
			tType,
			tValue,
		)
	}

	fmt.Printf(">>> var actual type '%s'\n", tType)
	sym := NewVar(scope, tType, node.Binding, node.Binding.Name)

	if defined := scope.Define(sym); defined != nil {
		return errorAlreadyDefined(sym.Ident(), defined.Ident())
	}

	return nil
}

func resolveVarValue(value ast.Node, scope *Scope) (types.Type, error) {
	if value != nil {
		t, err := scope.TypeOf(value)
		if err != nil {
			return nil, err
		}

		if types.IsTypeDesc(t) {
			return nil, NewErrorf(value, "expected value, got type '%s' instead", t.Underlying())
		}

		return types.SkipUntyped(t), nil
	}

	return nil, nil
}

func resolveVarType(typeExpr ast.Node, value types.Type, scope *Scope) (types.Type, error) {
	if typeExpr != nil {
		t, err := scope.TypeOf(typeExpr)
		if err != nil {
			return t, err
		}

		if typedesc := types.AsTypeDesc(t); typedesc != nil {
			return typedesc.Base(), nil
		}

		return nil, NewError(typeExpr, "expression is not a type")
	}

	return value, nil
}

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
