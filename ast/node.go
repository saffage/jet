package ast

import "github.com/saffage/jet/parser/token"

type Node interface {
	Range() token.Range // Entire tree range.
	Pos() token.Pos     // Start of the entire tree.
	PosEnd() token.Pos  // End of the entire tree.

	representable
	walkable
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
		DesiredRange token.Range `yaml:"desired_range"`
	}

	Empty struct {
		DesiredRange token.Range `yaml:"desired_range"`
	}

	Lower struct {
		Data string
		Rng  token.Range
	}

	Upper struct {
		Data string
		Rng  token.Range
	}

	Underscore struct {
		Data string
		Rng  token.Range
	}

	TypeVar struct {
		Data string
		Rng  token.Range
	}

	Literal struct {
		Data string
		Kind LiteralKind
		Rng  token.Range
	}
)

func (n *BadNode) Range() token.Range { return n.DesiredRange }
func (n *BadNode) Pos() token.Pos     { return n.DesiredRange.StartPos() }
func (n *BadNode) PosEnd() token.Pos  { return n.DesiredRange.EndPos() }

func (n *Empty) Range() token.Range { return n.DesiredRange }
func (n *Empty) Pos() token.Pos     { return n.DesiredRange.StartPos() }
func (n *Empty) PosEnd() token.Pos  { return n.DesiredRange.EndPos() }

func (n *Lower) Range() token.Range { return n.Rng }
func (n *Lower) Pos() token.Pos     { return n.Rng.StartPos() }
func (n *Lower) PosEnd() token.Pos  { return n.Rng.EndPos() }
func (n *Lower) String() string     { return n.Data }

func (n *Upper) Range() token.Range { return n.Rng }
func (n *Upper) Pos() token.Pos     { return n.Rng.StartPos() }
func (n *Upper) PosEnd() token.Pos  { return n.Rng.EndPos() }
func (n *Upper) String() string     { return n.Data }

func (n *Underscore) Range() token.Range { return n.Rng }
func (n *Underscore) Pos() token.Pos     { return n.Rng.StartPos() }
func (n *Underscore) PosEnd() token.Pos  { return n.Rng.EndPos() }
func (n *Underscore) String() string     { return n.Data }

func (n *TypeVar) Range() token.Range { return n.Rng }
func (n *TypeVar) Pos() token.Pos     { return n.Rng.StartPos() }
func (n *TypeVar) PosEnd() token.Pos  { return n.Rng.EndPos() }
func (n *TypeVar) String() string     { return n.Data }

func (n *Literal) Range() token.Range { return n.Rng }
func (n *Literal) Pos() token.Pos     { return n.Rng.StartPos() }
func (n *Literal) PosEnd() token.Pos  { return n.Rng.EndPos() }

//------------------------------------------------
// Declaration
//------------------------------------------------

type (
	// Represents '@[...attributes]'.
	AttributeList struct {
		List   *List
		TokPos token.Pos // '@' token.
	}

	// Represents 'let name Type = value'.
	LetDecl struct {
		Attrs  *AttributeList `yaml:",omitempty"`
		LetTok token.Pos
		Decl   *Decl
		Value  Node
	}

	// Represents 'type Name = Type' or 'type Name(args) { variants }'.
	TypeDecl struct {
		Attrs   *AttributeList `yaml:",omitempty"`
		Name    *Upper         ``
		Args    *Parens        `yaml:",omitempty"`
		Expr    Node           `yaml:",omitempty"`
		TypeTok token.Pos      ``
		EqTok   token.Pos      ``
	}

	// Represents `name T`, `name`, `type 'name T`, `type 'name`.
	Decl struct {
		TypeTok token.Pos // When TypeTok is not zero value, name is *TypeVar
		Name    Ident     ``
		Type    Node      `yaml:",omitempty"` // Optional.
	}

	// Represents 'Name' or 'Name(T)'.
	Variant struct {
		Name   *Upper
		Params *Parens `yaml:",omitempty"` // Optional.
	}
)

func (node *AttributeList) Range() token.Range { return node.TokPos.WithEnd(node.TokPos) }
func (node *AttributeList) Pos() token.Pos     { return node.TokPos }
func (node *AttributeList) PosEnd() token.Pos  { return node.List.PosEnd() }

func (node *LetDecl) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *LetDecl) Pos() token.Pos     { return node.LetTok }
func (node *LetDecl) PosEnd() token.Pos  { return node.Value.PosEnd() }

func (node *TypeDecl) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *TypeDecl) Pos() token.Pos     { return node.TypeTok }
func (node *TypeDecl) PosEnd() token.Pos {
	if node.Args != nil {
		return node.Args.PosEnd()
	}
	return node.Name.PosEnd()
}

func (node *Decl) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Decl) Pos() token.Pos {
	if node.TypeTok.IsValid() {
		return node.TypeTok
	}
	return node.Name.Pos()
}
func (node *Decl) PosEnd() token.Pos {
	if node.Type != nil {
		return node.Type.PosEnd()
	}
	return node.Name.PosEnd()
}

func (node *Variant) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Variant) Pos() token.Pos     { return node.Name.Pos() }
func (node *Variant) PosEnd() token.Pos {
	if node.Params != nil {
		return node.Params.PosEnd()
	}
	return node.Name.PosEnd()
}

//------------------------------------------------
// Composite nodes
//------------------------------------------------

