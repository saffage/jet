package cgen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

func (gen *generator) ExprString(expr ast.Node) string {
	if _, isDecl := expr.(ast.Decl); isDecl {
		return "ERROR_CGEN__EXPR_IS_DECL"
	}

	report.Debugf("expr = %s", expr)
	exprStr := ""

	switch node := expr.(type) {
	case *ast.Empty:
		return ""

	case *ast.BuiltInCall:
		exprStr = gen.BuiltInCall(node)

	case *ast.Ident:
		switch sym := gen.SymbolOf(node).(type) {
		case *checker.Var, *checker.Func:
			return gen.name(sym)

		case *checker.Const:
			return gen.constant(sym.Value())

		case nil:
			report.TaggedErrorf("cgen", "expression `%s` have no uses", expr)

		default:
			panic("idk")
		}

	case *ast.Literal:
		typedValue, ok := gen.Types[expr]
		if !ok {
			fmt.Printf("literal without type '%[1]T': %[1]s\n", expr)
			return "ERROR_CGEN__EXPR"
		}

		if typedValue.Value != nil {
			return gen.constant(typedValue.Value)
		}

	case *ast.MemberAccess:
		tv := gen.Types[node.X]
		if tv == nil {
			// Defined in another module?
			panic("idk")
		}

		if types.IsTypeDesc(tv.Type) {
			buf := strings.Builder{}
			t := types.SkipTypeDesc(tv.Type)

			if _struct := types.AsStruct(t); _struct != nil {
				buf.WriteString(fmt.Sprintf("(%s){\n", gen.TypeString(_struct)))
				gen.numIndent++
				buf.WriteString(gen.structInitFields(_struct, node.Selector))
				gen.numIndent--
				gen.indent(&buf)
				buf.WriteString("}")
				return buf.String()
			} else if _enum := types.AsEnum(t); _enum != nil {
				return gen.TypeString(_enum) + "__" + node.Selector.String()
			} else {
				return "ERROR_CGEN__INVALID_MEMBER_ACCESS"
			}
		} else {
			switch y := node.Selector.(type) {
			case *ast.Ident:
				return gen.ExprString(node.X) + "." + y.Name

			default:
				panic("not implemented")
			}
		}

	case *ast.PrefixOp:
		typedValue := gen.Types[node]

		if typedValue == nil {
			typedValue = gen.Types[expr]
		}

		if typedValue == nil {
			panic("cannot get a type of the expr node")
		}

		return gen.unary(node.X, typedValue.Type, node.Opr.Kind)

	case *ast.InfixOp:
		t := gen.TypeOf(node)
		if t == nil {
			panic("cannot get a type of the expr node")
		}

		return gen.binary(node.X, node.Y, t, node.Opr.Kind)

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
		return buf.String()

	case *ast.Index:
		buf := strings.Builder{}
		buf.WriteString(gen.ExprString(node.X))
		buf.WriteByte('[')

		if len(node.Args.Exprs) != 1 {
			panic("idk how to handle it")
		}

		buf.WriteString(gen.ExprString(node.Args.Exprs[0]))
		buf.WriteByte(']')
		return "(" + buf.String() + ")"

	case *ast.BracketList:
		if tv := gen.Types[expr]; tv != nil {
			buf := strings.Builder{}
			buf.WriteString(fmt.Sprintf("(%s){", gen.TypeString(tv.Type)))
			gen.numIndent++

			for i, elem := range node.Exprs {
				if i != 0 {
					buf.WriteString(", ")
				}

				buf.WriteString(gen.ExprString(elem))
			}

			gen.numIndent--
			buf.WriteString("}")
			return buf.String()
		}

	default:
		fmt.Printf("not implemented '%T'\n", node)
	}

	if exprStr == "" {
		fmt.Printf("empty expr at node '%T'\n", expr)
		return "ERROR_CGEN__EXPR"
	}

	// typeStr := gen.TypeString(typedValue.Type)
	// return fmt.Sprintf("((%s)%s)", typeStr, exprStr)
	return exprStr
}

func (gen *generator) structInitFields(t *types.Struct, selector ast.Node) string {
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

func (gen *generator) unary(x ast.Node, _ types.Type, op ast.OperatorKind) string {
	switch op {
	case ast.OperatorAddrOf:
		return fmt.Sprintf("(&%s)", gen.ExprString(x))

	case ast.OperatorStar:
		return fmt.Sprintf("(*%s)", gen.ExprString(x))

	case ast.OperatorNot:
		return fmt.Sprintf("(!%s)", gen.ExprString(x))

	default:
		panic(fmt.Sprintf("not a binary operator: '%s'", op))
	}
}

func (gen *generator) binary(x, y ast.Node, t types.Type, op ast.OperatorKind) string {
	t = types.SkipUntyped(t)

	switch op {
	case ast.OperatorBitAnd,
		ast.OperatorBitOr,
		ast.OperatorBitXor,
		ast.OperatorBitShl,
		ast.OperatorBitShr:
		return fmt.Sprintf("(%[3]s)((%[1]s) %[4]s (%[2]s))",
			gen.ExprString(x),
			gen.ExprString(y),
			gen.TypeString(t),
			op,
		)

	case ast.OperatorAdd,
		ast.OperatorSub,
		ast.OperatorMul,
		ast.OperatorDiv,
		ast.OperatorMod:
		return fmt.Sprintf("((%[3]s)(%[1]s) %[4]s (%[3]s)(%[2]s))",
			gen.ExprString(x),
			gen.ExprString(y),
			gen.TypeString(t),
			op,
		)

	case ast.OperatorEq,
		ast.OperatorNe,
		ast.OperatorGt,
		ast.OperatorGe,
		ast.OperatorLt,
		ast.OperatorLe:
		return fmt.Sprintf("((%[1]s) %[3]s (%[2]s))",
			gen.ExprString(x),
			gen.ExprString(y),
			op,
		)

	case ast.OperatorAnd:
		return fmt.Sprintf("((%[3]s)(%[1]s) && (%[3]s)(%[2]s))",
			gen.ExprString(x),
			gen.ExprString(y),
			gen.TypeString(t),
		)

	case ast.OperatorOr:
		return fmt.Sprintf("((%[3]s)(%[1]s) || (%[3]s)(%[2]s))",
			gen.ExprString(x),
			gen.ExprString(y),
			gen.TypeString(t),
		)

	case ast.OperatorAssign:
		t := gen.TypeOf(y)
		fmt.Printf("%s\n", t)
		if types.IsArray(t) {
			return fmt.Sprintf("memcpy(&%[1]s, %[2]s, sizeof(%[1]s))",
				gen.ExprString(x),
				gen.ExprString(y),
			)
		}

		return fmt.Sprintf("%s = %s",
			gen.ExprString(x),
			gen.ExprString(y),
		)

	default:
		panic(fmt.Sprintf("not a binary operator: '%s'", op))
	}
}

func (gen *generator) constant(value constant.Value) string {
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

	case constant.String:
		value := constant.AsString(value)
		return strconv.Quote(*value)

	default:
		panic("unreachable")
	}

	return "ERROR_CGEN__CONSTANT"
}
