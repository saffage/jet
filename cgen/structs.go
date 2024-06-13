package cgen

import (
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) structDecl(sym *checker.Struct) {
	buf := strings.Builder{}
	ty := types.AsStruct(types.SkipTypeDesc(sym.Type()))
	gen.flinef(&buf, "typedef struct %s {\n", sym.Name())
	gen.indent++
	for _, field := range ty.Fields() {
		gen.flinef(&buf, "%s %s;\n", gen.TypeString(field.Type), field.Name)
	}
	gen.indent--
	gen.flinef(&buf, "} %s;\n", gen.name(sym))
	gen.typeSect.WriteString(buf.String())
}

func (gen *generator) structInit(prefix string, value ast.Node, ty *types.Struct) {
	if call, _ := value.(*ast.Call); call != nil {
		gen.structInitList(prefix, call.Args.List, ty)
	} else {
		gen.linef("%s = %s;\n", prefix, gen.exprString(value))
	}

	// buf := strings.Builder{}

	// for _, node := range initList.Nodes {
	// 	switch node := node.(type) {
	// 	case *ast.Op:
	// 		if node.X == nil || node.Y == nil {
	// 			panic("todo")
	// 		}

	// 		value := ""
	// 		if lit, _ := node.Y.(*ast.BracketList); lit != nil {
	// 			ty := gen.TypeOf(lit)
	// 			if ty == nil || !types.IsArray(ty) {
	// 				panic("unreachable")
	// 			}
	// 			gen.arrayInit(node.X.(*ast.Ident).Name, lit, types.AsArray(ty))
	// 			// value = gen.arrayLit(lit, types.AsArray(ty))
	// 		} else {
	// 			value = gen.exprString(node.Y)
	// 		}
	// 		buf.WriteString(fmt.Sprintf(".%s = %s,\n", node.X.(*ast.Ident).Name, value))

	// 	default:
	// 		panic("unreachable")
	// 	}
	// }
}

func (gen *generator) structInitList(prefix string, initList *ast.List, _ *types.Struct) {
	for _, expr := range initList.Nodes {
		switch expr := expr.(type) {
		case *ast.Op:
			if expr.X == nil || expr.Y == nil {
				panic("unreachable")
			}
			field, ok := expr.X.(*ast.Ident)
			if !ok {
				panic("unreachable")
			}
			gen.assign(prefix+"."+field.Name, expr.Y)
			// gen.linef("%s.%s = %s;\n", prefix, field.Name, value)

		case *ast.Ident, *ast.Dot:
			println(expr.Repr())
			panic("not implemented")

		default:
			panic("unreachable")
		}
	}
}

func (gen *generator) structAssign(dest string, src ast.Node, ty *types.Struct) {
	if call, _ := src.(*ast.Call); call != nil {
		tv := gen.Types[call.X]
		if tv == nil {
			panic("cannot get a type of node")
		}
		if types.IsTypeDesc(tv.Type) &&
			types.IsStruct(types.SkipTypeDesc(tv.Type)) {
			gen.structInit(dest, call, ty)
			return
		}
	}
	gen.linef("%s = %s;\n", dest, gen.exprString(src))
}
