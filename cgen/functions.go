package cgen

import (
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *Generator) Func(sym *checker.Func) {
	t := sym.Type().(*types.Func)
	declBuf := strings.Builder{}
	tResultVar := types.Type(nil)

	if sym.Name() == "main" {
		gen.codeSect.WriteString(fnMainHead)
	} else {
		result := t.Result()

		if result.Len() == 0 {
			declBuf.WriteString("void")
		} else if result.Len() == 1 {
			declBuf.WriteString(gen.TypeString(result.Underlying()))
			tResultVar = result.Underlying()
		} else {
			gen.errorf(sym, "tuple are not supported")
			declBuf.WriteString("ERROR_CGEN__FUNC_TUPLE_RESULT")
		}

		declBuf.WriteByte(' ')
		declBuf.WriteString(sym.Name())
		declBuf.WriteByte('(')

		params := []*checker.Var{}

		// Find params.
		for _, def := range gen.Defs {
			_var, _ := def.(*checker.Var)
			if _var != nil && _var.Owner() == sym.Local() && _var.IsParam() {
				params = append(params, _var)
			}
		}

		// Gen params.
		for i, param := range params {
			if i != 0 {
				declBuf.WriteString(", ")
			}

			declBuf.WriteString(gen.TypeString(param.Type()))
			declBuf.WriteByte(' ')
			declBuf.WriteString("p_" + param.Name())
		}

		declBuf.WriteByte(')')

		gen.declFnsSect.WriteString(declBuf.String())
		gen.declFnsSect.WriteString(";\n")
		gen.codeSect.WriteString(declBuf.String())
	}

	node := sym.Node().(*ast.FuncDecl)

	gen.codeSect.WriteString(" {\n")
	gen.numIndent++

	if tResultVar != nil {
		gen.indent(&gen.codeSect)
		gen.codeSect.WriteString(gen.TypeString(tResultVar))
		gen.codeSect.WriteString(" __result;\n\n")
	}

	for i, stmt := range node.Body.Nodes {
		gen.indent(&gen.codeSect)

		if decl, ok := stmt.(ast.Decl); ok && decl != nil {
			switch decl := decl.(type) {
			case *ast.VarDecl:
				if def := gen.Defs[decl.Binding.Name]; def != nil {
					gen.codeSect.WriteString(gen.TypeString(def.Type()))
					gen.codeSect.WriteString(" " + def.Name() + ";\n")

					if decl.Value != nil {
						gen.indent(&gen.codeSect)
						gen.codeSect.WriteString(def.Name() + " = ")
						gen.codeSect.WriteString(gen.ExprString(decl.Value))
						gen.codeSect.WriteString(";\n")
					}
				} else {
					panic("unreachable")
				}

			default:
				panic("not implemented")
			}
		} else {
			if tResultVar != nil && i == len(node.Body.Nodes)-1 {
				gen.codeSect.WriteString("__result = ")
			}

			gen.codeSect.WriteString(gen.ExprString(stmt))
			gen.codeSect.WriteString(";\n")
		}
	}

	// gen.indent(&gen.codeSect)
	// gen.codeSect.WriteString("goto L_ret;\n")
	// gen.codeSect.WriteString("\nL_ret:;\n")

	if tResultVar != nil {
		gen.indent(&gen.codeSect)
		gen.codeSect.WriteString("return __result;\n")
	}

	gen.numIndent--
	gen.codeSect.WriteString("}\n")
}
