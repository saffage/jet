package ast

import "github.com/saffage/jet/token"

type Decl interface {
	Node
	Ident() *Ident
	Doc() string
	Attributes() *AttributeList
}

type (
	ModuleDecl struct {
		Attrs        *AttributeList
		CommentGroup *CommentGroup
		Name         *Ident
		Body         Node      // Can be either [CurlyList] or [List].
		Loc          token.Loc // `module` token.
	}

	VarDecl struct {
		Attrs        *AttributeList
		CommentGroup *CommentGroup
		Binding      *Binding
		Value        Node
		Loc          token.Loc
	}

	FuncDecl struct {
		Attrs        *AttributeList
		CommentGroup *CommentGroup
		Name         *Ident
		Signature    *Signature
		Body         *CurlyList
		Loc          token.Loc // `func` token.
	}

	TypeAliasDecl struct {
		Attrs        *AttributeList
		CommentGroup *CommentGroup
		Name         *Ident
		Expr         Node
		Loc          token.Loc // `alias` token.
	}
)

func (n *ModuleDecl) Pos() token.Loc             { return n.Loc }
func (n *ModuleDecl) LocEnd() token.Loc          { return n.Name.LocEnd() }
func (n *ModuleDecl) Ident() *Ident              { return n.Name }
func (n *ModuleDecl) Doc() string                { return n.CommentGroup.Merged() }
func (n *ModuleDecl) Attributes() *AttributeList { return n.Attrs }

func (n *VarDecl) Pos() token.Loc { return n.Loc }
func (n *VarDecl) LocEnd() token.Loc {
	if n.Value != nil {
		return n.Value.LocEnd()
	}
	return n.Binding.Type.LocEnd()
}
func (n *VarDecl) Ident() *Ident              { return n.Binding.Name }
func (n *VarDecl) Doc() string                { return n.CommentGroup.Merged() }
func (n *VarDecl) Attributes() *AttributeList { return n.Attrs }

func (n *FuncDecl) Pos() token.Loc { return n.Loc }
func (n *FuncDecl) LocEnd() token.Loc {
	if n.Body != nil {
		return n.Body.LocEnd()
	}
	return n.Signature.LocEnd()
}
func (n *FuncDecl) Ident() *Ident              { return n.Name }
func (n *FuncDecl) Doc() string                { return n.CommentGroup.Merged() }
func (n *FuncDecl) Attributes() *AttributeList { return n.Attrs }

func (n *TypeAliasDecl) Pos() token.Loc             { return n.Loc }
func (n *TypeAliasDecl) LocEnd() token.Loc          { return n.Expr.LocEnd() }
func (n *TypeAliasDecl) Ident() *Ident              { return n.Name }
func (n *TypeAliasDecl) Doc() string                { return n.CommentGroup.Merged() }
func (n *TypeAliasDecl) Attributes() *AttributeList { return n.Attrs }
