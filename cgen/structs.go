package cgen

import (
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *Generator) structure(sym *checker.Struct) {
	gen.typeSect.WriteString("typedef struct ")
	gen.typeSect.WriteString(sym.Name())
	gen.typeSect.WriteString(" {\n")
	gen.numIndent++

	for name, tField := range types.SkipTypeDesc(sym.Type()).(*types.Struct).Fields() {
		gen.indent(&gen.typeSect)
		gen.typeSect.WriteString(gen.TypeString(tField))
		gen.typeSect.WriteByte(' ')
		gen.typeSect.WriteString(name)
		gen.typeSect.WriteString(";\n")
	}

	gen.numIndent--
	gen.typeSect.WriteString("} Ty")
	gen.typeSect.WriteString(sym.Name())
	gen.typeSect.WriteString(";\n\n")
}
