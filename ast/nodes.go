package ast

import (
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

type Node interface {
	Range() text.Span

	// String representation of the node. This string must be equal to the
	// code from which this tree can be parsed.
	report.Renderer
}

type Ident interface {
	Node
	Name() string
}

type SomeNode interface {
	comparable
	Node
}

type SomeIdent interface {
	comparable
	Ident
}

//------------------------------------------------
// Atoms
//------------------------------------------------

type (
	BadNode struct {
		DesiredPos text.Pos `json:"desired_pos,omitzero"`
	}

	Lower struct {
		Data string    `json:"data,omitempty"`
		Span text.Span `json:"span,omitzero"`
	}

	Upper struct {
		Data string    `json:"data,omitempty"`
		Span text.Span `json:"span,omitzero"`
	}

	Placeholder struct {
		Data string    `json:"data,omitempty"`
		Span text.Span `json:"span,omitzero"`
	}

	Literal struct {
		Value string      `json:"value,omitempty"`
		Kind  LiteralKind `json:"kind,omitempty"`
		Span  text.Span   `json:"span,omitzero"`
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
		Decl   *Decl    `json:"decl,omitempty"`
		Value  Node     `json:"value,omitempty"`
		LetTok text.Pos `json:"let_tok,omitzero"`
	}

	// Represents 'val name Type = value'.
	ValDecl struct {
		Decl   *Decl    `json:"decl,omitempty"`
		Value  Node     `json:"value,omitempty"`
		ValTok text.Pos `json:"val_tok,omitzero"`
	}

	// Represents 'var name Type = value'.
	VarDecl struct {
		Decl   *Decl    `json:"decl,omitempty"`
		Value  Node     `json:"value,omitempty"`
		VarTok text.Pos `json:"var_tok,omitzero"`
	}

	// Represents 'type Name = Type' or 'type Name(params) = Type'.
	TypeAlias struct {
		Ident   *Upper   `json:"ident,omitempty"`
		Args    *Parens  `json:"args,omitempty"`
		Expr    Node     `json:"expr,omitempty"`
		TypeTok text.Pos `json:"type_tok,omitzero"`
		EqTok   text.Pos `json:"eq_tok,omitzero"`
	}

	// Represents 'type Name { fields and variants }' or 'type Name(params) { fields and variants }'.
	TypeDef struct {
		Ident   *Upper   `json:"ident,omitempty"`
		Args    *Parens  `json:"args,omitempty"`
		Body    *Block   `json:"body,omitempty"`
		TypeTok text.Pos `json:"type_tok,omitzero"`
	}

	// Represents `name T`, `name`, `type name T`, `type name`.
	Decl struct {
		Ident   Ident    `json:"ident,omitempty"`
		Type    Node     `json:"type,omitempty"`
		TypeTok text.Pos `json:"type_tok,omitzero"`
	}

	// Represents 'Name' or 'Name(T)'.
	Variant struct {
		Name   *Upper  `json:"name,omitempty"`
		Params *Parens `json:"params,omitempty"` // Optional.
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

func (node *ValDecl) Range() text.Span {
	return text.Span{
		From: node.ValTok,
		To:   node.Value.Range().To,
	}
}

func (node *VarDecl) Range() text.Span {
	return text.Span{
		From: node.VarTok,
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
		Name     *Lower   `json:"name,omitempty"`
		X        Node     `json:"x,omitempty"`
		ColonTok text.Pos `json:"colon_tok,omitzero"`
	}

	// Represents '() T with Effects'.
	Signature struct {
		Params *Parens `json:"params,omitempty"`
		Result Node    `json:"result,omitempty"` // can be nil in some cases
	}

	// Represents 'fn() R = expr'
	Function struct {
		Signature *Signature `json:"signature,omitempty"`
		Body      Node       `json:"body,omitempty"`
		FnTok     text.Pos   `json:"fn_tok,omitzero"`
		EqTok     text.Pos   `json:"eq_tok,omitzero"`
	}

	// Represents 'x(...args)'.
	Call struct {
		X    Node    `json:"x,omitempty"`
		Args *Parens `json:"args,omitempty"`
	}

	// Represents 'x.y'.
	Dot struct {
		X      Node     `json:"x,omitempty"`
		Y      Node     `json:"y,omitempty"`
		DotPos text.Pos `json:"dot_pos,omitzero"`
	}

	// Represents 'x OP y, where 'OP' is an operator.
	Op struct {
		X    Node         `json:"x,omitempty"`
		Y    Node         `json:"y,omitempty"`
		Span text.Span    `json:"span,omitzero"`
		Kind OperatorKind `json:"kind,omitzero"`
	}
)

func (node *Label) Range() text.Span {
	return text.Span{
		From: node.Name.Range().From,
		To:   node.X.Range().To,
	}
}

// # Short labels
//
// Short labels are syntactic sugar for specifying a label with the same name as
// in the expression. For example this…
//
//	label: label
//
// … can be written as this…
//
//	:label
//
// There are only a few cases when a short label is used.
//
// 1. lowercase identifier. Basically, any value.
//
//	:foo
//
// 2. Arbitrarily nested member access. Takes the last part of the expression as
// a label.
//
//	:foo.bar // label is 'bar'
//	:foo().bar // label will be 'bar' regardless of the operand.
//
// 3. Declaration of a value. This may be a field, a parameter, or a property.
//
//	:foo T
//
// In any other case, the label must be specified as a separate lowercase
// identifier before the colon.
//
//	label: ...
func (node *Label) Label() *Lower {
	if node.Name != nil {
		return node.Name
	}

	switch labeledExpr := node.X.(type) {
	case *Lower:
		return labeledExpr

	case *Decl:
		if name, _ := labeledExpr.Ident.(*Lower); name != nil {
			return name
		}

	case *Dot:
		y := labeledExpr.Y
		for {
			switch member := y.(type) {
			case *Lower:
				return member

			case *Dot:
				y = member.Y

			default:
				return nil
			}
		}
	}

	return nil
}

func (node *Label) IsShort() bool {
	return node.Name == nil
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
		Nodes []Node    `json:"nodes,omitempty"`
		Span  text.Span `json:"span,omitzero"`
	}

	Stmts struct {
		Items      []Node   `json:"items,omitempty"`
		DesiredPos text.Pos `json:"desired_pos,omitzero"`
	}

	// Represents '{ a; b; c }'.
	Block struct {
		Stmts *Stmts    `json:"stmts,omitempty"`
		Span  text.Span `json:"span,omitzero"`
	}

	// Represents '(a, b, c)'.
	Parens struct {
		Nodes []Node    `json:"nodes,omitempty"`
		Span  text.Span `json:"span,omitzero"`
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
		Expr    Node     `json:"expr,omitempty"`
		Body    *Block   `json:"body,omitempty"`
		WhenTok text.Pos `json:"when_tok,omitzero"`
	}

	Case struct {
		Pattern  Node     `json:"pattern,omitempty"`
		Expr     Node     `json:"expr,omitempty"`
		ArrowTok text.Pos `json:"arrow_tok,omitzero"`
	}

	Spread struct {
		Expr      Node     `json:"expr,omitempty"`
		SpreadTok text.Pos `json:"spread_tok,omitzero"`
	}

	As struct {
		Expr    Node     `json:"expr,omitempty"`
		NewName Ident    `json:"new_name,omitempty"`
		AsTok   text.Pos `json:"as_tok,omitzero"`
	}

	External struct {
		Args        *Parens  `json:"args,omitempty"`
		ExternalTok text.Pos `json:"external_tok,omitzero"`
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
		From: node.Expr.Range().From,
		To:   node.NewName.Range().To,
	}
}

func (node *External) Range() text.Span {
	span := text.Span{From: node.ExternalTok}

	if node.Args != nil {
		span.To = node.Args.Range().To
	} else {
		span.To = node.ExternalTok.WithOffset(len("external") - 1)
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
	_ Node = (*ValDecl)(nil)
	_ Node = (*VarDecl)(nil)
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
	_ Node = (*External)(nil)
)

var (
	_ Ident = (*Lower)(nil)
	_ Ident = (*Upper)(nil)
	_ Ident = (*Placeholder)(nil)
)
