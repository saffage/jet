package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/types"
)

func builtInBuiltin(node *ast.Parens, args []*TypedValue) (*TypedValue, error) {
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

	case "Char":
		return &TypedValue{types.NewTypeDesc(types.Char), nil}, nil

	case "Pointer":
		return &TypedValue{types.NewTypeDesc(types.Pointer), nil}, nil

	case "String":
		return &TypedValue{types.NewTypeDesc(types.String), nil}, nil

	default:
		return nil, newErrorf(node.Nodes[0], "unknown built-in '%s'", *strval)
	}
}

func builtInTypeOf(node *ast.Parens, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{
		Type:  types.NewTypeDesc(types.SkipUntyped(args[0].Type)),
		Value: nil,
	}, nil
}

func builtInPrint(node *ast.Parens, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{types.Unit, nil}, nil
}

func builtInAssert(node *ast.Parens, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{types.Unit, nil}, nil
}

func builtInAsPtr(node *ast.Parens, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{types.NewRef(types.U8), nil}, nil
}

func builtInCast(node *ast.Parens, args []*TypedValue) (*TypedValue, error) {
	// TODO some additional checks
	return &TypedValue{types.SkipTypeDesc(args[0].Type), nil}, nil
}

func builtInSizeOf(node *ast.Parens, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{types.U64, nil}, nil
}

func builtInEmit(node *ast.Parens, args []*TypedValue) (*TypedValue, error) {
	return &TypedValue{types.Unit, nil}, nil
}
