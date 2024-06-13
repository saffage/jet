package ast

import (
	"strings"

	"github.com/saffage/jet/token"
)

type Node interface {
	// Start of the entire tree. This position must also include nested nodes.
	Pos() token.Pos

	// End of the entire tree. This position must also include nested nodes.
	PosEnd() token.Pos

	// String representation of the node. This string must be equal to the
	// code from which this tree was parsed (ignoring location).
	Repr() string

	implNode()
}

//------------------------------------------------
// Atoms
//------------------------------------------------

type (
	BadNode struct {
		DesiredPos token.Pos
	}

	Empty struct {
		DesiredPos token.Pos
	}

	Ident struct {
		Name       string
		Start, End token.Pos
	}

	Literal struct {
		Value      string
		Kind       LiteralKind
		Start, End token.Pos
	}
)

func (n *BadNode) Pos() token.Pos    { return n.DesiredPos }
func (n *BadNode) PosEnd() token.Pos { return n.DesiredPos }

func (n *Empty) Pos() token.Pos    { return n.DesiredPos }
func (n *Empty) PosEnd() token.Pos { return n.DesiredPos }

func (n *Ident) Pos() token.Pos    { return n.Start }
func (n *Ident) PosEnd() token.Pos { return n.End }

func (n *Literal) Pos() token.Pos    { return n.Start }
func (n *Literal) PosEnd() token.Pos { return n.End }

//------------------------------------------------
// Declaration
//------------------------------------------------

type (
	Comment struct {
		Value string
		Start token.Pos
		End   token.Pos
	}

	CommentGroup struct {
		Comments []*Comment
	}

	// Represents `@[...attributes]`.
	AttributeList struct {
		List   *BracketList
		TokLoc token.Pos // `@` token.
	}

	// Represents `@[..attributes] mut name: T = expr`.
	Decl struct {
		Attrs *AttributeList
		Ident *Ident
		Mut   token.Pos // optional
		Type  Node      // optional
		Value Node      // optional
		IsVar bool      // indicates whether `=` is used before the value instead of `:`
	}
)

func (n *Comment) Pos() token.Pos    { return n.Start }
func (n *Comment) PosEnd() token.Pos { return n.End }

func (n *CommentGroup) Pos() token.Pos    { return n.Comments[0].Pos() }
func (n *CommentGroup) PosEnd() token.Pos { return n.Comments[len(n.Comments)-1].PosEnd() }

func (n *AttributeList) Pos() token.Pos    { return n.TokLoc }
func (n *AttributeList) PosEnd() token.Pos { return n.List.PosEnd() }

func (decl *Decl) Pos() token.Pos {
	if decl.Mut.IsValid() {
		return decl.Mut
	}
	return decl.Ident.Pos()
}

func (decl *Decl) PosEnd() token.Pos {
	if decl.Value != nil {
		return decl.Value.PosEnd()
	}
	if decl.Type != nil {
		return decl.Type.PosEnd()
	}
	return decl.Ident.PosEnd()
}

func (n *CommentGroup) Merged() string {
	buf := strings.Builder{}

	for _, comment := range n.Comments {
		buf.WriteString(comment.Value[1:])
	}

	return buf.String()
}

//------------------------------------------------
// Composite nodes
//------------------------------------------------

