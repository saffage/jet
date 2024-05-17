package cgen

import (
	"strings"

	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) structDecl(sym *checker.Struct) {
	buf := strings.Builder{}
	buf.WriteString("typedef struct ")
	buf.WriteString(sym.Name())
	buf.WriteString(" {\n")
	gen.numIndent++

	for _, field := range types.SkipTypeDesc(sym.Type()).(*types.Struct).Fields() {
		gen.indent(&buf)
		buf.WriteString(gen.TypeString(field.Type))
		buf.WriteByte(' ')
		buf.WriteString(field.Name)
		buf.WriteString(";\n")
	}

	gen.numIndent--
	buf.WriteString("} ")
	buf.WriteString(gen.name(sym))
	buf.WriteString(";\n\n")
	gen.typeSect.WriteString(buf.String())
}
