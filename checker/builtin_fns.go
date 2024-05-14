package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/types"
)

func builtInMagic(node *ast.ParenList, args []*TypedValue) (*TypedValue, error) {
	strval := constant.AsString(args[0].Value)

	if strval == nil {
		panic("unreachable")
	}

	switch *strval {
	case "Bool":
		return &TypedValue{types.NewTypeDesc(types.Bool), nil}, nil

	case "I32":
		return &TypedValue{types.NewTypeDesc(types.I32), nil}, nil

	case "U8":
		return &TypedValue{types.NewTypeDesc(types.U8), nil}, nil

	case "String":
		return &TypedValue{types.NewTypeDesc(types.String), nil}, nil

	default:
		return nil, NewErrorf(node.Exprs[0], "unknown magic '%s'", *strval)
	}
}

func builtInTypeOf(node *ast.ParenList, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{
		Type:  types.NewTypeDesc(types.SkipUntyped(args[0].Type)),
		Value: nil,
	}, nil
}

func builtInPrint(node *ast.ParenList, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{types.Unit, nil}, nil
}

func builtInAssert(node *ast.ParenList, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{types.Unit, nil}, nil
}
