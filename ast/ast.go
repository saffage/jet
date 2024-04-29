package ast

import (
	"github.com/saffage/jet/token"
)

type (
	Node interface {
		// Start of the entire tree. This location must also include nested nodes.
		Pos() token.Loc

		// End of the entire tree. This location must also include nested nodes.
		PosEnd() token.Loc

		// String representation of the node. This string must be equal to the
		// code from which this tree was parsed (ignoring location).
		String() string

		implNode()
	}

	// Invalid node.
	BadNode struct {
		Loc token.Loc // Desired location.
	}

	// Empty statement.
	Empty struct {
		Loc token.Loc // Desired location.
	}

	Ident struct {
		Name       string
		Start, End token.Loc
	}

	Literal struct {
		Value      string
		Kind       LiteralKind
		Start, End token.Loc
	}

	PrefixOpr struct {
		Start token.Loc
		End   token.Loc
		Kind  PrefixOpKind
	}

	InfixOpr struct {
		Start token.Loc
		End   token.Loc
		Kind  InfixOpKind
	}

	PostfixOpr struct {
		Start token.Loc
		End   token.Loc
		Kind  PostfixOpKind
	}

	// Used in `MemberAccess` nodes to represent `*` suffix.
	Star struct {
		Loc token.Loc // `*` token.
	}

	Comment struct {
		Data       string
		Start, End token.Loc
	}

	CommentGroup struct {
		Comments []*Comment
	}
)

func (*BadNode) implNode()      {}
func (*Empty) implNode()        {}
func (*Ident) implNode()        {}
func (*Literal) implNode()      {}
func (*PrefixOpr) implNode()    {}
func (*InfixOpr) implNode()     {}
func (*PostfixOpr) implNode()   {}
func (*Star) implNode()         {}
func (*Comment) implNode()      {}
func (*CommentGroup) implNode() {}

func (n *BadNode) Pos() token.Loc    { return n.Loc }
func (n *BadNode) PosEnd() token.Loc { return n.Loc }

func (n *Empty) Pos() token.Loc    { return n.Loc }
func (n *Empty) PosEnd() token.Loc { return n.Loc }

func (n *Ident) Pos() token.Loc    { return n.Start }
func (n *Ident) PosEnd() token.Loc { return n.End }

func (n *Literal) Pos() token.Loc    { return n.Start }
func (n *Literal) PosEnd() token.Loc { return n.End }

func (n *PrefixOpr) Pos() token.Loc    { return n.Start }
func (n *PrefixOpr) PosEnd() token.Loc { return n.End }

func (n *InfixOpr) Pos() token.Loc    { return n.Start }
func (n *InfixOpr) PosEnd() token.Loc { return n.End }

func (n *PostfixOpr) Pos() token.Loc    { return n.Start }
func (n *PostfixOpr) PosEnd() token.Loc { return n.End }

func (n *Star) Pos() token.Loc    { return n.Loc }
func (n *Star) PosEnd() token.Loc { return n.Loc }

func (n *Comment) Pos() token.Loc    { return n.Start }
func (n *Comment) PosEnd() token.Loc { return n.End }

func (n *CommentGroup) Pos() token.Loc    { return n.Comments[0].Pos() }
func (n *CommentGroup) PosEnd() token.Loc { return n.Comments[len(n.Comments)-1].PosEnd() }

type (
	// Represents `a, b Type = value`.
	Field struct {
		Names []*Ident
		Type  Node // maybe nil.
		Value Node // maybe nil.
	}

	// Represents `func(x T) T` or `(T) T`.
	Signature struct {
		Params *ParenList
		Result Node
		Loc    token.Loc // `func` token.
	}

	// Represents `x(args)`.
	Call struct {
		X    Node
		Args *ParenList
	}

	// Represents `x[args]`.
	Index struct {
		X    Node
		Args *BracketList
	}

	// Represents `[args]x`.
	ArrayType struct {
		X    Node
		Args *BracketList
	}

	// Represents `x.selector`.
	MemberAccess struct {
		X        Node
		Selector Node
		Loc      token.Loc // `.` token.
	}

	// Represents `!x`, where `!` is an unary operator.
	PrefixOp struct {
		X   Node
		Opr *PrefixOpr
	}

	// Represents `x ! y`, where `!` is a binary operator.
	InfixOp struct {
		X, Y Node
		Opr  *InfixOpr
	}

	PostfixOp struct {
		X   Node
		Opr *PostfixOpr
	}
)

