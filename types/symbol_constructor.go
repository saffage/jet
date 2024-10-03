package types

import "github.com/saffage/jet/ast"

type Constructor struct {
	owner  *Env
	node   *ast.Variant
	tOwner *TypeDef
	value  *Value
}

func newConstructor(owner *Env, sym *TypeDef, value *Value, decl *ast.Variant) *Constructor {
	return &Constructor{owner, decl, sym, value}
}

func newConstructor2(owner *Env, sym *TypeDef, value *Value, decl *ast.TypeDecl) *Constructor {
	return &Constructor{owner, &ast.Variant{Name: decl.Name}, sym, value}
}

func (sym *Constructor) Type() Type   { return sym.value.T }
func (sym *Constructor) Name() string { return sym.node.Name.String() }
func (sym *Constructor) Owner() *Env  { return sym.owner }

func (sym *Constructor) Node() ast.Node {
	if sym.node == nil {
		return sym.tOwner.Node()
	}
	return sym.node
}

func (sym *Constructor) Ident() ast.Ident {
	if sym.node == nil {
		return sym.tOwner.Ident()
	}
	return sym.node.Name
}
