package ast

import (
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

type Node interface {
	Range() text.Span
	// Valid() bool
	report.Renderer
}

type Ident interface {
	Node
	Name() string
}

//------------------------------------------------
// Atoms
//------------------------------------------------

type (
	BadNode struct {
		DesiredPos text.Pos
	}

	Lower struct {
		Data string
		Span text.Span
	}

	Upper struct {
		Data string
		Span text.Span
	}

	Placeholder struct {
		Data string
		Span text.Span
	}

	Literal struct {
		Value string
		Kind  LiteralKind
		Span  text.Span
	}
)

func (node *BadNode) Range() text.Span     { return text.Span{From: node.DesiredPos} }
func (node *Lower) Range() text.Span       { return node.Span }
func (node *Upper) Range() text.Span       { return node.Span }
func (node *Placeholder) Range() text.Span { return node.Span }
func (node *Literal) Range() text.Span     { return node.Span }

func (node *Lower) Name() string       { return node.Data }
func (node *Upper) Name() string       { return node.Data }
func (node *Placeholder) Name() string { return node.Data }

//------------------------------------------------
// Declaration
//------------------------------------------------

type (
	// CommentGroup struct {
	// 	Comments []struct {
	// 		Value string
	// 		Span  text.Span
	// 	}
	// }

	// Represents 'let name Type = value'.
	LetDecl struct {
		LetTok text.Pos
		Decl   *Decl
		Value  Node
	}

	// Represents 'type Name = Type' or 'type Name(params) = Type'.
	TypeAlias struct {
		Ident   *Upper
		Args    *Parens `yaml:",omitempty"`
		Expr    Node
		TypeTok text.Pos
		EqTok   text.Pos
	}

	// Represents 'type Name { fields and variants }' or 'type Name(params) { fields and variants }'.
	TypeDef struct {
		Ident   *Upper
		Args    *Parens `yaml:",omitempty"`
		Body    *Block
		TypeTok text.Pos
	}

	// Represents `name T`, `name`, `type name T`, `type name`.
	Decl struct {
		Ident   Ident    ``
		Type    Node     `yaml:",omitempty"`
		TypeTok text.Pos `yaml:",omitempty"`
	}

	// Represents 'Name' or 'Name(T)'.
	Variant struct {
		Name   *Upper
		Params *Parens `yaml:",omitempty"` // Optional.
	}
)

// func (node *CommentGroup) Range() text.Span {
// 	if len(node.Comments) > 0 {
// 		return text.Span{
// 			From: node.Comments[0].Span.From,
// 			To:   node.Comments[len(node.Comments)-1].Span.To,
// 		}
// 	}

// 	return text.Span{}
// }

func (node *LetDecl) Range() text.Span {
	return text.Span{
		From: node.LetTok,
		To:   node.Value.Range().To,
	}
}

func (node *TypeAlias) Range() text.Span {
	return text.Span{
		From: node.TypeTok,
		To:   node.Expr.Range().To,
	}
}

func (node *TypeDef) Range() text.Span {
	return text.Span{
		From: node.TypeTok,
		To:   node.Body.Range().To,
	}
}

func (node *Decl) Range() (span text.Span) {
	if node.TypeTok.IsValid() {
		span.From = node.TypeTok
	} else {
		span.From = node.Ident.Range().From
	}

	if node.Type != nil {
		span.To = node.Type.Range().To
	} else {
		span.To = node.Ident.Range().To
	}

	return
}

func (node *Variant) Range() (span text.Span) {
	span.From = node.Name.Range().From

	if node.Params != nil {
		span.To = node.Params.Range().To
	} else {
		span.To = node.Name.Span.To
	}

	return
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
		Params *Parens
		Result Node `yaml:",omitempty"` // can be nil in some cases
	}

	// Represents 'fn() R = expr'
	Function struct {
		Signature *Signature
		Body      Node
		FnTok     text.Pos
		EqTok     text.Pos
	}

	// Represents 'x(...args)'.
	Call struct {
		X    Node
		Args *Parens
	}

	// Represents 'x.y'.
	Dot struct {
		X      Node
		Y      Node
		DotPos text.Pos
	}

	// Represents 'x OP y, where 'OP' is an operator.
	Op struct {
		X    Node `yaml:",omitempty"`
		Y    Node `yaml:",omitempty"`
		Span text.Span
		Kind OperatorKind
	}
)

func (node *Label) Range() text.Span {
	return text.Span{
		From: node.Name.Range().From,
		To:   node.X.Range().To,
	}
}

func (node *Label) Label() *Lower {
	if node.Name != nil {
		return node.Name
	}

	switch x := node.X.(type) {
	case *Lower:
		return x

	case *Decl:
		if name, _ := x.Ident.(*Lower); name != nil {
			return name
		}
	}

	return nil
}

