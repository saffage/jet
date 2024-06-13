package cgen

import (
	"strings"

	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) enumDecl(sym *checker.Enum) {
	buf := strings.Builder{}
	gen.flinef(&buf, "typedef enum %s {\n", sym.Name())
	gen.indent++
	enumName := gen.name(sym)
	for i, field := range types.SkipTypeDesc(sym.Type()).(*types.Enum).Fields() {
		gen.flinef(&buf, "%s = %d,\n", enumName+"__"+field, i)
	}
	gen.indent--
	gen.flinef(&buf, "} %s;\n", enumName)
	gen.typeSect.WriteString(buf.String())
}
