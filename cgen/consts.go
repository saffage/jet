package cgen

import (
	"fmt"

	"github.com/saffage/jet/checker"
)

func (gen *generator) constDecl(sym *checker.Const) {
	gen.dataSect.WriteString(fmt.Sprintf(
		"#define %[1]s %[2]s // constant\n",
		gen.name(sym),
		sym.Value(),
	))
}
