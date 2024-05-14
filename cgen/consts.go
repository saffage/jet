package cgen

import (
	"fmt"

	"github.com/saffage/jet/checker"
)

func (gen *Generator) Const(sym *checker.Const) {
	gen.dataSect.WriteString(fmt.Sprintf(
		"#define %[1]s %[2]s // constant\n",
		sym.Name(),
		sym.Value(),
	))
}