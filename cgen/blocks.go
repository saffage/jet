package cgen

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
)

func (gen *generator) block(list *ast.StmtList, result *checker.Var) {
	gen.line("{\n")
	gen.indent++

	if _, ok := scopeIDs[gen.scope]; !ok {
		scopeIDs[gen.scope] = 0
	} else {
		scopeIDs[gen.scope]++
	}

	id := scopeIDs[gen.scope]
	scope := gen.scope.Children()[id]

	defer gen.setScope(gen.scope)
	gen.setScope(scope)

	var deferNodes []*ast.Defer

	if len(list.Nodes) > 0 {
		lastIdx := len(list.Nodes) - 1
		lastNode := list.Nodes[lastIdx]

		for _, stmt := range list.Nodes[:lastIdx] {
			if deferNode, _ := stmt.(*ast.Defer); deferNode != nil {
				println("deferred node: " + deferNode.Repr())
				deferNodes = append(deferNodes, deferNode)
				continue
			}

			gen.stmt(stmt)
		}

		if result != nil {
			gen.assign(gen.name(result), lastNode)
		} else if deferNode, _ := lastNode.(*ast.Defer); deferNode != nil {
			println("deferred node: " + deferNode.Repr())
			deferNodes = append(deferNodes, deferNode)
		} else {
			gen.stmt(lastNode)
		}
	}

	for i := len(deferNodes) - 1; i >= 0; i-- {
		gen.linef("L%d:;\n", gen.funcLabelID)
		gen.funcLabelID++
		gen.stmt(deferNodes[i].X)
	}

	delete(scopeIDs, gen.scope)
	gen.indent--
	gen.line("}\n")
}

func scopePath(scope *checker.Scope) string {
	buf := ""
	buf = scope.Name()

	for s := scope.Parent(); s != nil; s = s.Parent() {
		buf = s.Name() + " -> " + buf
	}

	return buf
}

var scopeIDs = map[*checker.Scope]int{}
