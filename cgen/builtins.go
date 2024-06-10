package cgen

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

func (gen *generator) BuiltInCall(call *ast.BuiltInCall) string {
	switch call.Name.Name {
	case "print":
		return gen.builtInPrint(call)

	case "println":
		return gen.builtInPrintln(call)

	case "assert":
		return gen.builtInAssert(call)

	case "as_ptr":
		return gen.builtInAsPtr(call)

	case "cast":
		return gen.builtInCast(call)

	case "size_of":
		return gen.builtInSizeOf(call)

	case "emit":
		return gen.builtInEmit(call)

	default:
		report.Warningf("unknown built-in: '@%s'", call.Name.Name)
		return "ERROR_CGEN"
	}
}

func (gen *generator) builtInPrint(call *ast.BuiltInCall) string {
	argList, _ := call.Args.(*ast.ParenList)
	value := gen.Types[argList.Nodes[0]]

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
					gen.exprString(argList.Nodes[0]),
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
				gen.exprString(argList.Nodes[0]),
			)

		default:
			panic("not implemented")
		}

	case *types.Enum:
		return fmt.Sprintf(
			`fprintf(stdout, "%%d", %s)`,
			gen.exprString(argList.Nodes[0]),
		)

	default:
		panic(fmt.Sprintf("printing type %T is not implemented", value.Type))
	}
}

func (gen *generator) builtInPrintln(call *ast.BuiltInCall) string {
	argList, _ := call.Args.(*ast.ParenList)
	value := gen.Types[argList.Nodes[0]]

	switch t := types.SkipAlias(value.Type).(type) {
	case *types.Primitive:
		switch t.Kind() {
		case types.KindUntypedString:
			if value.Value != nil {
				return fmt.Sprintf(
					`fwrite(%[1]s"\n", 1, sizeof(%[1]s"\n"), stdout)`,
					value.Value,
				)
			} else {
				return fmt.Sprintf(
					`fwrite(%[1]s"\n", 1, sizeof(%[1]s"\n"), stdout)`,
					gen.exprString(argList.Nodes[0]),
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
				`fprintf(stdout, "%%d\n", %s)`,
				gen.exprString(argList.Nodes[0]),
			)

		default:
			panic("not implemented")
		}

	case *types.Enum:
		return fmt.Sprintf(
			`fprintf(stdout, "%%d\n", %s)`,
			gen.exprString(argList.Nodes[0]),
		)

	default:
		panic(fmt.Sprintf("printing type %T is not implemented", value.Type))
	}
}

func (gen *generator) builtInAssert(call *ast.BuiltInCall) string {
	expr := gen.exprString(call.Args.(*ast.ParenList).Nodes[0])
	return fmt.Sprintf("assert(%s)", expr)
}

func (gen *generator) builtInAsPtr(call *ast.BuiltInCall) string {
	return gen.exprString(call.Args.(*ast.ParenList).Nodes[0])
}

func (gen *generator) builtInCast(call *ast.BuiltInCall) string {
	t := gen.TypeOf(call.Args.(*ast.ParenList).Nodes[0])
	if t == nil {
		panic("unreachable")
	}

	val := call.Args.(*ast.ParenList).Nodes[1]
	return fmt.Sprintf("(%s)%s", gen.TypeString(types.SkipTypeDesc(t)), gen.exprString(val))
}

func (gen *generator) builtInSizeOf(call *ast.BuiltInCall) string {
	val := gen.TypeOf(call.Args.(*ast.ParenList).Nodes[0])
	return fmt.Sprintf("sizeof(%s)", gen.TypeString(types.SkipTypeDesc(val)))
}

func (gen *generator) builtInEmit(call *ast.BuiltInCall) string {
	val := gen.ValueOf(call.Args.(*ast.ParenList).Nodes[0])
	return *constant.AsString(val.Value)
}
