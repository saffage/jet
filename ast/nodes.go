package ast

import (
	"strings"

	"github.com/saffage/jet/token"
)

type Node interface {
	Representable
	Pos() token.Pos    // Start of the entire tree.
	PosEnd() token.Pos // End of the entire tree.
	implNode()
}

type Ident interface {
	Node
	String() string
}

//------------------------------------------------
// Atoms
//------------------------------------------------

type (
	BadNode struct {
		DesiredPos token.Pos `yaml:"desired_pos"`
	}

	Empty struct {
		DesiredPos token.Pos `yaml:"desired_pos"`
	}

	Name struct {
		Data       string
		Start, End token.Pos
	}

	Type struct {
		Data       string
		Start, End token.Pos
	}

	Underscore struct {
		Data       string
		Start, End token.Pos
	}

	Literal struct {
		Data       string
		Kind       LiteralKind
		Start, End token.Pos
	}
)

func (n *BadNode) Pos() token.Pos    { return n.DesiredPos }
func (n *BadNode) PosEnd() token.Pos { return n.DesiredPos }

func (n *Empty) Pos() token.Pos    { return n.DesiredPos }
func (n *Empty) PosEnd() token.Pos { return n.DesiredPos }

func (n *Name) String() string    { return n.Data }
func (n *Name) Pos() token.Pos    { return n.Start }
func (n *Name) PosEnd() token.Pos { return n.End }

func (n *Type) String() string    { return n.Data }
func (n *Type) Pos() token.Pos    { return n.Start }
func (n *Type) PosEnd() token.Pos { return n.End }

func (n *Underscore) String() string    { return n.Data }
func (n *Underscore) Pos() token.Pos    { return n.Start }
func (n *Underscore) PosEnd() token.Pos { return n.End }

func (n *Literal) Pos() token.Pos    { return n.Start }
func (n *Literal) PosEnd() token.Pos { return n.End }

//------------------------------------------------
// Declaration
//------------------------------------------------

type (
	Comment struct {
		Value      string
		Start, End token.Pos
	}

	CommentGroup struct {
		Comments []*Comment
	}

	// Represents '@[...attributes]'.
	AttributeList struct {
		List   *List
		TokPos token.Pos // '@' token.
	}

	// Represents 'let name Type = value'.
	LetDecl struct {
		Attrs  *AttributeList `yaml:",omitempty"`
		LetTok token.Pos      ``
		Decl   *Decl          ``
		Value  Node           `yaml:",omitempty"`
	}

	// Represents 'type Name = Type'.
	TypeDecl struct {
		Attrs   *AttributeList `yaml:",omitempty"`
		TypeTok token.Pos      ``
		Name    *Type          ``
		Args    *Parens        `yaml:",omitempty"`
		Expr    Node           `yaml:",omitempty"`
	}

	// Represents 'name T' or just 'name'.
	Decl struct {
		Name Ident
		Type Node `yaml:",omitempty"` // Optional.
	}
)

func (n *Comment) Pos() token.Pos    { return n.Start }
func (n *Comment) PosEnd() token.Pos { return n.End }

func (n *CommentGroup) Pos() token.Pos    { return n.Comments[0].Pos() }
func (n *CommentGroup) PosEnd() token.Pos { return n.Comments[len(n.Comments)-1].PosEnd() }

func (n *AttributeList) Pos() token.Pos    { return n.TokPos }
func (n *AttributeList) PosEnd() token.Pos { return n.List.PosEnd() }

func (n *LetDecl) Pos() token.Pos    { return n.LetTok }
func (n *LetDecl) PosEnd() token.Pos { return n.Value.PosEnd() }

func (n *TypeDecl) Pos() token.Pos { return n.TypeTok }
func (n *TypeDecl) PosEnd() token.Pos {
	if n.Expr != nil {
		return n.Expr.PosEnd()
	}
	if n.Args != nil {
		return n.Args.PosEnd()
	}
	return n.Name.PosEnd()
}

