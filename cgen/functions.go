package cgen

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) funcDecl(sym *checker.Func) {
	t := sym.Type().(*types.Func)
	declBuf := strings.Builder{}
	tResultVar := types.Type(nil)

	if sym.Name() == "main" {
		gen.codeSect.WriteString(fnMainHead)
	} else {
		if sym.IsExtern() {
			declBuf.WriteString("extern ")
		}

		result := t.Result()

		if result.Len() == 0 {
			declBuf.WriteString("void")
		} else if result.Len() == 1 {
			declBuf.WriteString(gen.TypeString(result.Underlying()))
			tResultVar = result.Underlying()
		} else {
			gen.errorf(sym, "tuple are not supported")
			declBuf.WriteString("ERROR_CGEN__FUNC_TUPLE_RESULT")
		}

		declBuf.WriteByte(' ')
		declBuf.WriteString(gen.name(sym))
		declBuf.WriteByte('(')

		// Gen params.
		if len(sym.Params()) == 0 {
			declBuf.WriteString("void")
		} else {
			for i, param := range sym.Params() {
				if i != 0 {
					declBuf.WriteString(", ")
				}

				declBuf.WriteString(gen.TypeString(param.Type()))
				declBuf.WriteByte(' ')
				declBuf.WriteString(gen.name(param))
			}
		}

		declBuf.WriteByte(')')

		gen.declFnsSect.WriteString(declBuf.String())
		gen.declFnsSect.WriteString(";\n")

		if sym.IsExtern() {
			return
		}

		gen.codeSect.WriteString(declBuf.String())
	}

	node := sym.Node().(*ast.FuncDecl)

	gen.codeSect.WriteString(" {\n")
	gen.numIndent++

	if tResultVar != nil {
		gen.indent(&gen.codeSect)
		gen.codeSect.WriteString(gen.TypeString(tResultVar))
		gen.codeSect.WriteString(" __result;\n\n")
	}

	if sym.Name() == "main" {
		gen.indent(&gen.codeSect)
		gen.codeSect.WriteString(fmt.Sprintf("init%s();\n", gen.Module.Name()))
	}

	for i, stmt := range node.Body.Nodes {
		gen.indent(&gen.codeSect)

		if tResultVar != nil && i == len(node.Body.Nodes)-1 {
			gen.codeSect.WriteString("__result = ")
			gen.codeSect.WriteString(gen.ExprString(stmt))
			gen.codeSect.WriteString(";\n")
		} else {
			gen.codeSect.WriteString(gen.StmtString(stmt))
		}
	}

	// gen.indent(&gen.codeSect)
	// gen.codeSect.WriteString("goto L_ret;\n")
	// gen.codeSect.WriteString("\nL_ret:;\n")

	if tResultVar != nil {
		gen.indent(&gen.codeSect)
		gen.codeSect.WriteString("return __result;\n")
	}

	gen.numIndent--
	gen.codeSect.WriteString("}\n")
}
