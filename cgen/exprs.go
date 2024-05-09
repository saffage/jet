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
	if _, isDecl := expr.(ast.Decl); isDecl {
		return "ERROR_CGEN__EXPR_IS_DECL"
	}

	typedValue, ok := gen.Types[expr]
	if !ok {
		fmt.Printf("expr without type '%T'\n", expr)
		return "ERROR_CGEN__EXPR"
	}

	exprStr := ""

	if typedValue.Value != nil {
		return gen.constant(typedValue.Value)
	}

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
				return sym.Name()
			}

		case *checker.Func:
			return sym.Name()

		default:
			fmt.Printf("%#v\n", sym)
			panic("idk")
		}

	case *ast.MemberAccess:
		tv := gen.Types[node.X]

		if !types.IsStruct(types.SkipTypeDesc(tv.Type)) {
			return "ERROR_CGEN__INVALID_MEMBER_ACCESS"
		}

		if types.IsTypeDesc(tv.Type) {
			buf := strings.Builder{}
			buf.WriteString("{\n")

			gen.numIndent++
			buf.WriteString(gen.structInitFields(node.Selector))
			gen.numIndent--

			gen.indent(&buf)
			buf.WriteString("}")

			exprStr = buf.String()
		} else {
			switch y := node.Selector.(type) {
			case *ast.Ident:
				exprStr = gen.ExprString(node.X) + "." + y.Name

			default:
				panic("not implemented")
			}
		}

	case *ast.PrefixOp:
		switch node.Opr.Kind {
		case ast.OperatorNot:
			exprStr = "!" + gen.ExprString(node.X)

		default:
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

		case ast.OperatorEq:
			exprStr = fmt.Sprintf(
				"%s == %s",
				gen.ExprString(node.X),
				gen.ExprString(node.Y),
			)

		case ast.OperatorNe:
			exprStr = fmt.Sprintf(
				"%s != %s",
				gen.ExprString(node.X),
				gen.ExprString(node.Y),
			)

		case ast.OperatorAssign:
			

		default:
			panic(fmt.Sprintf("not implemented: infix operator '%s'", node.Opr.Kind))
		}

	case *ast.Call:
		buf := strings.Builder{}
		buf.WriteString(gen.ExprString(node.X))
		buf.WriteByte('(')
		for i, arg := range node.Args.Exprs {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(gen.ExprString(arg))
		}
		buf.WriteByte(')')
		return "(" + buf.String() + ")"

	default:
		fmt.Printf("not implemented '%T'\n", node)
	}

	if exprStr == "" {
		fmt.Printf("empty expr at node '%T'\n", expr)
		return "ERROR_CGEN__EXPR"
	}

	typeStr := gen.TypeString(typedValue.Type)
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
