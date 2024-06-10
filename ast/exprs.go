package ast

import "github.com/saffage/jet/token"

type (
	BadNode struct {
		DesiredLoc token.Loc
	}

	Empty struct {
		DesiredLoc token.Loc
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

	//------------------------------------------------
	// Composite nodes
	//------------------------------------------------

	// Represents `@foo()` or `@foo {}`.
	BuiltInCall struct {
		Name   *Ident
		Args   Node      // Either [ParenList] or [CurlyList].
		TokLoc token.Loc // `@` token.
	}

	Function struct {
		Signature *Signature
		Body      Node
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

	// Represents `struct{...}`.
	StructType struct {
		Fields []*Decl
		TokLoc token.Loc
		Open   token.Loc
		Close  token.Loc
	}

	// Represents `enum{...}`.
	EnumType struct {
		Fields []*Ident
		TokLoc token.Loc
		Open   token.Loc
		Close  token.Loc
	}

	// Represents `*X`.
	PointerType struct {
		X      Node
		TokLoc token.Loc
	}

	// Represents `() -> ()`.
	Signature struct {
		Params *ParenList
		Result Node // can be nil in some cases
	}

	// Represents `x.y`.
	Dot struct {
		X      Node
		Y      *Ident
		DotLoc token.Loc
	}

	// Represents `x.*`.
	Deref struct {
		X       Node
		DotLoc  token.Loc
		StarLoc token.Loc
	}

	// Represents `x ! y`, where `!` is an operator.
	Op struct {
		X     Node
		Y     Node
		Start token.Loc
		End   token.Loc
		Kind  OperatorKind
	}

	//------------------------------------------------
	// Lists
	//------------------------------------------------

	// Represents `[a, b, c]`.
	BracketList struct {
		*List
		Open, Close token.Loc // `[` and `]`.
	}

	// Represents `(a, b, c)`.
	ParenList struct {
		*List
		Open, Close token.Loc // `(` and `)`.
	}

	// Represents `{a; b; c}`.
	CurlyList struct {
		*StmtList
		Open, Close token.Loc // `{` and `}`.
	}

	//------------------------------------------------
	// Language constructions
	//------------------------------------------------

	If struct {
		Cond   Node
		Body   *CurlyList
		Else   *Else
		TokLoc token.Loc // `if` token.
	}

	Else struct {
		Body   Node      // Can be either [*If] or [*CurlyList].
		TokLoc token.Loc // `else` token.
	}
)

func (n *BadNode) Pos() token.Loc    { return n.DesiredLoc }
func (n *BadNode) LocEnd() token.Loc { return n.DesiredLoc }

func (n *Empty) Pos() token.Loc    { return n.DesiredLoc }
func (n *Empty) LocEnd() token.Loc { return n.DesiredLoc }

func (n *Ident) Pos() token.Loc    { return n.Start }
func (n *Ident) LocEnd() token.Loc { return n.End }

func (n *Literal) Pos() token.Loc    { return n.Start }
func (n *Literal) LocEnd() token.Loc { return n.End }

func (n *BuiltInCall) Pos() token.Loc    { return n.TokLoc }
func (n *BuiltInCall) LocEnd() token.Loc { return n.Args.LocEnd() }

func (n *Function) Pos() token.Loc    { return n.Signature.Pos() }
func (n *Function) LocEnd() token.Loc { return n.Body.LocEnd() }

func (n *Call) Pos() token.Loc    { return n.X.Pos() }
func (n *Call) LocEnd() token.Loc { return n.Args.LocEnd() }

func (n *Index) Pos() token.Loc    { return n.X.Pos() }
func (n *Index) LocEnd() token.Loc { return n.Args.LocEnd() }

func (n *ArrayType) Pos() token.Loc    { return n.Args.Pos() }
func (n *ArrayType) LocEnd() token.Loc { return n.X.LocEnd() }

func (n *StructType) Pos() token.Loc    { return n.TokLoc }
func (n *StructType) LocEnd() token.Loc { return n.Close }

func (n *EnumType) Pos() token.Loc    { return n.TokLoc }
func (n *EnumType) LocEnd() token.Loc { return n.Close }

func (n *Signature) Pos() token.Loc    { return n.Params.Pos() }
func (n *Signature) LocEnd() token.Loc { return n.Result.LocEnd() }

func (n *Dot) Pos() token.Loc    { return n.X.Pos() }
func (n *Dot) LocEnd() token.Loc { return n.Y.LocEnd() }

func (n *Deref) Pos() token.Loc    { return n.X.Pos() }
func (n *Deref) LocEnd() token.Loc { return n.StarLoc }

func (n *Op) Pos() token.Loc {
	if n.X != nil {
		return n.X.Pos()
	}
	return n.Start
}

func (n *Op) LocEnd() token.Loc {
	if n.Y != nil {
		return n.Y.LocEnd()
	}
	return n.End
}

func (n *ParenList) Pos() token.Loc    { return n.Open }
func (n *ParenList) LocEnd() token.Loc { return n.Close }

func (n *CurlyList) Pos() token.Loc    { return n.Open }
func (n *CurlyList) LocEnd() token.Loc { return n.Close }

func (n *BracketList) Pos() token.Loc    { return n.Open }
func (n *BracketList) LocEnd() token.Loc { return n.Close }

func (n *If) Pos() token.Loc { return n.TokLoc }
func (n *If) LocEnd() token.Loc {
	if n.Else != nil {
		return n.Else.LocEnd()
	}
	return n.Body.LocEnd()
}

func (n *Else) Pos() token.Loc    { return n.TokLoc }
func (n *Else) LocEnd() token.Loc { return n.Body.LocEnd() }
