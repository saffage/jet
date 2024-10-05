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
				err = errorf(node, "expression has no type")
			}
		}

	case *ast.TypeVar:
		panic("unimplemented")

	default:
		panic(errorf(expr, "ill-formed AST: expected type, got %T instead", expr))
	}

	return
}

func (check *checker) resolveTypeDecl(decl *ast.TypeDecl) {
	if decl.Args != nil {
		check.errorf(decl.Args, "type parameters are not implemented")
	}

	t := NewCustom(decl.Name.String(), nil, nil)
	env := NewNamedEnv(check.env, "type "+decl.Name.String())
	def := NewTypeDef(check.env, env, t, decl)

	switch expr := decl.Expr.(type) {
	case *ast.Extern:
		check.resolveExternTypeDecl(expr, decl)

	case *ast.Upper, *ast.TypeVar, *ast.Signature:
		panic("unimplemented")

	case *ast.Block:
		check.resolveTypeDeclBody(def, expr)

	default:
		panic(&errorIllFormedAst{decl})
	}
}

func (check *checker) resolveTypeAlias(decl *ast.TypeDecl, t Type) {
	typedesc := As[*TypeDesc](t)

	if typedesc == nil {
		check.errorf(decl.Expr, "expression is not a type")
		return
	}

	sym := NewTypeDef(check.env, nil, typedesc, decl)

	if defined := check.env.Define(sym); defined != nil {
		check.problem(&errorAlreadyDefined{sym.Ident(), defined.Ident()})
		return
	}

	check.newDef(decl.Name, sym)
	check.setType(decl, typedesc)
}

func (check *checker) resolveTypeDeclBody(def *TypeDef, body *ast.Block) {
	fields := make([]Field, len(body.Stmts.Nodes))
	variants := make([]Variant, len(body.Stmts.Nodes))

	for _, node := range body.Stmts.Nodes {
		publicName := (*ast.Lower)(nil)

		if label, _ := node.(*ast.Label); label != nil {
			publicName = label.Label()
			node = label.X
		}

		switch node := node.(type) {
		case *ast.Decl:
			t, err := check.typeOf(node.Type)
			check.problem(err)

			field := NewBinding(def.local, nil, &Value{T: t}, node, nil)
			field.isField = true
			field.label = publicName

			if defined := def.local.Define(field); defined != nil {
				check.problem(&errorAlreadyDefined{
					name: node.Name,
					prev: defined.Ident(),
				})
			}

		case *ast.Variant:
			env := (*Env)(nil)
			params := ([]*Binding)(nil)

			if node.Params != nil {
				err := error(nil)

				env = NewNamedEnv(def.local, "type "+def.Name()+"."+node.Name.String())
				params = make([]*Binding, len(node.Params.Nodes))

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
					check.problem(err)

					sym := NewBinding(env, nil, &Value{T: t}, decl, nil)
					sym.label = paramName
					sym.isParam = true

					if defined := env.Define(sym); defined != nil {
						check.problem(&errorAlreadyDefined{
							name: node.Name,
							prev: defined.Ident(),
						})
					}

					params[i] = sym
				}

				check.problem(err)
			}

			variant := NewVariant(def.local, env, params, def, node)

			if defined := def.local.Define(variant); defined != nil {
				check.problem(&errorAlreadyDefined{
					name: node.Name,
					prev: defined.Ident(),
				})
			}

		default:
			panic(&errorIllFormedAst{node})
		}
	}

	def.typedesc.base = NewCustom(def.Name(), fields, variants)
}
