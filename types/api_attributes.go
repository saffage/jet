package types

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
)

func HasAttribute(sym Symbol, name string) bool {
	return GetAttribute(sym, name) != nil
}

func GetAttribute(sym Symbol, name string) ast.Node {
	if decl, _ := sym.Node().(*ast.LetDecl); decl != nil {
		return FindAttr(decl.Attrs, name)
	}
	return nil
}

func FindAttr(attrList *ast.AttributeList, attr string) ast.Node {
	if attrList != nil {
		for _, expr := range attrList.List.Nodes {
			if ident, _ := expr.(*ast.Lower); ident != nil && ident.Data == attr {
				return expr
			} else if call, _ := expr.(*ast.Call); call != nil {
				if ident, _ := call.X.(*ast.Lower); ident != nil && ident.Data == attr {
					return expr
				}
			}
		}
	}
	return nil
}

func (check *checker) resolveFuncAttrs(sym *Binding) {
	if sym.node.Attrs == nil {
		return
	}

	for _, attr := range sym.node.Attrs.List.Nodes {
		switch attr := attr.(type) {
		case *ast.Call:
			attrIdent, _ := attr.X.(*ast.Lower)

			if attrIdent == nil {
				check.errorf(attr.X, "expected identifier")
				continue
			}

			switch attrIdent.Data {
			case "extern_c":
				check.attrExternC(sym, attr)

			case "header":
				// Unchecked attribute

			default:
				check.errorf(attrIdent, "unknown attribute")
			}

		case *ast.Lower:
			switch attr.Data {
			case "extern_c":
				check.attrExternC(sym, attr)

			default:
				check.errorf(attr, "unknown attribute")
			}

		default:
			panic(errorf(
				attr,
				"unexpected node type %T is attribute list",
				attr,
			))
		}
	}

	if sym.isExtern {
		if sym.node.Value != nil {
			check.errorf(
				sym.node.Decl.Name,
				"functions with @[extern_c] attribute must have no definition",
			)
		}
	} else {
		if sym.node.Value == nil {
			check.errorf(
				sym.node.Decl.Name,
				"functions without body is not allowed",
			)
		}
		if sym.Variadic() != nil {
			check.errorf(
				sym.node.Decl.Name,
				"only a function with the attribute @[extern_c] can be variadic",
			)
		}
	}
}

func (check *checker) attrExternC(sym Symbol, node ast.Node) {
	externName := ""

	switch node := node.(type) {
	case *ast.Call:
		tyExternCAttr := NewFunction(Params{UntypedStringType}, NoneType, nil)

		tyArgs := make(Params, 0, len(node.Args.Nodes))
		args := make([]*Value, 0, len(node.Args.Nodes))

		for _, arg := range node.Args.Nodes {
			value, err := check.valueOf(arg)

			if err != nil || value == nil {
				check.problem(err)
				continue
			}

			tyArgs = append(tyArgs, value.T)
			args = append(args, value)
		}

		if idx, err := tyExternCAttr.CheckArgs(tyArgs); err != nil {
			n := ast.Node(node.Args)
			if idx < len(node.Args.Nodes) {
				n = node.Args.Nodes[idx]
			}
			check.errorf(n, "%s", err.Error())
			return
		}

		externName = *constant.AsString(args[0].V)
	}

	switch sym := sym.(type) {
	case *Binding:
		sym.externName = externName
		sym.isExtern = true

	default:
		check.errorf(
			sym.Ident(),
			"expected function for @[extern_c] attribute",
		)
	}
}