type (
	// Represents `[...args]x`.
	ArrayType struct {
		X    Node
		Args *BracketList
	}

	// Represents `struct{...fields}`.
	StructType struct {
		Fields []*Decl
		TokPos token.Pos
		Open   token.Pos
		Close  token.Pos
	}

	// Represents `enum{...fields}`.
	EnumType struct {
		Fields []*Ident
		TokPos token.Pos
		Open   token.Pos
		Close  token.Pos
	}

	// Represents `() -> ()`.
	Signature struct {
		Params *ParenList
		Result Node // can be nil in some cases
	}

	// Represents an identifier, prefixed with a '$' sign.
	BuiltIn struct {
		*Ident
		TokPos token.Pos // '$' token.
	}

	// Represents `x(...args)`.
	Call struct {
		X    Node
		Args *ParenList
	}

	// Represents `x[...args]`.
	Index struct {
		X    Node
		Args *BracketList
	}

	// Represents `(...params) -> T {...}` or `() expr`
	Function struct {
		*Signature
		Body Node
	}

	// Represents `x.y`.
	Dot struct {
		X      Node
		Y      *Ident
		DotPos token.Pos
	}

	// Represents `x.*`.
	Deref struct {
		X       Node
		DotPos  token.Pos
		StarPos token.Pos
	}

	// Represents `x OP y`, where `OP` is an operator.
	Op struct {
		X     Node
		Y     Node
		Start token.Pos
		End   token.Pos
		Kind  OperatorKind
	}
)

func (n *ArrayType) Pos() token.Pos    { return n.Args.Pos() }
func (n *ArrayType) PosEnd() token.Pos { return n.X.PosEnd() }

func (n *StructType) Pos() token.Pos    { return n.TokPos }
func (n *StructType) PosEnd() token.Pos { return n.Close }

func (n *EnumType) Pos() token.Pos    { return n.TokPos }
func (n *EnumType) PosEnd() token.Pos { return n.Close }

func (n *Signature) Pos() token.Pos    { return n.Params.Pos() }
func (n *Signature) PosEnd() token.Pos { return n.Result.PosEnd() }

func (n *BuiltIn) Pos() token.Pos    { return n.TokPos }
func (n *BuiltIn) PosEnd() token.Pos { return n.Ident.PosEnd() }

func (n *Call) Pos() token.Pos    { return n.X.Pos() }
func (n *Call) PosEnd() token.Pos { return n.Args.PosEnd() }

func (n *Index) Pos() token.Pos    { return n.X.Pos() }
func (n *Index) PosEnd() token.Pos { return n.Args.PosEnd() }

func (n *Function) Pos() token.Pos    { return n.Signature.Pos() }
func (n *Function) PosEnd() token.Pos { return n.Body.PosEnd() }

func (n *Dot) Pos() token.Pos    { return n.X.Pos() }
func (n *Dot) PosEnd() token.Pos { return n.Y.PosEnd() }

func (n *Deref) Pos() token.Pos    { return n.X.Pos() }
func (n *Deref) PosEnd() token.Pos { return n.StarPos }

func (n *Op) Pos() token.Pos {
	if n.X != nil {
		return n.X.Pos()
	}
	return n.Start
}

func (n *Op) PosEnd() token.Pos {
	if n.Y != nil {
		return n.Y.PosEnd()
	}
	return n.End
}

//------------------------------------------------
// Lists
//------------------------------------------------

type (
	// Represents sequence of nodes, separated by comma.
	List struct {
		Nodes []Node
	}

	// Represents sequence of nodes, separated by semicolon\new line.
	StmtList struct {
		Nodes []Node
	}

	// Represents `[a, b, c]`.
	BracketList struct {
		*List
		Open, Close token.Pos // `[` and `]`.
	}

	// Represents `(a, b, c)`.
	ParenList struct {
		*List
		Open, Close token.Pos // `(` and `)`.
	}

	// Represents `{a; b; c}`.
	CurlyList struct {
		*StmtList
		Open, Close token.Pos // `{` and `}`.
	}
)

func (n *List) Pos() token.Pos    { return n.Nodes[0].Pos() }
func (n *List) PosEnd() token.Pos { return n.Nodes[len(n.Nodes)-1].PosEnd() }

func (n *StmtList) Pos() token.Pos    { return n.Nodes[0].Pos() }
func (n *StmtList) PosEnd() token.Pos { return n.Nodes[len(n.Nodes)-1].PosEnd() }

func (n *BracketList) Pos() token.Pos    { return n.Open }
func (n *BracketList) PosEnd() token.Pos { return n.Close }

func (n *ParenList) Pos() token.Pos    { return n.Open }
func (n *ParenList) PosEnd() token.Pos { return n.Close }

