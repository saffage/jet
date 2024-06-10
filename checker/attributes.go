package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/types"
)

func HasAttribute(sym Symbol, name string) bool {
	return GetAttribute(sym, name) != nil
}

func GetAttribute(sym Symbol, name string) ast.Node {
	if decl, _ := sym.Node().(*ast.Decl); decl != nil {
		return FindAttr(decl.Attrs, name)
	}
	return nil
}

func FindAttr(attrList *ast.AttributeList, attr string) ast.Node {
	if attrList != nil {
		for _, expr := range attrList.List.Nodes {
			if ident, _ := expr.(*ast.Ident); ident != nil && ident.Name == attr {
				return expr
			} else if call, _ := expr.(*ast.Call); call != nil {
				if ident, _ := call.X.(*ast.Ident); ident != nil && ident.Name == attr {
					return expr
				}
			}
		}
	}
	return nil
}

func (check *Checker) resolveFuncAttrs(sym *Func) {
	if sym.decl.Attrs == nil {
		return
	}

	for _, attr := range sym.decl.Attrs.List.Nodes {
		switch attr := attr.(type) {
		case *ast.Call:
			attrIdent, _ := attr.X.(*ast.Ident)

			if attrIdent == nil {
				check.errorf(attr.X, "expected identifier")
				continue
			}

			switch attrIdent.Name {
			case "extern_c":
				check.attrExternC(sym, attr)

			default:
				check.errorf(attrIdent, "unknown attribute")
			}

		case *ast.Ident:
			switch attr.Name {
			case "extern_c":
				check.attrExternC(sym, attr)

			default:
				check.errorf(attr, "unknown attribute")
			}

		default:
			check.errorf(
				attr,
				"Ill-formed AST: unexpected node type %T is attribute list",
				attr,
			)
			continue
		}
	}

	if sym.isExtern {
		if sym.body != nil {
			check.errorf(
				sym.decl.Ident,
				"functions with @[extern_c] attribute must have no definition",
			)
		}
	} else {
		if sym.body == nil {
			check.errorf(
				sym.decl.Ident,
				"functions without body is not allowed",
			)
		}
		if sym.ty.Variadic() != nil {
			check.errorf(
				sym.decl.Ident,
				"only a function with the attribute @[extern_c] can be variadic",
			)
		}
	}
}

func (check *Checker) attrExternC(sym Symbol, node ast.Node) {
	externName := ""

	switch node := node.(type) {
	case *ast.Call:
		tyExternCAttr := types.NewFunc(
			types.NewTuple(types.UntypedString),
			types.Unit,
			nil,
		)

		tyArgs := make([]types.Type, 0, len(node.Args.Nodes))
		args := make([]*TypedValue, 0, len(node.Args.Nodes))

		for _, arg := range node.Args.Nodes {
			if tv := check.valueOf(arg); tv != nil {
				tyArgs = append(tyArgs, tv.Type)
				args = append(args, tv)
			} else {
				continue
			}
		}

		tyArgsTuple := types.NewTuple(tyArgs...)
		if idx, err := tyExternCAttr.CheckArgs(tyArgsTuple); err != nil {
			n := ast.Node(node.Args)
			if idx < len(node.Args.Nodes) {
				n = node.Args.Nodes[idx]
			}
			check.errorf(n, err.Error())
			return
		}

		externName = *constant.AsString(args[0].Value)
	}

	switch sym := sym.(type) {
	case *Func:
		sym.externName = externName
		sym.isExtern = true

	default:
		check.errorf(
			sym.Ident(),
			"expected function for @[extern_c] attribute",
		)
	}
}