type (
	Label struct {
		Name *Lower
		X    Node
	}

	// Represents '() T with Effects'.
	Signature struct {
		Params  *Parens
		Result  Node `yaml:",omitempty"` // can be nil in some cases
		Effects Node `yaml:",omitempty"` // can be nil in some cases
	}

	// Represents 'x(...args)'.
	Call struct {
		X    Node
		Args *Parens
	}

	// Represents 'x.y'.
	Dot struct {
		X      Node
		Y      *Lower
		DotPos token.Pos
	}

	// Represents 'x OP y, where 'OP' is an operator.
	Op struct {
		X    Node `yaml:",omitempty"`
		Y    Node `yaml:",omitempty"`
		Rng  token.Range
		Kind OperatorKind
	}

	// AnonFunc struct {
	// 	Sig   *Signature
	// 	Body  []Node
	// 	Arrow token.Pos
	// 	Open  token.Pos
	// 	Close token.Pos
	// }
)

func (node *Label) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Label) Pos() token.Pos     { return node.Name.Pos() }
func (node *Label) PosEnd() token.Pos  { return node.X.PosEnd() }
func (node *Label) Label() *Lower {
	if node.Name != nil {
		return node.Name
	}

	switch x := node.X.(type) {
	case *Lower:
		return x
	case *Decl:
		if name, _ := x.Name.(*Lower); name != nil {
			return name
		}
	}

	return nil
}

func (node *Signature) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Signature) Pos() token.Pos     { return node.Params.Pos() }
func (node *Signature) PosEnd() token.Pos {
	if node.Effects != nil {
		return node.Effects.PosEnd()
	}
	if node.Result != nil {
		return node.Result.PosEnd()
	}
	return node.Result.PosEnd()
}

func (node *Call) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Call) Pos() token.Pos     { return node.X.Pos() }
func (node *Call) PosEnd() token.Pos  { return node.Args.PosEnd() }

func (node *Dot) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Dot) Pos() token.Pos     { return node.X.Pos() }
func (node *Dot) PosEnd() token.Pos  { return node.Y.PosEnd() }

func (node *Op) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Op) Pos() token.Pos {
	if node.X != nil {
		return node.X.Pos()
	}
	return node.Rng.StartPos()
}
func (node *Op) PosEnd() token.Pos {
	if node.Y != nil {
		return node.Y.PosEnd()
	}
	return node.Rng.EndPos()
}

func (n *Op) IsInfix() bool   { return n.X != nil && n.Y != nil }
func (n *Op) IsPrefix() bool  { return n.X != nil && n.Y == nil }
func (n *Op) IsPostfix() bool { return n.X == nil && n.Y != nil }
func (n *Op) IsName() bool    { return n.X == nil && n.Y == nil }

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
		Stmts *Stmts // Can be nil.
		Rng   token.Range
	}

	// Represents '[a, b, c]'.
	List struct {
		Nodes []Node
		Rng   token.Range
	}

	// Represents '(a, b, c)'.
	Parens struct {
		Nodes []Node
		Rng   token.Range
	}
)

func (node *Stmts) Range() token.Range { return token.RangeFrom(node.Start, node.PosEnd()) }
func (node *Stmts) Pos() token.Pos     { return node.Start }
func (node *Stmts) PosEnd() token.Pos {
	if len(node.Nodes) > 0 {
		return node.Nodes[len(node.Nodes)-1].PosEnd()
	}
	return node.Start
}

func (node *Block) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Block) Pos() token.Pos     { return node.Rng.StartPos() }
func (node *Block) PosEnd() token.Pos  { return node.Rng.EndPos() }

func (node *List) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *List) Pos() token.Pos     { return node.Rng.StartPos() }
func (node *List) PosEnd() token.Pos  { return node.Rng.EndPos() }

func (node *Parens) Range() token.Range { return node.Pos().WithEnd(node.PosEnd()) }
func (node *Parens) Pos() token.Pos     { return node.Rng.StartPos() }
func (node *Parens) PosEnd() token.Pos  { return node.Rng.EndPos() }

//------------------------------------------------
// Language constructions
//------------------------------------------------

type (
	When struct {
		TokPos token.Pos ``
		Expr   Node      `yaml:",omitempty"`
		Body   *Block
	}

	Extern struct {
		TokPos token.Pos ``
		Args   *Parens   `yaml:",omitempty"`
	}
)

func (node *When) Range() token.Range { return node.TokPos.WithEnd(node.PosEnd()) }
func (node *When) Pos() token.Pos     { return node.TokPos }
func (node *When) PosEnd() token.Pos  { return node.Body.PosEnd() }

func (node *Extern) Range() token.Range { return node.TokPos.WithEnd(node.PosEnd()) }
func (node *Extern) Pos() token.Pos     { return node.TokPos }
func (node *Extern) PosEnd() token.Pos {
	if node.Args != nil {
		return node.Args.PosEnd()
	}
	const length = len("extern") - 1
	end := node.TokPos
	end.Char += uint32(length)
	end.Offset += uint64(length)
	return end
}

var (
	_ Node = (*BadNode)(nil)
	_ Node = (*Empty)(nil)
	_ Node = (*Lower)(nil)
	_ Node = (*Upper)(nil)
	_ Node = (*TypeVar)(nil)
	_ Node = (*Underscore)(nil)
	_ Node = (*Literal)(nil)

	_ Node = (*AttributeList)(nil)
	_ Node = (*LetDecl)(nil)
	_ Node = (*TypeDecl)(nil)
	_ Node = (*Decl)(nil)
	_ Node = (*Variant)(nil)

	_ Node = (*Label)(nil)
	_ Node = (*Signature)(nil)
	_ Node = (*Call)(nil)
	_ Node = (*Dot)(nil)
	_ Node = (*Op)(nil)

	_ Node = (*Stmts)(nil)
	_ Node = (*Block)(nil)
	_ Node = (*List)(nil)
	_ Node = (*Parens)(nil)

	_ Node = (*When)(nil)
	_ Node = (*Extern)(nil)
)