func (n *Decl) Pos() token.Pos { return n.Name.Pos() }
func (n *Decl) PosEnd() token.Pos {
	if n.Type != nil {
		return n.Type.PosEnd()
	}
	return n.Name.PosEnd()
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
	Label struct {
		Label *Name
		X     Node
	}

	// Represents '[...args]x'.
	ArrayType struct {
		X    Node
		Args *List `yaml:",omitempty"`
	}

	// Represents 'struct {...fields}'.
	StructType struct {
		Fields []*LetDecl `yaml:",omitempty"`
		TokPos token.Pos
		Open   token.Pos
		Close  token.Pos
	}

	// Represents 'enum {...fields}'.
	EnumType struct {
		Fields []*Name `yaml:",omitempty"`
		TokPos token.Pos
		Open   token.Pos
		Close  token.Pos
	}

	// Represents '() -> ()'.
	Signature struct {
		Params *Parens `yaml:",omitempty"`
		Result Node    `yaml:",omitempty"` // can be nil in some cases
	}

	// Represents an identifier, prefixed with a '$' sign.
	BuiltIn struct {
		Name   *Name
		TokPos token.Pos // '$' token.
	}

	// Represents 'x(...args)'.
	Call struct {
		X    Node
		Args *Parens
	}

	// Represents 'x[...args]'.
	Index struct {
		X    Node
		Args *List
	}

	// Represents '(...params) -> T {...}' or '() expr'
	Function struct {
		*Signature
		Body Node
	}

	// Represents 'x.y'.
	Dot struct {
		X      Node
		Y      *Name
		DotPos token.Pos
	}

	// Represents 'x.*'.
	Deref struct {
		X       Node
		DotPos  token.Pos
		StarPos token.Pos
	}

	// Represents 'x OP y, where 'OP' is an operator.
	Op struct {
		X     Node `yaml:",omitempty"`
		Y     Node `yaml:",omitempty"`
		Start token.Pos
		End   token.Pos
		Kind  OperatorKind
	}
)

func (n *Label) Pos() token.Pos    { return n.Label.Start }
func (n *Label) PosEnd() token.Pos { return n.X.PosEnd() }

func (n *ArrayType) Pos() token.Pos    { return n.Args.Pos() }
func (n *ArrayType) PosEnd() token.Pos { return n.X.PosEnd() }

func (n *StructType) Pos() token.Pos    { return n.TokPos }
func (n *StructType) PosEnd() token.Pos { return n.Close }

func (n *EnumType) Pos() token.Pos    { return n.TokPos }
func (n *EnumType) PosEnd() token.Pos { return n.Close }

func (n *Signature) Pos() token.Pos    { return n.Params.Pos() }
func (n *Signature) PosEnd() token.Pos { return n.Result.PosEnd() }

func (n *BuiltIn) Pos() token.Pos    { return n.TokPos }
func (n *BuiltIn) PosEnd() token.Pos { return n.Name.PosEnd() }

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

func (n *Op) IsInfix() bool   { return n.X != nil && n.Y != nil }
func (n *Op) IsPrefix() bool  { return n.X != nil && n.Y == nil }
func (n *Op) IsPostfix() bool { return n.X == nil && n.Y != nil }

//------------------------------------------------
// Lists
//------------------------------------------------

type (
	// Represents 'a; b; c'.
	Stmts struct {
		Nodes []Node
		Start token.Pos // Beginning of the statement list.
	}

	// Represents '{ a; b; c }'.
	Block struct {
		Stmts *Stmts    // Optional.
		Open  token.Pos // '{'
		Close token.Pos // '}'
	}

	// Represents '[a, b, c]'.
	List struct {
		Nodes []Node
		Open  token.Pos // '['
		Close token.Pos // ']'
	}

	// Represents '(a, b, c)'.
	Parens struct {
		Nodes []Node
		Open  token.Pos // '('
		Close token.Pos // ')'
	}
)

