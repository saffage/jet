package ast

import (
	"strings"

	"github.com/saffage/jet/token"
)

type Node interface {
	// Start of the entire tree. This location must also include nested nodes.
	Pos() token.Loc

	// End of the entire tree. This location must also include nested nodes.
	LocEnd() token.Loc

	// String representation of the node. This string must be equal to the
	// code from which this tree was parsed (ignoring location).
	Repr() string

	implNode()
}

func (*Decl) implNode() {}

// Exprs.

func (*BadNode) implNode() {}
func (*Empty) implNode()   {}
func (*Ident) implNode()   {}
func (*Literal) implNode() {}

func (*BuiltInCall) implNode() {}
func (*Function) implNode()    {}
func (*Call) implNode()        {}
func (*Index) implNode()       {}
func (*ArrayType) implNode()   {}
func (*StructType) implNode()  {}
func (*EnumType) implNode()    {}
func (*PointerType) implNode() {}
func (*Signature) implNode()   {}
func (*Dot) implNode()         {}
func (*Deref) implNode()       {}
func (*Op) implNode()          {}

func (*ParenList) implNode()   {}
func (*CurlyList) implNode()   {}
func (*BracketList) implNode() {}

func (*If) implNode()   {}
func (*Else) implNode() {}

// Stmts.

func (*Comment) implNode()      {}
func (*CommentGroup) implNode() {}

func (*StmtList) implNode()      {}
func (*List) implNode()          {}
func (*AttributeList) implNode() {}

func (*While) implNode()    {}
func (*For) implNode()      {}
func (*Return) implNode()   {}
func (*Break) implNode()    {}
func (*Continue) implNode() {}
func (*Import) implNode()   {}

type (
	Comment struct {
		Data       string
		Start, End token.Loc
	}

	CommentGroup struct {
		Comments []*Comment
	}

	//------------------------------------------------
	// Helper nodes
	//------------------------------------------------

	// Represents sequence of nodes, separated by semicolon\new line.
	StmtList struct {
		Nodes []Node
	}

	// Represents sequence of nodes, separated by comma.
	List struct {
		Nodes []Node
	}

	// Represents `@()`.
	AttributeList struct {
		List   *BracketList
		TokLoc token.Loc // `@` token.
	}

	//------------------------------------------------
	// Language constructions
	//------------------------------------------------

	While struct {
		Cond   Node
		Body   *CurlyList
		TokLoc token.Loc // `while` token.
	}

	For struct {
		DeclList *List
		IterExpr Node
		Body     *CurlyList
		TokLoc   token.Loc // `for` token.
	}

	Return struct {
		X      Node
		TokLoc token.Loc // `return` token.
	}

	Break struct {
		Label  *Ident
		TokLoc token.Loc // `break` token.
	}

	Continue struct {
		Label  *Ident
		TokLoc token.Loc // `continue` token.
	}

	Import struct {
		Module *Ident
		TokLoc token.Loc // `import` token.
	}
)

func (n *Comment) Pos() token.Loc    { return n.Start }
func (n *Comment) LocEnd() token.Loc { return n.End }

func (n *CommentGroup) Pos() token.Loc    { return n.Comments[0].Pos() }
func (n *CommentGroup) LocEnd() token.Loc { return n.Comments[len(n.Comments)-1].LocEnd() }

func (n *StmtList) Pos() token.Loc    { return n.Nodes[0].Pos() }
func (n *StmtList) LocEnd() token.Loc { return n.Nodes[len(n.Nodes)-1].LocEnd() }

func (n *List) Pos() token.Loc    { return n.Nodes[0].Pos() }
func (n *List) LocEnd() token.Loc { return n.Nodes[len(n.Nodes)-1].LocEnd() }

func (n *AttributeList) Pos() token.Loc    { return n.TokLoc }
func (n *AttributeList) LocEnd() token.Loc { return n.List.LocEnd() }

func (n *While) Pos() token.Loc    { return n.TokLoc }
func (n *While) LocEnd() token.Loc { return n.Body.LocEnd() }

func (n *For) Pos() token.Loc    { return n.TokLoc }
func (n *For) LocEnd() token.Loc { return n.Body.LocEnd() }

func (n *Return) Pos() token.Loc    { return n.TokLoc }
func (n *Return) LocEnd() token.Loc { return n.X.LocEnd() }

func (n *Break) Pos() token.Loc { return n.TokLoc }
func (n *Break) LocEnd() token.Loc {
	if n.Label != nil {
		return n.Label.LocEnd()
	}
	const length = uint32(len("break") - 1)
	end := n.TokLoc
	end.Char += length
	end.Offset += uint64(length)
	return end
}

func (n *Continue) Pos() token.Loc { return n.TokLoc }
func (n *Continue) LocEnd() token.Loc {
	if n.Label != nil {
		return n.Label.LocEnd()
	}
	const length = uint32(len("continue") - 1)
	end := n.TokLoc
	end.Char += length
	end.Offset += uint64(length)
	return end
}

func (n *Import) Pos() token.Loc    { return n.TokLoc }
func (n *Import) LocEnd() token.Loc { return n.Module.LocEnd() }

// Additional methods for nodes.

func (n *CommentGroup) Merged() string {
	buf := strings.Builder{}

	for _, comment := range n.Comments {
		buf.WriteString(comment.Data[1:])
	}

	return buf.String()
}
