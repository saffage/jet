package ast

import "github.com/saffage/jet/token"

type (
	BadNode struct {
		Loc token.Loc // Desired location.
	}

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

	Operator struct {
		Start token.Loc
		End   token.Loc
		Kind  OperatorKind
	}

	//------------------------------------------------
	// Composite nodes
	//------------------------------------------------

	// Represents `name Type`.
	Binding struct {
		Attrs *AttributeList
		Name  *Ident
		Type  Node
	}

	// Represents `name Type = value`.
	BindingWithValue struct {
		*Binding
		Operator *Operator // Can be any assignment operator.
		Value    Node      // maybe nil.
	}

	// Represents `@foo()` or `@foo {}`.
	BuiltInCall struct {
		Name *Ident
		Args Node      // Either [ParenList] or [CurlyList].
		Loc  token.Loc // `@` token.
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

	// Represents `func(x T) T` or `(T) T`.
	Signature struct {
		Params *ParenList
		Result Node
		Loc    token.Loc // `func` token.
	}

	// Represents `x.selector` or `x.{...}`.
	MemberAccess struct {
		X        Node
		Selector Node      // Can be [*Ident] or [*CurlyList]
		Loc      token.Loc // `.` token.
	}

	// Represents `!x`, where `!` is an prefix operator.
	PrefixOp struct {
		X   Node
		Opr *Operator
	}

	// Represents `x ! y`, where `!` is a infix operator.
	InfixOp struct {
		X, Y Node
		Opr  *Operator
	}

	// Represents `x!`, where `!` is a postfix operator.
	PostfixOp struct {
		X   Node
		Opr *Operator
	}

	//------------------------------------------------
	// Lists
	//------------------------------------------------

	// Represents `[a, b, c]`.
	BracketList struct {
		*ExprList
		Open, Close token.Loc // `[` and `]`.
	}

	// Represents `(a, b, c)`.
	ParenList struct {
		*ExprList
		Open, Close token.Loc // `(` and `)`.
	}

	// Represents `{a; b; c}`.
	CurlyList struct {
		*List
		Open, Close token.Loc // `{` and `}`.
	}

	//------------------------------------------------
	// Language constructions
	//------------------------------------------------

	If struct {
		Cond Node
		Body *CurlyList
		Else *Else
		Loc  token.Loc // `if` token.
	}

	Else struct {
		Body Node      // Can be either [*If] or [*CurlyList].
		Loc  token.Loc // `else` token.
	}
)

func (n *BadNode) Pos() token.Loc    { return n.Loc }
func (n *BadNode) LocEnd() token.Loc { return n.Loc }

func (n *Empty) Pos() token.Loc    { return n.Loc }
func (n *Empty) LocEnd() token.Loc { return n.Loc }

func (n *Ident) Pos() token.Loc    { return n.Start }
func (n *Ident) LocEnd() token.Loc { return n.End }

func (n *Literal) Pos() token.Loc    { return n.Start }
func (n *Literal) LocEnd() token.Loc { return n.End }

func (n *Operator) Pos() token.Loc    { return n.Start }
func (n *Operator) LocEnd() token.Loc { return n.End }

func (n *BindingWithValue) Pos() token.Loc { return n.Binding.Pos() }
func (n *BindingWithValue) LocEnd() token.Loc {
	if n.Value != nil {
		return n.Value.LocEnd()
	}
	return n.Binding.LocEnd()
}

func (n *Binding) Pos() token.Loc { return n.Name.Pos() }
func (n *Binding) LocEnd() token.Loc {
	if n.Type != nil {
		return n.Type.LocEnd()
	}
	return n.Name.LocEnd()
}

func (n *BuiltInCall) Pos() token.Loc    { return n.Loc }
func (n *BuiltInCall) LocEnd() token.Loc { return n.Args.LocEnd() }

func (n *Call) Pos() token.Loc    { return n.X.Pos() }
func (n *Call) LocEnd() token.Loc { return n.Args.LocEnd() }

func (n *Index) Pos() token.Loc    { return n.X.Pos() }
func (n *Index) LocEnd() token.Loc { return n.Args.LocEnd() }

func (n *ArrayType) Pos() token.Loc    { return n.Args.Pos() }
func (n *ArrayType) LocEnd() token.Loc { return n.X.LocEnd() }

func (n *Signature) Pos() token.Loc {
	if n.Loc.Line == 0 {
		return n.Params.Pos()
	}
	return n.Loc
}
func (n *Signature) LocEnd() token.Loc { return n.Result.LocEnd() }

func (n *MemberAccess) Pos() token.Loc    { return n.X.Pos() }
func (n *MemberAccess) LocEnd() token.Loc { return n.Selector.LocEnd() }

func (n *PrefixOp) Pos() token.Loc    { return n.Opr.Pos() }
func (n *PrefixOp) LocEnd() token.Loc { return n.X.LocEnd() }

func (n *InfixOp) Pos() token.Loc    { return n.X.Pos() }
func (n *InfixOp) LocEnd() token.Loc { return n.Y.LocEnd() }

func (n *PostfixOp) Pos() token.Loc    { return n.X.Pos() }
func (n *PostfixOp) LocEnd() token.Loc { return n.Opr.LocEnd() }

func (n *ParenList) Pos() token.Loc    { return n.Open }
func (n *ParenList) LocEnd() token.Loc { return n.Close }

func (n *CurlyList) Pos() token.Loc    { return n.Open }
func (n *CurlyList) LocEnd() token.Loc { return n.Close }

func (n *BracketList) Pos() token.Loc    { return n.Open }
func (n *BracketList) LocEnd() token.Loc { return n.Close }

func (n *If) Pos() token.Loc { return n.Loc }
func (n *If) LocEnd() token.Loc {
	if n.Else != nil {
		return n.Else.LocEnd()
	}
	return n.Body.LocEnd()
}

func (n *Else) Pos() token.Loc    { return n.Loc }
func (n *Else) LocEnd() token.Loc { return n.Body.LocEnd() }
