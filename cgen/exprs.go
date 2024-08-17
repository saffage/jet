package cgen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

func (gen *generator) exprString(expr ast.Node) string {
	if _, isDecl := expr.(*ast.LetDecl); isDecl {
		return "ERROR_CGEN__EXPR_IS_DECL"
	}

	report.Debugf("expr = %s", expr.Repr())

	switch node := expr.(type) {
	case *ast.Empty:
		return ""

	case *ast.Name:
		switch sym := gen.SymbolOf(node).(type) {
		case *checker.Var, *checker.Func:
			return gen.name(sym)

		case *checker.Const:
			return gen.constant(sym.Value())

		case nil:
			panic(fmt.Sprintf("expression `%s` have no uses", expr))

		default:
			panic(fmt.Sprintf("invalid symbol '%T' for an expression: '%s'", sym, sym.Node().Repr()))
		}

	case *ast.Literal:
		typedValue, ok := gen.Types[expr]
		if !ok {
			report.Warningf("literal without type '%[1]T': %[1]s", expr)
			return "ERROR_CGEN__EXPR"
		}

		if typedValue.Value != nil {
			return gen.constant(typedValue.Value)
		}

	case *ast.Dot:
		tv := gen.Types[node.X]
		if tv == nil {
			// Defined in another module?
			panic("unreachable")
		}

		if types.IsTypeDesc(tv.Type) {
			ty := types.SkipTypeDesc(tv.Type)

			if _enum := types.AsEnum(ty); _enum != nil {
				// if tyY := gen.TypeOf(node.Y); tyY != nil && tyY.Equals(ty) {
				// 	// Enum field.
				// }
				return gen.TypeString(_enum) + "__" + node.Y.Data
				// return "ERROR_CGEN__INVALID_ENUM_FIELD"
			} else {
				return "ERROR_CGEN__INVALID_MEMBER_ACCESS"
			}
		}

		return gen.exprString(node.X) + "." + node.Y.Data

	case *ast.Deref:
		return fmt.Sprintf("(*%s)", gen.exprString(node.X))

	// case *ast.PrefixOp:
	// 	typedValue := gen.Types[node]

	// 	if typedValue == nil {
	// 		typedValue = gen.Types[expr]
	// 	}

	// 	if typedValue == nil {
	// 		panic("cannot get a type of the expression")
	// 	}

	// 	return gen.unary(node.X, typedValue.Type, node.Opr.Kind)

	case *ast.Op:
		tv := gen.Types[expr]
		if tv == nil {
			panic("cannot get a type of the expression")
		}

		if node.Y == nil {
			if node.X == nil {
				panic("unreachable")
			}

			return gen.unary(node.Y, tv.Type, node.Kind)
		}

		if node.X == nil {
			return gen.unary(node.Y, tv.Type, node.Kind)
		}

		return gen.binary(node.X, node.Y, tv.Type, node.Kind)

	case *ast.Call:
		if builtIn, _ := node.X.(*ast.BuiltIn); builtIn != nil {
			return gen.BuiltInCall(builtIn, node)
		}

		tv := gen.Types[node.X]
		if tv == nil {
			// Defined in another module?
			panic("unreachable")
		}

		if types.IsTypeDesc(tv.Type) {
			if ty := types.AsStruct(types.SkipTypeDesc(tv.Type)); ty != nil {
				tmp := gen.tempVar(ty)
				gen.structInit(gen.name(tmp), node, ty)
				return gen.name(tmp)
			} else {
				// Error in the checker
				return "ERROR_CGEN__INVALID_CALL"
			}
		} else if fn := types.AsFunc(tv.Type); fn != nil {
			buf := strings.Builder{}
			buf.WriteString(gen.exprString(node.X))
			buf.WriteByte('(')
			for i, arg := range node.Args.Nodes {
				if i != 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(gen.exprString(arg))
			}
			if fn.Result().Len() == 1 && types.IsArray(fn.Result().Types()[0]) {
				if len(node.Args.Nodes) > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString("/*RESULT*/")
			}
			buf.WriteByte(')')
			return buf.String()
		}

	case *ast.Index:
		buf := strings.Builder{}
		buf.WriteString(gen.exprString(node.X))
		buf.WriteByte('[')

		if len(node.Args.Nodes) != 1 {
			// Error in the checker
			panic("invalid arguments count for the index expression")
		}

		buf.WriteString(gen.exprString(node.Args.Nodes[0]))
		buf.WriteByte(']')
		return "(" + buf.String() + ")"

	case *ast.BracketList:
		// NOTE when array is used not in assignment they
		// must be prefixes with the type.
		tv := gen.Types[expr]
		if tv == nil || !types.IsArray(tv.Type) {
			panic("unreachable")
		}
		ty := types.AsArray(types.SkipUntyped(tv.Type))
		tmpVar := gen.tempVar(ty)
		gen.arrayInit(gen.name(tmpVar), node, ty)
		if tmpVar == nil {
			return ""
		}
		return gen.name(tmpVar)

	case *ast.If:
		ty := gen.TypeOf(expr)
		if ty == nil {
			panic("if expression have no type")
		}

		tmpVar := gen.tempVar(types.SkipUntyped(ty))
		gen.ifExpr(node, tmpVar)
		if tmpVar == nil {
			return ""
		}
		return gen.name(tmpVar)

	case *ast.CurlyList:
		ty := gen.TypeOf(expr)
		if ty == nil {
			panic("if expression have no type")
		}
		tmpVar := gen.tempVar(types.SkipUntyped(ty))
		gen.block(node.StmtList, tmpVar)
		if tmpVar == nil {
			return ""
		}
		return gen.name(tmpVar)

	case *ast.Defer:
		return ""

	default:
		report.TaggedErrorf(
			"internal: cgen",
			"expression '%T' is not implemented",
			node,
		)
	}

	report.TaggedWarningf("internal: cgen", "empty expression was generated: '%T'", expr)
	return "ERROR_CGEN__EXPR"
}

