package cgen

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) fn(sym *checker.Func) {
	prevScope := gen.Scope
	gen.Scope = sym.Local()
	gen.funcTempVarId = 0
	decl := ""
	tyResult := types.Type(nil)

	if sym.Name() == "main" {
		gen.line(fnMainHead)
	} else {
		decl, tyResult = gen.fnDecl(sym)
		gen.flinef(&gen.declFnsSect, "%s;\n", decl)
	}

	if !sym.IsExtern() {
		gen.fnDef(sym, tyResult, decl)
	}

	gen.Scope = prevScope
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
				"%s result",
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
				", %s result",
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

func (gen *generator) fnDef(sym *checker.Func, tyResult types.Type, decl string) {
	isArrayResult := types.IsArray(tyResult)
	gen.line(decl)

	node := sym.Node().(*ast.Decl)
	value, _ := node.Value.(*ast.Function)

	var body []ast.Node

	if list, _ := value.Body.(*ast.CurlyList); list != nil {
		body = list.StmtList.Nodes
	} else {
		body = []ast.Node{value.Body}
	}

	gen.codeSect.WriteString(" {\n")
	gen.indent++

	if tyResult != nil && !isArrayResult {
		gen.linef("%s result;\n", gen.TypeString(tyResult))
	}

	if sym.Name() == "main" {
		gen.linef("init%s();\n", gen.Module.Name())
	}

	for i, stmt := range body {
		if tyResult != nil && i == len(body)-1 {
			if isArrayResult {
				gen.arrayAssign("result", stmt, types.AsArray(tyResult))
			} else {
				gen.linef("result = %s;\n", gen.exprString(stmt))
			}
		} else {
			gen.codeSect.WriteString(gen.StmtString(stmt))
		}
	}

	if tyResult != nil && !isArrayResult {
		gen.line("return result;\n")
	}

	gen.indent--
	gen.line("}\n")
}
