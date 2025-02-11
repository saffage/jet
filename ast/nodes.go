package ast

import (
	"strings"

	"github.com/saffage/jet/text"
)

type Node interface {
	// Start of the entire tree. This position must also include nested nodes.
	Pos() text.Pos

	// End of the entire tree. This position must also include nested nodes.
	PosEnd() text.Pos

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
		DesiredPos text.Pos
	}

	Empty struct {
		DesiredPos text.Pos
	}

	Ident struct {
		Name       string
		Start, End text.Pos
	}

	Literal struct {
		Value      string
		Kind       LiteralKind
		Start, End text.Pos
	}
)

func (n *BadNode) Pos() text.Pos    { return n.DesiredPos }
func (n *BadNode) PosEnd() text.Pos { return n.DesiredPos }

func (n *Empty) Pos() text.Pos    { return n.DesiredPos }
func (n *Empty) PosEnd() text.Pos { return n.DesiredPos }

func (n *Ident) Pos() text.Pos    { return n.Start }
func (n *Ident) PosEnd() text.Pos { return n.End }

func (n *Literal) Pos() text.Pos    { return n.Start }
func (n *Literal) PosEnd() text.Pos { return n.End }

//------------------------------------------------
// Declaration
//------------------------------------------------

type (
	Comment struct {
		Value string
		Start text.Pos
		End   text.Pos
	}

	CommentGroup struct {
		Comments []*Comment
	}

	// Represents '@[...attributes]'.
	AttributeList struct {
		List   *BracketList
		TokLoc text.Pos // '@' token.
	}

	// Represents '@[..attributes] mut name: T = expr'.
	Decl struct {
		Attrs *AttributeList
		Ident *Ident
		Mut   text.Pos // optional
		Type  Node     // optional
		Value Node     // optional
	}
)

func (n *Comment) Pos() text.Pos    { return n.Start }
func (n *Comment) PosEnd() text.Pos { return n.End }

func (n *CommentGroup) Pos() text.Pos    { return n.Comments[0].Pos() }
func (n *CommentGroup) PosEnd() text.Pos { return n.Comments[len(n.Comments)-1].PosEnd() }

func (n *AttributeList) Pos() text.Pos    { return n.TokLoc }
func (n *AttributeList) PosEnd() text.Pos { return n.List.PosEnd() }

func (decl *Decl) Pos() text.Pos {
	if decl.Mut.IsValid() {
		return decl.Mut
	}
	return decl.Ident.Pos()
}

func (decl *Decl) PosEnd() text.Pos {
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
	// Represents '[...args]x'.
	ArrayType struct {
		X    Node
		Args *BracketList
	}

	// Represents 'struct {...fields}'.
	StructType struct {
		Fields []*Decl
		TokPos text.Pos
		Open   text.Pos
		Close  text.Pos
	}

	// Represents 'enum {...fields}'.
	EnumType struct {
		Fields []*Ident
		TokPos text.Pos
		Open   text.Pos
		Close  text.Pos
	}

	// Represents '() -> ()'.
	Signature struct {
		Params *ParenList
		Result Node // can be nil in some cases
	}

	// Represents an identifier, prefixed with a '$' sign.
	BuiltIn struct {
		*Ident
		TokPos text.Pos // '$' token.
	}

	// Represents 'x(...args)'.
	Call struct {
		X    Node
		Args *ParenList
	}

	// Represents 'x[...args]'.
	Index struct {
		X    Node
		Args *BracketList
	}

	// Represents '(...params) -> T {...}' or '() expr'
	Function struct {
		*Signature
		Body Node
	}

	// Represents 'x.y'.
	Dot struct {
		X      Node
		Y      *Ident
		DotPos text.Pos
	}

	// Represents 'x.*'.
	Deref struct {
		X       Node
		DotPos  text.Pos
		StarPos text.Pos
	}

	// Represents 'x OP y', where 'OP' is an operator.
	Op struct {
		X     Node
		Y     Node
		Start text.Pos
		End   text.Pos
		Kind  OperatorKind
	}
)

func (n *ArrayType) Pos() text.Pos    { return n.Args.Pos() }
func (n *ArrayType) PosEnd() text.Pos { return n.X.PosEnd() }

func (n *StructType) Pos() text.Pos    { return n.TokPos }
func (n *StructType) PosEnd() text.Pos { return n.Close }

func (n *EnumType) Pos() text.Pos    { return n.TokPos }
func (n *EnumType) PosEnd() text.Pos { return n.Close }

func (n *Signature) Pos() text.Pos    { return n.Params.Pos() }
func (n *Signature) PosEnd() text.Pos { return n.Result.PosEnd() }

func (n *BuiltIn) Pos() text.Pos    { return n.TokPos }
func (n *BuiltIn) PosEnd() text.Pos { return n.Ident.PosEnd() }

func (n *Call) Pos() text.Pos    { return n.X.Pos() }
func (n *Call) PosEnd() text.Pos { return n.Args.PosEnd() }

func (n *Index) Pos() text.Pos    { return n.X.Pos() }
func (n *Index) PosEnd() text.Pos { return n.Args.PosEnd() }

