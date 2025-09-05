// Package ast provides Abstract Syntax Tree types and some functions
// to work with.
package ast

import "github.com/saffage/jet/text"

type Node interface {
	Range() text.Span
	IsValid() bool

	// String representation of the node. This string must be equal to the
	// code from which this tree can be parsed.
	text.Renderer
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
		Data string    `json:"data"`
		Span text.Span `json:"span,omitzero"`
	}

	Capitalized struct {
		Data string    `json:"data"`
		Span text.Span `json:"span,omitzero"`
	}

	Placeholder struct {
		Data string    `json:"data"`
		Span text.Span `json:"span,omitzero"`
	}

	TypeVariable struct {
		Data string    `json:"name"`
		Span text.Span `json:"span,omitzero"`
	}

	Literal struct {
		Value string      `json:"value"`
		Kind  LiteralKind `json:"kind"`
		Span  text.Span   `json:"span,omitzero"`
	}
)

func (node *BadNode) Range() text.Span      { return text.Span{From: node.DesiredPos} }
func (node *Lower) Range() text.Span        { return node.Span }
func (node *Capitalized) Range() text.Span  { return node.Span }
func (node *Placeholder) Range() text.Span  { return node.Span }
func (node *TypeVariable) Range() text.Span { return node.Span }
func (node *Literal) Range() text.Span      { return node.Span }

func (node *Lower) Name() string        { return node.Data }
func (node *Capitalized) Name() string  { return node.Data }
func (node *Placeholder) Name() string  { return node.Data }
func (node *TypeVariable) Name() string { return node.Data }

//------------------------------------------------
// Declaration
//------------------------------------------------

type ValueKind byte

const (
	ValueLet ValueKind = iota
	ValueVal
	ValueVar
)

type (
	// ValueDecl represents let\val\var binding:
	//
	//	let a = 10
	//	val b = "..."
	//	var c = 0.0
	ValueDecl struct {
		Pattern    Pattern   `json:"pattern"`
		Value      Node      `json:"value"`
		KeywordPos text.Pos  `json:"keyword_pos,omitzero"`
		Kind       ValueKind `json:"kind"`
	}

	// TypeAliasDecl represents 'type Name = Type' or 'type Name(params) = Type'.
	TypeAliasDecl struct {
		Ident   *Capitalized `json:"ident"`
		Args    *Parens      `json:"args,omitzero"`
		Expr    Node         `json:"expr"`
		TypeTok text.Pos     `json:"type_tok,omitzero"`
		EqTok   text.Pos     `json:"eq_tok,omitzero"`
	}

	// TypeDecl represents 'type Name { fields and variants }' or 'type Name(params) { fields and variants }'.
	TypeDecl struct {
		Ident   *Capitalized `json:"ident"`
		Args    *Parens      `json:"args,omitzero"`
		Body    *Block       `json:"body"`
		TypeTok text.Pos     `json:"type_tok,omitzero"`
	}

	// FieldDecl represents [TypeDecl] field:
	//
	//	type T {
	//		field1        Type
	//		label: field2 Type
	//		:field3       Type // label is field3
	//	}
	FieldDecl struct {
		Label *Label `json:"label,omitzero"`
		Name  *Lower `json:"name"`
		Type  Node   `json:"type"`
	}

	// VariantDecl represents 'Name' or 'Name(T, ...)'.
	VariantDecl struct {
		Name   *Capitalized `json:"name"`
		Params *Parens      `json:"params,omitempty"` // Optional.
	}
)

func (node *ValueDecl) Range() text.Span {
	return text.Span{
		From: node.KeywordPos,
		To:   node.Value.Range().To,
	}
}

func (node *TypeAliasDecl) Range() text.Span {
	return text.Span{
		From: node.TypeTok,
		To:   node.Expr.Range().To,
	}
}

func (node *TypeDecl) Range() text.Span {
	return text.Span{
		From: node.TypeTok,
		To:   node.Body.Range().To,
	}
}

func (node *FieldDecl) Range() (span text.Span) {
	return text.Span{
		From: node.Label.Range().From,
		To:   node.Type.Range().To,
	}
}

func (node *VariantDecl) Range() (span text.Span) {
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
		X        Node     `json:"x"`
		ColonTok text.Pos `json:"colon_tok,omitzero"`
	}

	// Signature represents signature type '(...) T'.
	// It also used for 'fn(...) T' but have different semantics.
	Signature struct {
		Params *Parens `json:"params"`
		Result Node    `json:"result,omitempty"`
	}

	// FnType represents 'fn() R'
	FnType struct {
		Signature *Signature `json:"signature"`
		FnTok     text.Pos   `json:"fn_tok,omitzero"`
	}

	// Fn represents 'fn() R { expr }'
	Fn struct {
		Type *FnType `json:"fn_type,omitempty"`
		Body *Block  `json:"body,omitempty"`
	}

	// Call represents 'x(...args)'.
	Call struct {
		X    Node    `json:"x,omitempty"`
		Args *Parens `json:"args,omitempty"`
	}

	// Dot represents 'x.y'.
	Dot struct {
		X      Node     `json:"x,omitempty"`
		Y      Node     `json:"y,omitempty"`
		DotPos text.Pos `json:"dot_pos,omitzero"`
	}

	// Op represents 'x OP y, where 'OP' is an operator.
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

