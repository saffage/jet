package cgen

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) enumDecl(sym *checker.Enum) {
	buf := strings.Builder{}
	buf.WriteString("typedef enum ")
	buf.WriteString(sym.Name())
	buf.WriteString(" {\n")
	gen.numIndent++

	enumName := gen.name(sym)

	for i, field := range types.SkipTypeDesc(sym.Type()).(*types.Enum).Fields() {
		gen.indent(&buf)
		buf.WriteString(fmt.Sprintf("%s = %d,\n", enumName+"__"+field, i))
	}

	gen.numIndent--
	buf.WriteString("} ")
	buf.WriteString(enumName)
	buf.WriteString(";\n\n")
	gen.typeSect.WriteString(buf.String())
}
