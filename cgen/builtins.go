package cgen

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (gen *Generator) BuiltInCall(call *ast.BuiltInCall) string {
	switch call.Name.Name {
	case "print":
		return gen.builtInPrint(call)

	case "assert":
		return gen.builtInAssert(call)

	default:
		return "ERROR_CGEN"
	}
}

func (gen *Generator) builtInPrint(call *ast.BuiltInCall) string {
	argList, _ := call.Args.(*ast.ParenList)
	value := gen.Types[argList.Exprs[0]]

	switch t := types.SkipAlias(value.Type).(type) {
	case *types.Primitive:
		switch t.Kind() {
		case types.KindUntypedString:
			if value.Value != nil {
				return fmt.Sprintf(
					"fwrite(%[1]s, 1, sizeof(%[1]s), stdout)",
					value.Value,
				)
			} else {
				return fmt.Sprintf(
					"fwrite(%[1]s, 1, sizeof(%[1]s), stdout)",
					gen.ExprString(argList.Exprs[0]),
				)
			}

		case types.KindI8,
			types.KindI16,
			types.KindI32,
			types.KindI64,
			types.KindU8,
			types.KindU16,
			types.KindU32,
			types.KindU64,
			types.KindUntypedInt:
			return fmt.Sprintf(
				`fprintf(stdout, "%%d", %s)`,
				gen.ExprString(argList.Exprs[0]),
			)

		default:
			panic("not implemented")
		}

	default:
		panic(fmt.Sprintf("printing type %T is not implemented", value.Type))
	}
}

func (gen *Generator) builtInAssert(call *ast.BuiltInCall) string {
	expr := gen.ExprString(call.Args.(*ast.ParenList).Exprs[0])
	return fmt.Sprintf("assert(%s)", expr)
}
