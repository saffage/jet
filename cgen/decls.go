package cgen

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) decl(sym checker.Symbol) {
	node := sym.Node().(*ast.Decl)
	ty := sym.Type()
	gen.linef("%s %s;\n", gen.TypeString(ty), gen.name(sym))

	if node.Value != nil {
		if array := types.AsArray(ty); array != nil {
			gen.arrayAssign(gen.name(sym), node.Value, array)
			// gen.line(";\n")
		} else if _struct := types.AsStruct(ty); _struct != nil {
			gen.structAssign(gen.name(sym), node.Value, _struct)
		} else {
			gen.linef("%s;\n",
				gen.binary(
					node.Name,
					node.Value,
					types.Unit,
					ast.OperatorAssign,
				),
			)
		}
	}
}