func (gen *generator) unary(x ast.Node, _ types.Type, op ast.OperatorKind) string {
	switch op {
	case ast.OperatorAddrOf, ast.OperatorMutAddrOf:
		return fmt.Sprintf("(&%s)", gen.exprString(x))

	case ast.OperatorNot:
		return fmt.Sprintf("(!%s)", gen.exprString(x))

	case ast.OperatorNeg:
		return fmt.Sprintf("(-%s)", gen.exprString(x))

	default:
		panic(fmt.Sprintf("not a unary operator: '%s'", op))
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
			gen.exprString(x),
			gen.exprString(y),
			gen.TypeString(t),
			op,
		)

	case ast.OperatorAdd,
		ast.OperatorSub,
		ast.OperatorMul,
		ast.OperatorDiv,
		ast.OperatorMod:
		return fmt.Sprintf("((%[3]s)(%[1]s) %[4]s (%[3]s)(%[2]s))",
			gen.exprString(x),
			gen.exprString(y),
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
			gen.exprString(x),
			gen.exprString(y),
			op,
		)

	case ast.OperatorAnd:
		return fmt.Sprintf("((%[3]s)(%[1]s) && (%[3]s)(%[2]s))",
			gen.exprString(x),
			gen.exprString(y),
			gen.TypeString(t),
		)

	case ast.OperatorOr:
		return fmt.Sprintf("((%[3]s)(%[1]s) || (%[3]s)(%[2]s))",
			gen.exprString(x),
			gen.exprString(y),
			gen.TypeString(t),
		)

	case ast.OperatorAs:
		return fmt.Sprintf("((%s)%s)",
			gen.TypeString(t),
			gen.exprString(x),
		)

	case ast.OperatorAssign:
		ty := gen.Module.TypeOf(y)

		if array := types.AsArray(ty); array != nil {
			gen.arrayAssign(gen.exprString(x), y, array)
			return ""
		}

		return fmt.Sprintf("%s = %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorAddAssign:
		return fmt.Sprintf("%s += %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorSubAssign:
		return fmt.Sprintf("%s -= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorMultAssign:
		return fmt.Sprintf("%s *= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorDivAssign:
		return fmt.Sprintf("%s /= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorModAssign:
		return fmt.Sprintf("%s %%= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorBitAndAssign:
		return fmt.Sprintf("%s &= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorBitOrAssign:
		return fmt.Sprintf("%s |= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorBitXorAssign:
		return fmt.Sprintf("%s ^= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorBitShlAssign:
		return fmt.Sprintf("%s <<= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	case ast.OperatorBitShrAssign:
		return fmt.Sprintf("%s >>= %s",
			gen.exprString(x),
			gen.exprString(y),
		)

	default:
		panic(fmt.Sprintf("not a binary operator: '%s'", op))
	}
}

func (gen *generator) assign(dest string, value ast.Node) {
	tv := gen.Types[value]
	if tv == nil {
		panic("cannot get a type of node")
	}
	switch ty := tv.Type.(type) {
	case *types.Array:
		gen.arrayAssign(dest, value, ty)

	case *types.Struct:
		gen.structAssign(dest, value, ty)

	default:
		if ty.Equals(types.Unit) {
			gen.linef("(void)%s;\n", gen.exprString(value))
		} else {
			gen.linef("%s = %s;\n", dest, gen.exprString(value))
		}
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

	case constant.Float:
		return (*constant.AsFloat(value)).String()

	case constant.String:
		value := constant.AsString(value)
		return strconv.Quote(*value)

	default:
		panic("unreachable")
	}
}

func (gen *generator) ifExpr(node *ast.If, result *checker.Var) {
	gen.linef("if (%s)\n", gen.exprString(node.Cond))
	gen.block(node.Body.StmtList, result)

	if node.Else != nil {
		gen.elseExpr(node.Else, result)
	}
}

func (gen *generator) elseExpr(node *ast.Else, result *checker.Var) {
	switch body := node.Body.(type) {
	case *ast.If:
		gen.linef("else if (%s)\n", gen.exprString(body.Cond))
		gen.block(body.Body.StmtList, result)

		if body.Else != nil {
			gen.elseExpr(body.Else, result)
		}

	case *ast.CurlyList:
		gen.line("else\n")
		gen.block(body.StmtList, result)

	default:
		panic("unreachable")
	}
}
