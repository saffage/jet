package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (check *Checker) builtInMagic(args ast.Node, scope *Scope) *TypedValue {
	argList, _ := args.(*ast.ParenList)
	if argList == nil {
		check.errorf(args, "expected argument list")
		return nil
	}

	arg1, _ := argList.Exprs[0].(*ast.Literal)
	if arg1 == nil || arg1.Kind != ast.StringLiteral {
		check.errorf(argList.Exprs[0], "expected string literal")
		return nil
	}

	switch arg1.Value {
	case "Bool":
		return &TypedValue{types.NewTypeDesc(types.Primitives[types.Bool]), nil}

	case "I32":
		return &TypedValue{types.NewTypeDesc(types.Primitives[types.I32]), nil}

	default:
		check.errorf(arg1, "unknown magic '%s'", arg1.Value)
		return nil
	}
}

func (check *Checker) builtInTypeOf(args ast.Node, scope *Scope) *TypedValue {
	argList, _ := args.(*ast.ParenList)
	if argList == nil {
		check.errorf(args, "expected argument list")
		return nil
	}

	t := check.typeOf(argList.Exprs[0])
	if t == nil {
		return nil
	}

	return &TypedValue{types.NewTypeDesc(types.SkipUntyped(t)), nil}
}

func (check *Checker) builtInPrint(args ast.Node, scope *Scope) *TypedValue {
	argList, _ := args.(*ast.ParenList)
	if argList == nil {
		check.errorf(args, "expected argument list")
		return nil
	}

	t := check.typeOf(argList)
	if t == nil {
		return nil
	}

	return &TypedValue{types.Unit, nil}
}
