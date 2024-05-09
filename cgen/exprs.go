package cgen

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/types"
)

func (gen *Generator) ExprString(expr ast.Node) string {
	typedValue, ok := gen.Types[expr]
	if !ok {
		fmt.Printf("expr without type '%T'", expr)
		return "ERROR_CGEN__EXPR"
	}

	typeStr := gen.TypeString(typedValue.Type)
	exprStr := ""

	if typedValue.Value != nil {
		return gen.constant(typedValue.Value)
	} else {
		switch node := expr.(type) {
		case *ast.BuiltInCall:
			exprStr = gen.BuiltInCall(node)

		case *ast.Ident:
		outer:
			switch sym := gen.Uses[node].(type) {
			case *checker.Var:
				switch {
				case sym.IsParam():
					exprStr = "p_" + sym.Name()
					break outer

				default:
					break
				}

			default:
				fmt.Printf("%#v\n", sym)
				panic("idk")
			}

		case *ast.MemberAccess:
			tv := gen.Types[node.X]

			if types.IsTypeDesc(tv.Type) && types.IsStruct(types.SkipTypeDesc(tv.Type)) {
				buf := strings.Builder{}
				buf.WriteString("{\n")

				gen.numIndent++
				buf.WriteString(gen.structInitFields(node.Selector))
				gen.numIndent--

				gen.indent(&buf)
				buf.WriteString("}")

				exprStr = buf.String()
			} else {
				panic("not implemented")
			}

		case *ast.InfixOp:
			switch node.Opr.Kind {
			case ast.OperatorLt:
				exprStr = fmt.Sprintf(
					"%s < %s",
					gen.ExprString(node.X),
					gen.ExprString(node.Y),
				)

			case ast.OperatorLe:
				exprStr = fmt.Sprintf(
					"%s <= %s",
					gen.ExprString(node.X),
					gen.ExprString(node.Y),
				)

			case ast.OperatorGt:
				exprStr = fmt.Sprintf(
					"%s > %s",
					gen.ExprString(node.X),
					gen.ExprString(node.Y),
				)

			case ast.OperatorGe:
				exprStr = fmt.Sprintf(
					"%s >= %s",
					gen.ExprString(node.X),
					gen.ExprString(node.Y),
				)
			}

		default:
			fmt.Printf("not implemented '%T'", node)
		}
	}

	if exprStr == "" {
		fmt.Printf("empty expr node '%T'", expr)
		return "ERROR_CGEN__EXPR"
	}

	return fmt.Sprintf("((%s)%s)", typeStr, exprStr)
}

func (gen *Generator) structInitFields(selector ast.Node) string {
	buf := strings.Builder{}

	if list, _ := selector.(*ast.CurlyList); list != nil {
		for _, node := range list.Nodes {
			switch node := node.(type) {
			case *ast.InfixOp:
				gen.indent(&buf)
				buf.WriteString(fmt.Sprintf(
					".%s = %s,\n",
					node.X.(*ast.Ident).Name,
					gen.ExprString(node.Y),
				))

			default:
				panic("unreachable")
			}
		}
	}

	return buf.String()
}

func (gen *Generator) constant(value constant.Value) string {
	if value == nil {
		panic("nil constant value")
	}

	switch value.Kind() {
	case constant.Bool:
		if *constant.AsBool(value) {
			return "true"
		}
		return "false"

	case constant.Int:
		return (*constant.AsInt(value)).String()

	default:
		panic("unreachable")
	}

	return "ERROR_CGEN__CONSTANT"
}