func (n *Stmts) Pos() token.Pos { return n.Start }
func (n *Stmts) PosEnd() token.Pos {
	if len(n.Nodes) > 0 {
		return n.Nodes[len(n.Nodes)-1].PosEnd()
	}
	return n.Start
}

func (n *Block) Pos() token.Pos    { return n.Open }
func (n *Block) PosEnd() token.Pos { return n.Close }

func (n *List) Pos() token.Pos    { return n.Open }
func (n *List) PosEnd() token.Pos { return n.Close }

func (n *Parens) Pos() token.Pos    { return n.Open }
func (n *Parens) PosEnd() token.Pos { return n.Close }

//------------------------------------------------
// Language constructions
//------------------------------------------------

type (
	If struct {
		Cond   Node
		Body   *Stmts
		Else   *Else     `yaml:",omitempty"`
		TokPos token.Pos // 'if' token.
	}

	Else struct {
		Body   Node      // Can be either [*If] or [*CurlyList].
		TokPos token.Pos // 'else' token.
	}

	When struct {
		TokPos token.Pos ``
		Expr   Node      `yaml:",omitempty"`
		Body   *Block
	}

	Defer struct {
		X      Node
		TokPos token.Pos // 'defer' token.
	}

	Return struct {
		X      Node      `yaml:",omitempty"` // optional
		TokPos token.Pos // 'return' token.
	}

	Break struct {
		Label  *Name `yaml:",omitempty"`
		TokPos token.Pos
	}

	Continue struct {
		Label  *Name `yaml:",omitempty"`
		TokPos token.Pos
	}

	Import struct {
		Module *Name
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

func (n *When) Pos() token.Pos    { return n.TokPos }
func (n *When) PosEnd() token.Pos { return n.Body.PosEnd() }

func (n *Defer) Pos() token.Pos    { return n.TokPos }
func (n *Defer) PosEnd() token.Pos { return n.X.PosEnd() }

func (n *Return) Pos() token.Pos { return n.TokPos }
func (n *Return) PosEnd() token.Pos {
	if n.X != nil {
		return n.X.PosEnd()
	}
	const length = len("return") - 1
	end := n.TokPos
	end.Char += uint32(length)
	end.Offset += uint64(length)
	return end
}

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

func (*BadNode) implNode()    {}
func (*Empty) implNode()      {}
func (*Name) implNode()       {}
func (*Type) implNode()       {}
func (*Underscore) implNode() {}
func (*Literal) implNode()    {}

func (*AttributeList) implNode() {}
func (*LetDecl) implNode()       {}
func (*TypeDecl) implNode()      {}
func (*Decl) implNode()          {}

func (*Label) implNode()     {}
func (*Signature) implNode() {}
func (*Call) implNode()      {}
func (*Index) implNode()     {}
func (*Function) implNode()  {}
func (*Dot) implNode()       {}
func (*Op) implNode()        {}

func (*Stmts) implNode()  {}
func (*Block) implNode()  {}
func (*List) implNode()   {}
func (*Parens) implNode() {}

func (*When) implNode() {}

var (
	_ Node = (*BadNode)(nil)
	_ Node = (*Empty)(nil)
	_ Node = (*Name)(nil)
	_ Node = (*Type)(nil)
	_ Node = (*Underscore)(nil)
	_ Node = (*Literal)(nil)

	_ Node = (*AttributeList)(nil)
	_ Node = (*LetDecl)(nil)
	_ Node = (*TypeDecl)(nil)
	_ Node = (*Decl)(nil)

	_ Node = (*Label)(nil)
	_ Node = (*Signature)(nil)
	_ Node = (*Call)(nil)
	_ Node = (*Index)(nil)
	_ Node = (*Function)(nil)
	_ Node = (*Dot)(nil)
	_ Node = (*Op)(nil)

	_ Node = (*Stmts)(nil)
	_ Node = (*Block)(nil)
	_ Node = (*List)(nil)
	_ Node = (*Parens)(nil)

	_ Node = (*When)(nil)
)