// Label returns a label, as it may be indirect.
//
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

	case *PatternTypeTest:
		if name, _ := labeledExpr.X.(*PatternBinding); name != nil {
			return name.Name
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

func (node *FnType) Range() text.Span {
	return text.Span{
		From: node.FnTok,
		To:   node.Signature.Range().To,
	}
}

func (node *Fn) Range() text.Span {
	return text.Span{
		From: node.Type.Range().From,
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
	// List represents '[a, b, c]'.
	List struct {
		Nodes []Node    `json:"nodes,omitempty"`
		Span  text.Span `json:"span,omitzero"`
	}

	Stmts struct {
		Items      []Node   `json:"items,omitempty"`
		DesiredPos text.Pos `json:"desired_pos,omitzero"`
	}

	// Block represents '{ a; b; c }'.
	Block struct {
		Stmts *Stmts    `json:"stmts,omitempty"`
		Span  text.Span `json:"span,omitzero"`
	}
)

func (node *List) Range() text.Span  { return node.Span }
func (node *Block) Range() text.Span { return node.Span }

func (stmts *Stmts) Range() text.Span {
	if len(stmts.Items) > 0 {
		return text.Span{
			From: stmts.Items[0].Range().From,
			To:   stmts.Items[len(stmts.Items)-1].Range().To,
		}
	}
	return text.Span{From: stmts.DesiredPos}
}

// Parens represents '(a, b, c)'.
type Parens struct {
	Nodes []Node    `json:"nodes,omitempty"`
	Span  text.Span `json:"span,omitzero"`
}

type Parens2[T Node] struct {
	Items []T       `json:"items"`
	Span  text.Span `json:"span"`
}

func (node *Parens) Range() text.Span     { return node.Span }
func (node *Parens2[T]) Range() text.Span { return node.Span }

//------------------------------------------------
// Language constructions
//------------------------------------------------

type (
	When struct {
		Cond    Node     `json:"cond,omitempty"`
		Clauses *Block   `json:"clauses,omitempty"`
		WhenTok text.Pos `json:"when_tok,omitzero"`
	}

	CaseClause struct {
		Pattern  Node     `json:"pattern,omitempty"`
		Expr     Node     `json:"expr,omitempty"`
		ArrowTok text.Pos `json:"arrow_tok,omitzero"`
	}

	External struct {
		Args        *Parens  `json:"args,omitempty"`
		ExternalTok text.Pos `json:"external_tok,omitzero"`
	}
)

func (node *When) Range() text.Span {
	return text.Span{
		From: node.WhenTok,
		To:   node.Clauses.Range().To,
	}
}

func (node *CaseClause) Range() text.Span {
	return text.Span{
		From: node.Pattern.Range().From,
		To:   node.Expr.Range().To,
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

//------------------------------------------------
// Patterns
//------------------------------------------------

type (
	Pattern interface {
		Node
		isPattern()
	}

	// PatternInvalid represents incorrectly parsed pattern.
	PatternInvalid struct {
		Span text.Span
	}

	// PatternLiteral represents literal pattern (10, "...", 1.25).
	PatternLiteral struct {
		Value *Literal
	}

	// PatternBinding represents binding pattern (name).
	PatternBinding struct {
		Name *Lower
	}

	// PatternPlaceholder represents placeholder pattern (_).
	PatternPlaceholder struct {
		Name *Placeholder
	}

	// PatternRebinding represents binding another pattern to name:
	//
	//	when [1, 2, 3] {
	//		[1, ...] as x -> /* x is List(Int) */
	//	}
	//
	// It also valid to do this:
	//
	//	a as b
	PatternRebinding struct {
		X          Pattern
		Name       *Lower
		KeywordPos text.Pos
	}

	// PatternList represents 'List('item)' type pattern, it may have any
	// number of values or ranges between values (...), example:
	//
	//	when [1, 2, 3] {
	//		[1, 2, 3]   -> {/* matches exact length and values */}
	//		[1, ...]    -> {/* matches first value and >=0 values after */}
	//		[1, ..., 3] -> {/* matches first and last values and >=0 values between */}
	//		[..., 3]    -> {/* matches last value and >=0 values before */}
	//		[...]       -> {/* matches any list */}
	//	}
	PatternList struct {
		Items    []Pattern
		Brackets text.Span
	}

	// PatternRange represents 'List('item)' range pattern.
	//
	//	when [1, 2, 3] {
	//		[...rest] -> {}	/*
	//		 ^  ^
	//		 |  name
	//		 |
	//		 range token	*/
	//	}
	PatternRange struct {
		Name     *Lower // optional
		DotsSpan text.Span
	}

	// PatternVariant represents type variant pattern:
	//
	//	when get_int() {
	//		Some(value) -> {}
	//		None        -> {}
	//	}
	PatternVariant struct {
		Name   *Capitalized
		Values []Pattern
		Parens text.Span
	}

	// PatternLabeled represents labeled value in variant pattern:
	//
	//	type T {
	//		One(a: Int)
	//		Two(b: Int)
	//	}
	//	when t() {
	//		One(:a)   -> {}
	//		Two(b: c) -> {}
	//	}
	PatternLabeled struct {
		Label    *Lower
		X        Pattern
		ColonPos text.Pos
	}

	// PatternTypeTest represents type test operation (aka type annotation):
	//
	//	let x Int = 10		/*
	//	    ^^^^^
	//	    type test pattern	*/
	//
	//	when anything {
	//		int   Int       -> {}
	//		float Float     -> {}
	//		ints  List(Int) -> {}
	//	}
	PatternTypeTest struct {
		X    Pattern
		Type Node
	}

	// PatternAlternative represents alternative patterns:
	//
	//	when int {
	//		0 | 1 | 2 -> {}
	//	}
	PatternAlternative struct {
		Items []Pattern
	}
)

func (*PatternInvalid) isPattern()     {}
func (*PatternLiteral) isPattern()     {}
func (*PatternBinding) isPattern()     {}
func (*PatternPlaceholder) isPattern() {}
func (*PatternRebinding) isPattern()   {}
func (*PatternList) isPattern()        {}
func (*PatternRange) isPattern()       {}
func (*PatternVariant) isPattern()     {}
func (*PatternLabeled) isPattern()     {}
func (*PatternTypeTest) isPattern()    {}
func (*PatternAlternative) isPattern() {}

func (*PatternInvalid) Range() text.Span {
	return text.Span{}
}

func (*PatternLiteral) Range() text.Span {
	return text.Span{}
}

func (*PatternBinding) Range() text.Span {
	return text.Span{}
}

func (*PatternPlaceholder) Range() text.Span {
	return text.Span{}
}

func (*PatternRebinding) Range() text.Span {
	return text.Span{}
}

func (*PatternList) Range() text.Span {
	return text.Span{}
}

func (*PatternRange) Range() text.Span {
	return text.Span{}
}

func (*PatternVariant) Range() text.Span {
	return text.Span{}
}

func (*PatternLabeled) Range() text.Span {
	return text.Span{}
}

func (*PatternTypeTest) Range() text.Span {
	return text.Span{}
}

func (*PatternAlternative) Range() text.Span {
	return text.Span{}
}

var (
	_ Ident = (*Lower)(nil)
	_ Ident = (*Capitalized)(nil)
	_ Ident = (*Placeholder)(nil)
	_ Ident = (*TypeVariable)(nil)

	_ Node = (*BadNode)(nil)
	_ Node = (*Literal)(nil)

	_ Node = (*ValueDecl)(nil)
	_ Node = (*TypeAliasDecl)(nil)
	_ Node = (*TypeDecl)(nil)
	_ Node = (*FieldDecl)(nil)
	_ Node = (*VariantDecl)(nil)

	_ Node = (*Label)(nil)
	_ Node = (*Signature)(nil)
	_ Node = (*Fn)(nil)
	_ Node = (*Call)(nil)
	_ Node = (*Dot)(nil)
	_ Node = (*Op)(nil)

	_ Node = (*List)(nil)
	_ Node = (*Stmts)(nil)
	_ Node = (*Block)(nil)
	// _ Node = (*Parens)(nil)

	_ Node = (*When)(nil)
	_ Node = (*CaseClause)(nil)
	_ Node = (*External)(nil)

	_ Pattern = (*PatternLiteral)(nil)
	_ Pattern = (*PatternBinding)(nil)
	_ Pattern = (*PatternPlaceholder)(nil)
	_ Pattern = (*PatternRebinding)(nil)
	_ Pattern = (*PatternList)(nil)
	_ Pattern = (*PatternRange)(nil)
	_ Pattern = (*PatternVariant)(nil)
	_ Pattern = (*PatternLabeled)(nil)
	_ Pattern = (*PatternTypeTest)(nil)
	_ Pattern = (*PatternAlternative)(nil)
)