func (*Field) implNode()        {}
func (*Signature) implNode()    {}
func (*Call) implNode()         {}
func (*Index) implNode()        {}
func (*ArrayType) implNode()    {}
func (*MemberAccess) implNode() {}
func (*PrefixOp) implNode()     {}
func (*InfixOp) implNode()      {}
func (*PostfixOp) implNode()    {}

func (n *Field) Pos() token.Loc { return n.Names[0].Pos() }
func (n *Field) PosEnd() token.Loc {
	if n.Value != nil {
		return n.Value.PosEnd()
	}
	if n.Type != nil {
		return n.Type.PosEnd()
	}
	panic("node must have as least a type or a value")
}

func (n *Signature) Pos() token.Loc {
	if n.Loc.Line == 0 {
		return n.Params.Pos()
	}
	return n.Loc
}
func (n *Signature) PosEnd() token.Loc { return n.Result.PosEnd() }

func (n *MemberAccess) Pos() token.Loc    { return n.X.Pos() }
func (n *MemberAccess) PosEnd() token.Loc { return n.Selector.PosEnd() }

func (n *Call) Pos() token.Loc    { return n.X.Pos() }
func (n *Call) PosEnd() token.Loc { return n.Args.PosEnd() }

func (n *Index) Pos() token.Loc    { return n.X.Pos() }
func (n *Index) PosEnd() token.Loc { return n.Args.PosEnd() }

func (n *ArrayType) Pos() token.Loc    { return n.Args.Pos() }
func (n *ArrayType) PosEnd() token.Loc { return n.X.PosEnd() }

func (n *PrefixOp) Pos() token.Loc    { return n.Opr.Pos() }
func (n *PrefixOp) PosEnd() token.Loc { return n.X.PosEnd() }

func (n *InfixOp) Pos() token.Loc    { return n.X.Pos() }
func (n *InfixOp) PosEnd() token.Loc { return n.Y.PosEnd() }

func (n *PostfixOp) Pos() token.Loc    { return n.X.Pos() }
func (n *PostfixOp) PosEnd() token.Loc { return n.Opr.PosEnd() }

type (
	// Represents sequence of nodes, separated by comma or semicolon\new line.
	List struct {
		Nodes []Node
	}

	// Represents `(a, b, c)`.
	ParenList struct {
		*List
		Open, Close token.Loc // `(` and `)`.
	}

	// Represents `{a, b, c}`.
	CurlyList struct {
		*List
		Open, Close token.Loc // `{` and `}`.
	}

	// Represents `[a, b, c]`.
	BracketList struct {
		*List
		Open, Close token.Loc // `[` and `]`.
	}
)

func (*List) implNode()        {}
func (*ParenList) implNode()   {}
func (*CurlyList) implNode()   {}
func (*BracketList) implNode() {}

func (n *List) Pos() token.Loc    { return n.Nodes[0].Pos() }
func (n *List) PosEnd() token.Loc { return n.Nodes[len(n.Nodes)-1].PosEnd() }

func (n *ParenList) Pos() token.Loc    { return n.Open }
func (n *ParenList) PosEnd() token.Loc { return n.Close }

func (n *CurlyList) Pos() token.Loc    { return n.Open }
func (n *CurlyList) PosEnd() token.Loc { return n.Close }

func (n *BracketList) Pos() token.Loc    { return n.Open }
func (n *BracketList) PosEnd() token.Loc { return n.Close }

type (
	// Represents `@()`.
	AttributeList struct {
		List *ParenList
		Loc  token.Loc // `@` token.
	}

	// Represents `@foo()` or `@foo {}`.
	BuiltInCall struct {
		Name *Ident
		X    Node      // Either [ParenList] or [CurlyList].
		Loc  token.Loc // `@` token.
	}
)

func (*AttributeList) implNode() {}
func (*BuiltInCall) implNode()   {}

func (n *AttributeList) Pos() token.Loc    { return n.Loc }
func (n *AttributeList) PosEnd() token.Loc { return n.List.PosEnd() }

func (n *BuiltInCall) Pos() token.Loc    { return n.Loc }
func (n *BuiltInCall) PosEnd() token.Loc { return n.X.PosEnd() }

