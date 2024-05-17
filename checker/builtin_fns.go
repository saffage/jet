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

	case "I8":
		return &TypedValue{types.NewTypeDesc(types.I8), nil}, nil

	case "I16":
		return &TypedValue{types.NewTypeDesc(types.I16), nil}, nil

	case "I32":
		return &TypedValue{types.NewTypeDesc(types.I32), nil}, nil

	case "I64":
		return &TypedValue{types.NewTypeDesc(types.I64), nil}, nil

	case "U8":
		return &TypedValue{types.NewTypeDesc(types.U8), nil}, nil

	case "U16":
		return &TypedValue{types.NewTypeDesc(types.U16), nil}, nil

	case "U32":
		return &TypedValue{types.NewTypeDesc(types.U32), nil}, nil

	case "U64":
		return &TypedValue{types.NewTypeDesc(types.U64), nil}, nil

	case "F32":
		return &TypedValue{types.NewTypeDesc(types.F32), nil}, nil

	case "F64":
		return &TypedValue{types.NewTypeDesc(types.F64), nil}, nil

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

func builtInAsPtr(node *ast.ParenList, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{types.NewRef(types.U8), nil}, nil
}
