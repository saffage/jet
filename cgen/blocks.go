package cgen

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
)

func (gen *generator) block(list *ast.StmtList, result *checker.Var) {
	gen.line("{\n")
	gen.indent++
	for _, stmt := range list.Nodes[:len(list.Nodes)-1] {
		gen.codeSect.WriteString(gen.StmtString(stmt))
	}
	if result != nil {
		gen.assign(gen.name(result), list.Nodes[len(list.Nodes)-1])
	} else {
		gen.codeSect.WriteString(gen.StmtString(list.Nodes[len(list.Nodes)-1]))
	}
	gen.indent--
	gen.line("}\n")
}