func (n *Function) Pos() text.Pos    { return n.Signature.Pos() }
func (n *Function) PosEnd() text.Pos { return n.Body.PosEnd() }

func (n *Dot) Pos() text.Pos    { return n.X.Pos() }
func (n *Dot) PosEnd() text.Pos { return n.Y.PosEnd() }

func (n *Deref) Pos() text.Pos    { return n.X.Pos() }
func (n *Deref) PosEnd() text.Pos { return n.StarPos }

func (n *Op) Pos() text.Pos {
	if n.X != nil {
		return n.X.Pos()
	}
	return n.Start
}

func (n *Op) PosEnd() text.Pos {
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

	// Represents '[a, b, c]'.
	BracketList struct {
		*List
		Open, Close text.Pos // '[' and ']'.
	}

	// Represents '(a, b, c)'.
	ParenList struct {
		*List
		Open, Close text.Pos // '(' and ')'.
	}

	// Represents '{a; b; c}'.
	CurlyList struct {
		*StmtList
		Open, Close text.Pos // '{' and '}'.
	}
)

func (n *List) Pos() text.Pos    { return n.Nodes[0].Pos() }
func (n *List) PosEnd() text.Pos { return n.Nodes[len(n.Nodes)-1].PosEnd() }

func (n *StmtList) Pos() text.Pos    { return n.Nodes[0].Pos() }
func (n *StmtList) PosEnd() text.Pos { return n.Nodes[len(n.Nodes)-1].PosEnd() }

func (n *BracketList) Pos() text.Pos    { return n.Open }
func (n *BracketList) PosEnd() text.Pos { return n.Close }

func (n *ParenList) Pos() text.Pos    { return n.Open }
func (n *ParenList) PosEnd() text.Pos { return n.Close }

func (n *CurlyList) Pos() text.Pos    { return n.Open }
func (n *CurlyList) PosEnd() text.Pos { return n.Close }

//------------------------------------------------
// Language constructions
//------------------------------------------------

type (
	If struct {
		Cond   Node
		Body   *CurlyList
		Else   *Else
		TokPos text.Pos // 'if' token.
	}

	Else struct {
		Body   Node     // Can be either [*If] or [*CurlyList].
		TokPos text.Pos // 'else' token.
	}

	While struct {
		Cond   Node
		Body   *CurlyList
		TokPos text.Pos // 'while' token.
	}

	For struct {
		DeclList *List
		IterExpr Node
		Body     *CurlyList
		TokPos   text.Pos // 'for' token.
	}

	Defer struct {
		X      Node
		TokPos text.Pos // 'defer' token.
	}

	Return struct {
		X      Node     // optional
		TokPos text.Pos // 'return' token.
	}

	Break struct {
		Label  *Ident
		TokPos text.Pos
	}

	Continue struct {
		Label  *Ident
		TokPos text.Pos
	}

	Import struct {
		Module *Ident
		TokPos text.Pos
	}
)

func (n *If) Pos() text.Pos { return n.TokPos }
func (n *If) PosEnd() text.Pos {
	if n.Else != nil {
		return n.Else.PosEnd()
	}
	return n.Body.PosEnd()
}

func (n *Else) Pos() text.Pos    { return n.TokPos }
func (n *Else) PosEnd() text.Pos { return n.Body.PosEnd() }

func (n *While) Pos() text.Pos    { return n.TokPos }
func (n *While) PosEnd() text.Pos { return n.Body.PosEnd() }

func (n *For) Pos() text.Pos    { return n.TokPos }
func (n *For) PosEnd() text.Pos { return n.Body.PosEnd() }

func (n *Defer) Pos() text.Pos    { return n.TokPos }
func (n *Defer) PosEnd() text.Pos { return n.X.PosEnd() }

func (n *Return) Pos() text.Pos { return n.TokPos }
func (n *Return) PosEnd() text.Pos {
	if n.X != nil {
		return n.X.PosEnd()
	}
	const length = len("return") - 1
	return text.PosFrom(n.TokPos.ID(), n.TokPos.Offset()+length)
}

func (n *Break) Pos() text.Pos { return n.TokPos }
func (n *Break) PosEnd() text.Pos {
	if n.Label != nil {
		return n.Label.PosEnd()
	}
	const length = len("break") - 1
	return text.PosFrom(n.TokPos.ID(), n.TokPos.Offset()+length)
}

func (n *Continue) Pos() text.Pos { return n.TokPos }
func (n *Continue) PosEnd() text.Pos {
	if n.Label != nil {
		return n.Label.PosEnd()
	}
	const length = len("continue") - 1
	return text.PosFrom(n.TokPos.ID(), n.TokPos.Offset()+length)
}

func (n *Import) Pos() text.Pos    { return n.TokPos }
func (n *Import) PosEnd() text.Pos { return n.Module.PosEnd() }

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
func (*Defer) implNode()    {}
func (*Return) implNode()   {}
func (*Break) implNode()    {}
func (*Continue) implNode() {}
func (*Import) implNode()   {}
