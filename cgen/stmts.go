package cgen

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (gen *generator) StmtString(stmt ast.Node) string {
	buf := strings.Builder{}

	switch stmt := stmt.(type) {
	case *ast.VarDecl:
		if sym, _ := gen.Defs.Get(stmt.Binding.Name); sym != nil {
			buf.WriteString(gen.TypeString(sym.Type()))
			buf.WriteString(" " + gen.name(sym) + ";\n")

			if stmt.Value != nil {
				gen.indent(&buf)
				buf.WriteString(gen.binary(
					stmt.Binding.Name,
					stmt.Value,
					types.Unit,
					ast.OperatorAssign,
				))
				buf.WriteString(";\n")
			}
		} else {
			panic("unreachable")
		}

	case *ast.While:
		buf.WriteString(fmt.Sprintf("while (%s) {\n", gen.ExprString(stmt.Cond)))
		gen.numIndent++
		for _, stmt := range stmt.Body.Nodes {
			gen.indent(&buf)
			buf.WriteString(gen.StmtString(stmt))
		}
		gen.numIndent--
		gen.indent(&buf)
		buf.WriteString("}\n")

	case *ast.If:
		buf.WriteString(fmt.Sprintf("if (%s) {\n", gen.ExprString(stmt.Cond)))
		gen.numIndent++
		for _, stmt := range stmt.Body.Nodes {
			gen.indent(&buf)
			buf.WriteString(gen.StmtString(stmt))
		}
		gen.numIndent--
		gen.indent(&buf)
		buf.WriteString("}\n")

	case *ast.Break:
		return "break;\n"

	case *ast.Continue:
		return "continue;\n"

	default:
		return gen.ExprString(stmt) + ";\n"
	}

	return buf.String()
}
