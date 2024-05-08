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
	String() string

	implNode()
}

// Exprs.

func (*BadNode) implNode()  {}
func (*Empty) implNode()    {}
func (*Ident) implNode()    {}
func (*Literal) implNode()  {}
func (*Operator) implNode() {}

func (*BindingWithValue) implNode() {}
func (*Binding) implNode()          {}
func (*BuiltInCall) implNode()      {}
func (*Call) implNode()             {}
func (*Index) implNode()            {}
func (*ArrayType) implNode()        {}
func (*Signature) implNode()        {}
func (*MemberAccess) implNode()     {}
func (*PrefixOp) implNode()         {}
func (*InfixOp) implNode()          {}
func (*PostfixOp) implNode()        {}

func (*ParenList) implNode()   {}
func (*CurlyList) implNode()   {}
func (*BracketList) implNode() {}

func (*If) implNode()   {}
func (*Else) implNode() {}

// Decls.

func (*ModuleDecl) implNode()    {}
func (*VarDecl) implNode()       {}
func (*FuncDecl) implNode()      {}
func (*StructDecl) implNode()    {}
func (*TypeAliasDecl) implNode() {}

// Stmts.

func (*Comment) implNode()      {}
func (*CommentGroup) implNode() {}

func (*List) implNode()          {}
func (*ExprList) implNode()      {}
func (*AttributeList) implNode() {}

func (*While) implNode()    {}
func (*Return) implNode()   {}
func (*Break) implNode()    {}
func (*Continue) implNode() {}

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
	List struct {
		Nodes []Node
	}

	// Represents sequence of nodes, separated by comma.
	ExprList struct {
		Exprs []Node
	}

	// Represents `@()`.
	AttributeList struct {
		List *ParenList
		Loc  token.Loc // `@` token.
	}

	//------------------------------------------------
	// Language constructions
	//------------------------------------------------

	While struct {
		Cond Node
		Body Node
		Loc  token.Loc // `while` token.
	}

	Return struct {
		X   Node
		Loc token.Loc // `return` token.
	}

	Break struct {
		Label *Ident
		Loc   token.Loc // `break` token.
	}

	Continue struct {
		Label *Ident
		Loc   token.Loc // `continue` token.
	}
)

func (n *Comment) Pos() token.Loc    { return n.Start }
func (n *Comment) LocEnd() token.Loc { return n.End }

func (n *CommentGroup) Pos() token.Loc    { return n.Comments[0].Pos() }
func (n *CommentGroup) LocEnd() token.Loc { return n.Comments[len(n.Comments)-1].LocEnd() }

func (n *List) Pos() token.Loc    { return n.Nodes[0].Pos() }
func (n *List) LocEnd() token.Loc { return n.Nodes[len(n.Nodes)-1].LocEnd() }

func (n *ExprList) Pos() token.Loc    { return n.Exprs[0].Pos() }
func (n *ExprList) LocEnd() token.Loc { return n.Exprs[len(n.Exprs)-1].LocEnd() }

func (n *AttributeList) Pos() token.Loc    { return n.Loc }
func (n *AttributeList) LocEnd() token.Loc { return n.List.LocEnd() }

func (n *While) Pos() token.Loc    { return n.Loc }
func (n *While) LocEnd() token.Loc { return n.Body.LocEnd() }

func (n *Return) Pos() token.Loc    { return n.Loc }
func (n *Return) LocEnd() token.Loc { return n.X.LocEnd() }

func (n *Break) Pos() token.Loc { return n.Loc }
func (n *Break) LocEnd() token.Loc {
	if n.Label != nil {
		return n.Label.LocEnd()
	}
	const length = len("break") - 1
	end := n.Loc
	end.Char += length
	end.Offset += length
	return end
}

func (n *Continue) Pos() token.Loc { return n.Loc }
func (n *Continue) LocEnd() token.Loc {
	if n.Label != nil {
		return n.Label.LocEnd()
	}
	const length = len("continue") - 1
	end := n.Loc
	end.Char += length
	end.Offset += length
	return end
}

// Additional methods for nodes.

func (n *CommentGroup) Merged() string {
	buf := strings.Builder{}

	for _, comment := range n.Comments {
		buf.WriteString(comment.Data[1:])
	}

	return buf.String()
}
