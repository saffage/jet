package types

import (
	"strconv"

	"github.com/saffage/jet/ast"
)

func (check *checker) typeSymOf(expr ast.Node) (t *TypeDef, err error) {
	switch node := expr.(type) {
	case *ast.Upper:
		if typedef := check.typeSymbolOf(node); typedef != nil {
			if typedef.Type() != nil {
				check.newUse(node, typedef)
				t = typedef
			} else {
				err = internalErrorf(node, "expression has no type")
			}
		}

	case *ast.TypeVar:
		panic("unimplemented")

	default:
		panic(internalErrorf(expr, "ill-formed AST: expected type, got %T instead", expr))
	}

	return
}

func (check *checker) resolveTypeDecl(decl *ast.TypeDecl) {
	if decl.Args != nil {
		check.error(&errorUnimplementedFeature{
			node:    decl.Args,
			feature: "type parameters",
		})
	}

	t := NewCustom(decl.Name.String(), nil, nil)
	env := NewNamedEnv(check.env, "type "+decl.Name.String())
	def := NewTypeDef(check.env, env, t, decl)

	switch expr := decl.Expr.(type) {
	case *ast.Extern:
		check.resolveExternTypeDecl(expr, decl)

	case *ast.Upper, *ast.TypeVar, *ast.Signature:
		check.error(&errorUnimplementedFeature{
			rng:     decl.EqTok.WithEnd(expr.PosEnd()),
			feature: "type aliases",
		})

	case *ast.Block:
		check.resolveTypeDeclBody(def, expr)

	default:
		panic(&errorIllFormedAst{decl})
	}
}

func (check *checker) resolveTypeAlias(decl *ast.TypeDecl, t Type) {
	typedesc, _ := As[*TypeDesc](t)

	if typedesc == nil {
		check.internalErrorf(decl.Expr, "expression is not a type")
		return
	}

	sym := NewTypeDef(check.env, nil, typedesc, decl)

	if defined := check.env.Define(sym); defined != nil {
		check.error(&errorAlreadyDefined{
			name: sym.Ident(),
			prev: defined.Ident(),
		})
		return
	}

	check.newDef(decl.Name, sym)
	check.setType(decl, typedesc)
}

func (check *checker) resolveTypeDeclBody(def *TypeDef, body *ast.Block) {
	fields := make([]Field, 0, len(body.Stmts.Nodes))
	variants := make([]Variant, 0, len(body.Stmts.Nodes))

	for _, node := range body.Stmts.Nodes {
		sym := Symbol(nil)

		switch node := node.(type) {
		case *ast.Label:
			decl, isDecl := node.X.(*ast.Decl)

			if !isDecl {
				panic(&errorIllFormedAst{node})
			}

			field := check.resolveTypeField(def, decl, node.Label())
			fields = append(fields, Field{
				Name: field.Name(),
				T:    field.Type(),
			})
			sym = field

		case *ast.Decl:
			field := check.resolveTypeField(def, node, nil)
			fields = append(fields, Field{
				Name: field.Name(),
				T:    field.Type(),
			})
			sym = field

		case *ast.Variant:
			variant := check.resolveTypeVariant(def, node)

			if variant == nil {
				continue
			}

			variants = append(variants, Variant{
				Name:   variant.Name(),
				Params: variant.ParamTypes(),
			})
			sym = variant

		default:
			panic(&errorIllFormedAst{node})
		}

		if defined := def.local.Define(sym); defined != nil {
			check.error(&errorAlreadyDefined{
				name: sym.Ident(),
				prev: defined.Ident(),
			})
		}
	}

	def.typedesc.base = NewCustom(def.Name(), fields, variants)
}

func (check *checker) resolveTypeField(
	def *TypeDef,
	node *ast.Decl,
	label *ast.Lower,
) *Binding {
	t, err := check.typeOf(node.Type)
	check.error(err)

	return NewField(def.local, def, t, node, label)
}

func (check *checker) resolveTypeVariant(
	def *TypeDef,
	node *ast.Variant,
) *Binding {
	if node.Name.String() == def.Name() {
		check.error(&errorAlreadyDefined{
			name: node.Name,
			prev: def.Ident(),
			hint: "The variant type cannot be named the same as the type in which it's defined",
		})
		return nil
	}

	if node.Params == nil {
		return NewVariant(def.local, nil, nil, def, node)
	}

	err := error(nil)
	env := NewNamedEnv(def.local, "type "+def.Name()+"."+node.Name.String())
	params := make([]*Binding, len(node.Params.Nodes))

	for i, param := range node.Params.Nodes {
		paramName := (*ast.Lower)(nil)
		decl := &ast.Decl{Type: param}

		if label, _ := param.(*ast.Label); label != nil {
			paramName = label.Label()

			if paramName == nil {
				panic(&errorIllFormedAst{param})
			}

			decl.Name = paramName
			param = label.X
		} else {
			if err == nil && i > 0 && params[i-1].label != nil {
				err = &errorPositionalParamAfterNamed{
					node:  param,
					named: params[i-1].Node(),
				}
			}

			decl.Name = &ast.Underscore{Data: "_" + strconv.Itoa(i)}
		}

		t, err := check.typeOf(param)
		check.error(err)

		sym := NewBinding(env, nil, &Value{T: t}, decl, nil)
		sym.label = paramName
		sym.isParam = true

		if defined := env.Define(sym); defined != nil {
			check.error(&errorAlreadyDefined{
				name: node.Name,
				prev: defined.Ident(),
			})
		}

		params[i] = sym
	}

	check.error(err)
	return NewVariant(def.local, env, params, def, node)
}
