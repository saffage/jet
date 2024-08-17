package cgen

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

func (gen *generator) stmt(stmt ast.Node) {
	report.Debugf("stmt = %s", stmt.Repr())
	switch stmt := stmt.(type) {
	case *ast.Empty:
		gen.line("\n")

	case *ast.LetDecl:
		// if !stmt.IsVar {
		// 	break
		// }
		sym, _ := gen.Defs.Get(stmt.Decl.Name)
		if sym == nil {
			panic("unreachable")
		}
		gen.decl(sym)

	case *ast.While:
		gen.linef("while (%s)\n", gen.exprString(stmt.Cond))
		gen.block(stmt.Body.StmtList, nil)

	case *ast.For:
		loopVar := gen.SymbolOf(stmt.Decls.Nodes[0].(*ast.LetDecl).Decl.Name)
		iterExpr := stmt.IterExpr.(*ast.Op)
		cmpOp := ast.OperatorLt
		if iterExpr.Kind == ast.OperatorRangeInclusive {
			cmpOp = ast.OperatorLe
		}
		gen.linef(
			"for (%[1]s %[2]s=%[3]s; %[4]s; %[2]s+=1)\n",
			gen.TypeString(loopVar.Type()),
			gen.name(loopVar),
			gen.exprString(iterExpr.X),
			gen.binary(loopVar.Ident(), iterExpr.Y, types.Bool, cmpOp),
		)
		gen.block(stmt.Body.StmtList, nil)

	case *ast.If:
		gen.ifExpr(stmt, nil)
		// gen.linef("if (%s) {\n", gen.exprString(stmt.Cond))
		// gen.indent++
		// for _, stmt := range stmt.Body.Nodes {
		// 	gen.codeSect.WriteString(gen.StmtString(stmt))
		// }
		// gen.indent--
		// if stmt.Else != nil {
		// 	gen.linef("} else %s", gen.StmtString(stmt.Else.Body))
		// } else {
		// 	gen.line("}\n")
		// }

	case *ast.CurlyList:
		gen.block(stmt.StmtList, nil)

	case *ast.Break:
		gen.line("break;\n")

	case *ast.Continue:
		gen.line("continue;\n")

	default:
		expr := gen.exprString(stmt)
		if expr != "" {
			if ty := gen.TypeOf(stmt); ty != nil && !ty.Equals(types.Unit) {
				gen.linef("(void)%s;\n", expr)
			} else {
				gen.linef("%s;\n", expr)
			}
		}
	}
}
