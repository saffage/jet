package cgen

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) varDecl(sym *checker.Var) {
	t := gen.TypeString(sym.Type())
	gen.declVarsSect.WriteString(fmt.Sprintf("%s %s;\n", t, gen.name(sym)))
}

func (gen *generator) initFunc() string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("void init%s(void) {\n", gen.Module.Name()))
	gen.numIndent++

	for def := gen.Defs.Front(); def != nil; def = def.Next() {
		def := def.Value

		if _var, _ := def.(*checker.Var); _var != nil && _var.IsGlobal() && _var.Value() != nil {
			gen.indent(&buf)
			buf.WriteString(gen.binary(
				_var.Node().(*ast.Binding).Name,
				_var.Value(),
				types.Unit,
				ast.OperatorAssign,
			))
			buf.WriteString(";\n")
		}
	}

	gen.numIndent--
	buf.WriteString("}\n")
	return buf.String()
}