type (
	Decl interface {
		Node
		Ident() *Ident
		Doc() string
		Attributes() *AttributeList
	}

	ModuleDecl struct {
		Attrs *AttributeList
		Name  *Ident
		Body  Node      // Can be either [CurlyList] or [List].
		Loc   token.Loc // `module` token.
	}

	GenericDecl struct {
		Attrs *AttributeList
		Field *Field
		Loc   token.Loc // `const`, `var`, `val` token.
		Kind  GenericDeclKind
	}

	FuncDecl struct {
		Attrs     *AttributeList
		Name      *Ident
		Signature *Signature
		Body      Node
		Loc       token.Loc // `func` token.
	}

	TypeAliasDecl struct {
		Attrs *AttributeList
		Name  *Ident
		Expr  Node
		Loc   token.Loc // `alias` token.
	}
)

func (*ModuleDecl) implNode()    {}
func (*GenericDecl) implNode()   {}
func (*FuncDecl) implNode()      {}
func (*TypeAliasDecl) implNode() {}

func (n *ModuleDecl) Pos() token.Loc             { return n.Loc }
func (n *ModuleDecl) PosEnd() token.Loc          { return n.Name.PosEnd() }
func (n *ModuleDecl) Ident() *Ident              { return n.Name }
func (*ModuleDecl) Doc() string                  { return "" }
func (n *ModuleDecl) Attributes() *AttributeList { return n.Attrs }

func (n *GenericDecl) Pos() token.Loc    { return n.Loc }
func (n *GenericDecl) PosEnd() token.Loc { return n.Field.PosEnd() }
func (n *GenericDecl) Ident() *Ident {
	if len(n.Field.Names) > 0 {
		return n.Field.Names[0]
	}
	return nil
}
func (*GenericDecl) Doc() string                  { return "" }
func (n *GenericDecl) Attributes() *AttributeList { return n.Attrs }

func (n *FuncDecl) Pos() token.Loc { return n.Loc }
func (n *FuncDecl) PosEnd() token.Loc {
	if n.Body != nil {
		return n.Body.PosEnd()
	}
	return n.Signature.PosEnd()
}
func (n *FuncDecl) Ident() *Ident              { return n.Name }
func (*FuncDecl) Doc() string                  { return "" }
func (n *FuncDecl) Attributes() *AttributeList { return n.Attrs }

func (n *TypeAliasDecl) Pos() token.Loc             { return n.Loc }
func (n *TypeAliasDecl) PosEnd() token.Loc          { return n.Expr.PosEnd() }
func (n *TypeAliasDecl) Ident() *Ident              { return n.Name }
func (*TypeAliasDecl) Doc() string                  { return "" }
func (n *TypeAliasDecl) Attributes() *AttributeList { return n.Attrs }

type (
	If struct {
		Cond Node
		Body Node
		Else *Else
		Loc  token.Loc // `if` token.
	}

	Else struct {
		Body Node
		Loc  token.Loc // `else` token.
	}

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

func (*If) implNode()       {}
func (*Else) implNode()     {}
func (*While) implNode()    {}
func (*Return) implNode()   {}
func (*Break) implNode()    {}
func (*Continue) implNode() {}

func (n *If) Pos() token.Loc { return n.Loc }
func (n *If) PosEnd() token.Loc {
	if n.Else != nil {
		return n.Else.PosEnd()
	}
	return n.Body.PosEnd()
}

func (n *Else) Pos() token.Loc    { return n.Loc }
func (n *Else) PosEnd() token.Loc { return n.Body.PosEnd() }

func (n *While) Pos() token.Loc    { return n.Loc }
func (n *While) PosEnd() token.Loc { return n.Body.PosEnd() }

func (n *Return) Pos() token.Loc    { return n.Loc }
func (n *Return) PosEnd() token.Loc { return n.X.PosEnd() }

func (n *Break) Pos() token.Loc { return n.Loc }
func (n *Break) PosEnd() token.Loc {
	if n.Label != nil {
		return n.Label.PosEnd()
	}
	const length = len("break") - 1
	end := n.Loc
	end.Char += length
	end.Offset += length
	return end
}

func (n *Continue) Pos() token.Loc { return n.Loc }
func (n *Continue) PosEnd() token.Loc {
	if n.Label != nil {
		return n.Label.PosEnd()
	}
	const length = len("continue") - 1
	end := n.Loc
	end.Char += length
	end.Offset += length
	return end
}
