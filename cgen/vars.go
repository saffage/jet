package cgen

import (
	"fmt"

	"github.com/saffage/jet/checker"
)

func (gen *Generator) Var(sym *checker.Var) {
	t := gen.TypeString(sym.Type())
	gen.declVarsSect.WriteString(fmt.Sprintf("%s %s;\n", t, sym.Name()))
}
