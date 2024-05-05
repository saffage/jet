package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func builtInMagic(args ast.Node, scope *Scope) (*Value, error) {
	argList, ok := args.(*ast.ParenList)
	if !ok {
		return nil, NewError(args, "expected argument list")
	}

	arg1, ok := argList.Exprs[0].(*ast.Literal)
	if !ok {
		return nil, NewError(argList.Exprs[0], "expected string literal")
	}

	// tArg1, err := scope.TypeOf(argList.Exprs[0])
	// if err != nil {
	// 	return nil, NewError(argList.Exprs[0], "expected literal")
	// }

	// if !types.Primitives[types.UntypedString].Equals(tArg1) {
	// 	return nil, NewErrorf(
	// 		argList.Exprs[0],
	// 		"expected 'untyped string', got '%s' instead",
	// 		tArg1,
	// 	)
	// }

	switch arg1.Value {
	case "Bool":
		return &Value{types.NewTypeDesc(types.Primitives[types.Bool]), nil}, nil

	case "I32":
		return &Value{types.NewTypeDesc(types.Primitives[types.I32]), nil}, nil

	default:
		return nil, NewErrorf(arg1, "unknown magic '%s'", arg1.Value)
	}
}

func builtInTypeOf(args ast.Node, scope *Scope) (*Value, error) {
	argList, ok := args.(*ast.ParenList)
	if !ok {
		return nil, NewError(args, "expected argument list")
	}

	arg1 := argList.Exprs[0]

	t, err := scope.TypeOf(arg1)
	if err != nil {
		return nil, err
	}

	return &Value{types.NewTypeDesc(types.SkipUntyped(t)), nil}, nil
}

func builtInPrint(args ast.Node, scope *Scope) (*Value, error) {
	argList, ok := args.(*ast.ParenList)
	if !ok {
		return nil, NewError(args, "expected argument list")
	}

	_, err := scope.TypeOf(argList)
	if err != nil {
		return nil, err
	}

	return &Value{types.Unit, nil}, nil
}