func (n *CurlyList) Pos() token.Pos    { return n.Open }
func (n *CurlyList) PosEnd() token.Pos { return n.Close }

//------------------------------------------------
// Language constructions
//------------------------------------------------

type (
	If struct {
		Cond   Node
		Body   *CurlyList
		Else   *Else
		TokPos token.Pos // `if` token.
	}

	Else struct {
		Body   Node      // Can be either [*If] or [*CurlyList].
		TokPos token.Pos // `else` token.
	}

	While struct {
		Cond   Node
		Body   *CurlyList
		TokPos token.Pos // `while` token.
	}

	For struct {
		DeclList *List
		IterExpr Node
		Body     *CurlyList
		TokPos   token.Pos // `for` token.
	}

	Return struct {
		X      Node      // optional
		TokPos token.Pos // `return` token.
	}

	Break struct {
		Label  *Ident
		TokPos token.Pos
	}

	Continue struct {
		Label  *Ident
		TokPos token.Pos
	}

	Import struct {
		Module *Ident
		TokPos token.Pos
	}
)

func (n *If) Pos() token.Pos { return n.TokPos }
func (n *If) PosEnd() token.Pos {
	if n.Else != nil {
		return n.Else.PosEnd()
	}
	return n.Body.PosEnd()
}

func (n *Else) Pos() token.Pos    { return n.TokPos }
func (n *Else) PosEnd() token.Pos { return n.Body.PosEnd() }

func (n *While) Pos() token.Pos    { return n.TokPos }
func (n *While) PosEnd() token.Pos { return n.Body.PosEnd() }

func (n *For) Pos() token.Pos    { return n.TokPos }
func (n *For) PosEnd() token.Pos { return n.Body.PosEnd() }

func (n *Return) Pos() token.Pos    { return n.TokPos }
func (n *Return) PosEnd() token.Pos { return n.X.PosEnd() }

func (n *Break) Pos() token.Pos { return n.TokPos }
func (n *Break) PosEnd() token.Pos {
	if n.Label != nil {
		return n.Label.PosEnd()
	}
	const length = uint32(len("break") - 1)
	end := n.TokPos
	end.Char += length
	end.Offset += uint64(length)
	return end
}

func (n *Continue) Pos() token.Pos { return n.TokPos }
func (n *Continue) PosEnd() token.Pos {
	if n.Label != nil {
		return n.Label.PosEnd()
	}
	const length = uint32(len("continue") - 1)
	end := n.TokPos
	end.Char += length
	end.Offset += uint64(length)
	return end
}

func (n *Import) Pos() token.Pos    { return n.TokPos }
func (n *Import) PosEnd() token.Pos { return n.Module.PosEnd() }

//-----------------------------------------------
// TODO name it
//-----------------------------------------------

func (*BadNode) implNode() {}
func (*Empty) implNode()   {}
func (*Ident) implNode()   {}
func (*Literal) implNode() {}

func (*Comment) implNode()       {}
func (*CommentGroup) implNode()  {}
func (*AttributeList) implNode() {}
func (*Decl) implNode()          {}

func (*ArrayType) implNode()  {}
func (*StructType) implNode() {}
func (*EnumType) implNode()   {}
func (*Signature) implNode()  {}
func (*BuiltIn) implNode()    {}
func (*Call) implNode()       {}
func (*Index) implNode()      {}
func (*Function) implNode()   {}
func (*Dot) implNode()        {}
func (*Deref) implNode()      {}
func (*Op) implNode()         {}

func (*List) implNode()        {}
func (*StmtList) implNode()    {}
func (*BracketList) implNode() {}
func (*ParenList) implNode()   {}
func (*CurlyList) implNode()   {}

func (*If) implNode()       {}
func (*Else) implNode()     {}
func (*While) implNode()    {}
func (*For) implNode()      {}
func (*Return) implNode()   {}
func (*Break) implNode()    {}
func (*Continue) implNode() {}
func (*Import) implNode()   {}
