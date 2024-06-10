package ast

import "github.com/saffage/jet/token"

type Decl struct {
	Attrs *AttributeList
	Name  *Ident
	Mut   token.Loc // optional
	Type  Node      // can be nil if [Value] is specified
	Value Node      // can be nil if [Type] is specified
	IsVar bool      // indicates whether `=` is used before the value instead of `:`
}

func (decl *Decl) Pos() token.Loc {
	if decl.Mut.IsValid() {
		return decl.Mut
	}
	return decl.Name.Pos()
}

func (decl *Decl) LocEnd() token.Loc {
	if decl.Value != nil {
		return decl.Value.LocEnd()
	}
	if decl.Type != nil {
		return decl.Type.LocEnd()
	}
	return decl.Name.LocEnd()
}
