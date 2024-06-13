package cgen

import (
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (gen *generator) arrayInit(prefix string, node *ast.BracketList, _ *types.Array) {
	for i, expr := range node.Nodes {
		gen.linef("%s[%d] = %s;\n", prefix, i, gen.exprString(expr))
	}
}

func (gen *generator) arrayAssign(dest string, src ast.Node, ty *types.Array) {
	if _, ok := src.(*ast.Call); ok {
		expr := gen.exprString(src)
		gen.linef("%s;\n", strings.Replace(expr, "/*RESULT*/", dest, -1))
	} else if lit, ok := src.(*ast.BracketList); ok {
		gen.arrayInit(dest, lit, ty)
	} else {
		gen.linef(
			"memcpy((void*)%s, (const void*)%s, sizeof(%s));\n",
			dest,
			gen.exprString(src),
			gen.TypeString(ty),
		)
	}
}

// func (gen *generator) arrayLit(node *ast.BracketList, _ *types.Array) string {
// 	if len(node.Nodes) == 0 {
// 		return "{}"
// 	}
// 	buf := strings.Builder{}
// 	buf.WriteString("{\n")
// 	gen.numIndent++
// 	for i, elem := range node.Nodes {
// 		if i != 0 {
// 			buf.WriteString(",\n")
// 			gen.indent(&buf)
// 		} else {
// 			gen.indent(&buf)
// 		}
// 		buf.WriteString(gen.exprString(elem))
// 	}
// 	buf.WriteString("\n")
// 	gen.numIndent--
// 	gen.indent(&buf)
// 	buf.WriteString("}")
// 	return buf.String()
// }