func (node *Signature) Range() (span text.Span) {
	span.From = node.Params.Range().From

	if node.Result != nil {
		span.To = node.Result.Range().To
	} else {
		span.To = node.Params.Range().To
	}

	return
}

func (node *Function) Range() text.Span {
	return text.Span{
		From: node.FnTok,
		To:   node.Body.Range().To,
	}
}

func (node *Call) Range() text.Span {
	return text.Span{
		From: node.X.Range().From,
		To:   node.Args.Range().To,
	}
}

func (node *Dot) Range() text.Span {
	return text.Span{
		From: node.X.Range().From,
		To:   node.Y.Range().To,
	}
}

func (node *Op) Range() (span text.Span) {
	if node.X != nil {
		span.From = node.X.Range().From
	} else {
		span.From = node.Span.From
	}

	if node.Y != nil {
		span.To = node.Y.Range().To
	} else {
		span.To = node.Span.To
	}

	return
}

func (node *Op) IsInfix() bool   { return node.X != nil && node.Y != nil }
func (node *Op) IsPrefix() bool  { return node.X != nil && node.Y == nil }
func (node *Op) IsPostfix() bool { return node.X == nil && node.Y != nil }
func (node *Op) IsName() bool    { return node.X == nil && node.Y == nil }

//------------------------------------------------
// Lists
//------------------------------------------------

type (
	// Represents '[a, b, c]'.
	List struct {
		Nodes []Node
		Span  text.Span
	}

	Stmts struct {
		Items      []Node
		DesiredPos text.Pos
	}

	// Represents '{ a; b; c }'.
	Block struct {
		Stmts *Stmts
		Span  text.Span
	}

	// Represents '(a, b, c)'.
	Parens struct {
		Nodes []Node
		Span  text.Span
	}
)

func (node *List) Range() text.Span   { return node.Span }
func (node *Block) Range() text.Span  { return node.Span }
func (node *Parens) Range() text.Span { return node.Span }

func (stmts *Stmts) Range() text.Span {
	if len(stmts.Items) > 0 {
		return text.Span{
			From: stmts.Items[0].Range().From,
			To:   stmts.Items[len(stmts.Items)-1].Range().To,
		}
	}
	return text.Span{From: stmts.DesiredPos}
}

//------------------------------------------------
// Language constructions
//------------------------------------------------

type (
	When struct {
		Expr    Node `yaml:",omitempty"`
		Body    *Block
		WhenTok text.Pos
	}

	Case struct {
		Pattern  Node
		Expr     Node
		ArrowTok text.Pos
	}

	Spread struct {
		Expr      Node `yaml:",omitempty"`
		SpreadTok text.Pos
	}

	As struct {
		Lhs   Node
		Rhs   Node
		AsTok text.Pos
	}

	Extern struct {
		Args      *Parens `yaml:",omitempty"`
		ExternTok text.Pos
	}
)

func (node *When) Range() text.Span {
	return text.Span{
		From: node.WhenTok,
		To:   node.Body.Range().To,
	}
}

func (node *Case) Range() text.Span {
	return text.Span{
		From: node.Pattern.Range().From,
		To:   node.Expr.Range().To,
	}
}

func (node *Spread) Range() text.Span {
	return text.Span{
		From: node.SpreadTok,
		To:   node.Expr.Range().To,
	}
}

func (node *As) Range() text.Span {
	return text.Span{
		From: node.Lhs.Range().From,
		To:   node.Rhs.Range().To,
	}
}

func (node *Extern) Range() text.Span {
	span := text.Span{From: node.ExternTok}

	if node.Args != nil {
		span.To = node.Args.Range().To
	} else {
		span.To = node.ExternTok.WithOffset(len("extern") - 1)
	}

	return span
}

var (
	_ Node = (*BadNode)(nil)
	_ Node = (*Lower)(nil)
	_ Node = (*Upper)(nil)
	_ Node = (*Placeholder)(nil)
	_ Node = (*Literal)(nil)

	_ Node = (*LetDecl)(nil)
	_ Node = (*TypeAlias)(nil)
	_ Node = (*TypeDef)(nil)
	_ Node = (*Decl)(nil)
	_ Node = (*Variant)(nil)

	_ Node = (*Label)(nil)
	_ Node = (*Signature)(nil)
	_ Node = (*Function)(nil)
	_ Node = (*Call)(nil)
	_ Node = (*Dot)(nil)
	_ Node = (*Op)(nil)

	_ Node = (*List)(nil)
	_ Node = (*Stmts)(nil)
	_ Node = (*Block)(nil)
	_ Node = (*Parens)(nil)

	_ Node = (*When)(nil)
	_ Node = (*Case)(nil)
	_ Node = (*Extern)(nil)
)

var (
	_ Ident = (*Lower)(nil)
	_ Ident = (*Upper)(nil)
	_ Ident = (*Placeholder)(nil)
)
