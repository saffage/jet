package cgen

import (
	"fmt"
	"reflect"
	"strconv"

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

	switch t := value.Type.(type) {
	case *types.Primitive:
		switch t.Kind() {
		case types.KindUntypedString:
			return fmt.Sprintf(
				"fwrite(%[1]s, 1, sizeof(%[1]s), stdout)",
				"str_lit_"+strconv.Itoa(int(reflect.ValueOf(argList.Exprs[0]).Pointer())),
			)

		default:
			panic("not implemented")
		}

	default:
		panic("not implemented")
	}
}

func (gen *Generator) builtInAssert(call *ast.BuiltInCall) string {
	expr := gen.ExprString(call.Args.(*ast.ParenList).Exprs[0])
	return fmt.Sprintf("assert(%s)", expr)
}
