package cgen

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

func (gen *generator) BuiltInCall(node *ast.BuiltIn, call *ast.Call) string {
	switch node.Data {
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
		report.Warningf("unknown built-in function '%s'", node.Repr())
		return "ERROR_CGEN"
	}
}

func (gen *generator) builtInPrint(call *ast.Call) string {
	value := gen.Types[call.Args.Nodes[0]]

	switch t := types.SkipAlias(value.Type).(type) {
	case *types.Primitive:
		switch t.Kind() {
		case types.KindUntypedString:
			if value.Value != nil {
				return fmt.Sprintf(
					`fwrite(%s, 1, %d, stdout)`,
					value.Value,
					len(*constant.AsString(value.Value)),
				)
			} else {
				return fmt.Sprintf(
					`fwrite(%[1]s, 1, strlen(%[1]s), stdout)`,
					gen.exprString(call.Args.Nodes[0]),
				)
			}

		case types.KindI8,
			types.KindI16,
			types.KindI32,
			types.KindI64,
			types.KindUntypedInt:
			return fmt.Sprintf(
				`fprintf(stdout, "%%d", %s)`,
				gen.exprString(call.Args.Nodes[0]),
			)

		case types.KindU8,
			types.KindU16,
			types.KindU32,
			types.KindU64:
			return fmt.Sprintf(
				`fprintf(stdout, "%%u", %s)`,
				gen.exprString(call.Args.Nodes[0]),
			)

		case types.KindF32, types.KindF64:
			return fmt.Sprintf(
				`fprintf(stdout, "%%f", %s)`,
				gen.exprString(call.Args.Nodes[0]),
			)

		default:
		}

	case *types.Enum:
		return fmt.Sprintf(
			`fprintf(stdout, "%%d", %s)`,
			gen.exprString(call.Args.Nodes[0]),
		)
	}

	panic(fmt.Sprintf("'$print' for the type %s is not implemented", value.Type))
}

func (gen *generator) builtInPrintln(call *ast.Call) string {
	value := gen.Types[call.Args.Nodes[0]]

	switch t := types.SkipAlias(value.Type).(type) {
	case *types.Primitive:
		switch t.Kind() {
		case types.KindUntypedString:
			if value.Value != nil {
				return fmt.Sprintf(
					`fwrite(%[1]s"\n", 1, %d+1, stdout)`,
					value.Value,
					len(*constant.AsString(value.Value)),
				)
			} else {
				return fmt.Sprintf(
					`fwrite(%[1]s, 1, strlen(%[1]s), stdout); fwrite("\n", 1, 1, stdout)`,
					gen.exprString(call.Args.Nodes[0]),
				)
			}

		case types.KindI8,
			types.KindI16,
			types.KindI32,
			types.KindI64,
			types.KindUntypedInt:
			return fmt.Sprintf(
				`fprintf(stdout, "%%d\n", %s)`,
				gen.exprString(call.Args.Nodes[0]),
			)

		case types.KindU8,
			types.KindU16,
			types.KindU32,
			types.KindU64:
			return fmt.Sprintf(
				`fprintf(stdout, "%%u\n", %s)`,
				gen.exprString(call.Args.Nodes[0]),
			)

		case types.KindF32, types.KindF64:
			return fmt.Sprintf(
				`fprintf(stdout, "%%f\n", %s)`,
				gen.exprString(call.Args.Nodes[0]),
			)

		default:
		}

	case *types.Enum:
		return fmt.Sprintf(
			`fprintf(stdout, "%%d\n", %s)`,
			gen.exprString(call.Args.Nodes[0]),
		)
	}

	panic(fmt.Sprintf("'$println' for the type %s is not implemented", value.Type))
}

func (gen *generator) builtInAssert(call *ast.Call) string {
	expr := gen.exprString(call.Args.Nodes[0])
	return fmt.Sprintf("assert(%s)", expr)
}

func (gen *generator) builtInAsPtr(call *ast.Call) string {
	return gen.exprString(call.Args.Nodes[0])
}

func (gen *generator) builtInCast(call *ast.Call) string {
	t := gen.TypeOf(call.Args.Nodes[0])
	if t == nil {
		panic("unreachable")
	}

	val := call.Args.Nodes[1]
	return fmt.Sprintf("(%s)%s", gen.TypeString(types.SkipTypeDesc(t)), gen.exprString(val))
}

func (gen *generator) builtInSizeOf(call *ast.Call) string {
	val := gen.TypeOf(call.Args.Nodes[0])
	return fmt.Sprintf("sizeof(%s)", gen.TypeString(types.SkipTypeDesc(val)))
}

func (gen *generator) builtInEmit(call *ast.Call) string {
	val := gen.ValueOf(call.Args.Nodes[0])
	return *constant.AsString(val.Value)
}
