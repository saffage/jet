package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/types"
)

func (check *Checker) builtInMagic(node *ast.ParenList, args []*TypedValue) *TypedValue {
	strval := constant.AsString(args[0].Value)

	if strval == nil {
		panic("unreachable")
	}

	switch *strval {
	case "Bool":
		return &TypedValue{types.NewTypeDesc(types.Bool), nil}

	case "I32":
		return &TypedValue{types.NewTypeDesc(types.I32), nil}

	default:
		check.errorf(node.Exprs[0], "unknown magic '%s'", *strval)
		return nil
	}
}

func (check *Checker) builtInTypeOf(node *ast.ParenList, args []*TypedValue) *TypedValue {
	return &TypedValue{
		Type:  types.NewTypeDesc(types.SkipUntyped(args[0].Type)),
		Value: nil,
	}
}

func (check *Checker) builtInPrint(node *ast.ParenList, args []*TypedValue) *TypedValue {
	return &TypedValue{types.Unit, nil}
}

func (check *Checker) builtInAssert(node *ast.ParenList, args []*TypedValue) *TypedValue {
	return &TypedValue{types.Unit, nil}
}
