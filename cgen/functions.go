package cgen

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) fn(sym *checker.Func) {
	defer gen.setScope(gen.scope)
	gen.setScope(sym.Local())
	gen.funcTempVarId = 0
	gen.funcLabelID = 0
	decl := ""
	tyResult := types.Type(nil)

	if sym.Name() == "main" {
		gen.line(fnMainHead)
	} else {
		decl, tyResult = gen.fnDecl(sym)
		gen.flinef(&gen.declFnsSect, "%s;\n", decl)
	}

	if !sym.IsExtern() {
		gen.linef("%s\n", decl)
		gen.fnDef(sym, tyResult)
	}
}

func (gen *generator) fnDecl(sym *checker.Func) (decl string, tyResult types.Type) {
	buf := strings.Builder{}
	ty := types.AsFunc(sym.Type())
	isArrayResult := false

	if sym.IsExtern() {
		gen.fline(&buf, "extern ")
	}

	result := ty.Result()

	if result.Len() == 0 {
		buf.WriteString("void")
	} else if result.Len() == 1 {
		ty := result.Types()[0]
		if !types.IsArray(ty.Underlying()) {
			buf.WriteString(gen.TypeString(ty.Underlying()))
		} else {
			isArrayResult = true
			buf.WriteString("void")
		}
		tyResult = ty.Underlying()
	} else {
		gen.errorf(sym, "tuples are not supported")
		buf.WriteString("ERROR_CGEN__FUNC_TUPLE_RESULT")
	}

	buf.WriteByte(' ')

	if sym.IsExtern() && sym.ExternName() != "" {
		buf.WriteString(sym.ExternName())
	} else {
		buf.WriteString(gen.name(sym))
	}

	buf.WriteByte('(')

	// Gen params.
	if len(sym.Params()) == 0 {
		if isArrayResult {
			buf.WriteString(fmt.Sprintf(
				"%s __result",
				gen.TypeString(tyResult),
			))
		} else {
			buf.WriteString("void")
		}
	} else {
		for i, param := range sym.Params() {
			if i != 0 {
				buf.WriteString(", ")
			}

			// TODO this is not a valic place for the const qualifier,
			// but currently its here for making `const char*` param.
			if checker.HasAttribute(param, "const_c") {
				buf.WriteString("const ")
			}

			buf.WriteString(gen.TypeString(param.Type()))
			buf.WriteByte(' ')
			buf.WriteString(gen.name(param))
		}

		if isArrayResult {
			buf.WriteString(fmt.Sprintf(
				", %s __result",
				gen.TypeString(tyResult),
			))
		}

		if sym.Variadic() != nil {
			if checker.HasAttribute(sym, "extern_c") {
				buf.WriteString(", ...")
			} else {
				panic("variadic functions that is not extern is not implemented")
			}
		}
	}

	buf.WriteByte(')')
	decl = buf.String()
	return
}

func (gen *generator) fnDef(sym *checker.Func, tyResult types.Type) {
	node := sym.Node().(*ast.Decl)
	value, _ := node.Value.(*ast.Function)

	gen.line("{\n")
	gen.indent++

	resultVar := gen.resultVar(tyResult)

	if sym.Name() == "main" {
		gen.linef("init%s();\n", gen.Module.Name())
	}

	if list, _ := value.Body.(*ast.CurlyList); list != nil {
		gen.block(list.StmtList, resultVar)
	} else if resultVar != nil {
		gen.assign("__result", value.Body)
	}

	if resultVar != nil && !types.IsArray(tyResult) {
		gen.line("return __result;\n")
	}

	gen.indent--
	gen.line("}\n")
}
