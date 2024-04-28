package ast

import (
	"github.com/saffage/jet/token"
)

type (
	Node interface {
		Pos() token.Loc
		PosEnd() token.Loc
		String() string
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

	Comment struct {
		Data       string
		Start, End token.Loc
	}

	CommentGroup struct {
		Comments []*Comment
	}
)

func (n *BadNode) Pos() token.Loc    { return n.Loc }
func (n *BadNode) PosEnd() token.Loc { return n.Loc }

func (n *Empty) Pos() token.Loc    { return n.Loc }
func (n *Empty) PosEnd() token.Loc { return n.Loc }

func (n *Ident) Pos() token.Loc    { return n.Start }
func (n *Ident) PosEnd() token.Loc { return n.End }

func (n *Literal) Pos() token.Loc    { return n.Start }
func (n *Literal) PosEnd() token.Loc { return n.End }

func (n *Comment) Pos() token.Loc    { return n.Start }
func (n *Comment) PosEnd() token.Loc { return n.End }

func (n *CommentGroup) Pos() token.Loc    { return n.Comments[0].Pos() }
func (n *CommentGroup) PosEnd() token.Loc { return n.Comments[len(n.Comments)-1].PosEnd() }

type (
	// Represents a `...` token in AST.
	// Used in parameter lists.
	Ellipsis struct {
		X   Node      // Not nil if used with expression.
		Loc token.Loc // `...` token.
	}

	// Represents `[]x` or `[N]x`.
	ArrayType struct {
		N, X        Node
		Open, Close token.Loc
	}

	// Represents function signature.
	Signature struct {
		Params *ParenList
		Result Node
		Loc    token.Loc // `func` token.
	}

	// Represents `@Foo` or `@Foo(args)`.
	Annotation struct {
		Name *Ident
		Args *ParenList
		Loc  token.Loc
	}

	// Represents `#foo`, `#foo expr`, `#foo()`, `#foo{}`.
	Attribute struct {
		Name *Ident
		X    Node      // Next statement\expression, 'ast.ParenList' or 'ast.CurlyList'.
		Loc  token.Loc // `#` token.
	}

	// Represents `?` suffix.
	Try struct {
		X   Node
		Loc token.Loc // `?` token.
	}

	// Represents `!` suffix.
	Unwrap struct {
		X   Node
		Loc token.Loc // `!` token.
	}

	// Represents `x.selector`.
	MemberAccess struct {
		X        Node
		Selector Node
		Loc      token.Loc // `.` token.
	}

	// Used in `MemberAccess` nodes to represent `.*` suffix.
	Star struct {
		Loc token.Loc // `*` token.
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

	// Represents `a, b Type = value`.
	Field struct {
		Names []*Ident
		Type  Node // maybe nil.
		Value Node // maybe nil.
	}

	UnaryOpr struct {
		Start token.Loc
		End   token.Loc
		Kind  UnaryOpKind
	}

	BinaryOpr struct {
		Start token.Loc
		End   token.Loc
		Kind  BinaryOpKind
	}

	UnaryOp struct {
		X   Node
		Opr *UnaryOpr
	}

	BinaryOp struct {
		X, Y Node
		Opr  *BinaryOpr
	}
)

func (n *Ellipsis) Pos() token.Loc    { return n.Loc }
func (n *Ellipsis) PosEnd() token.Loc { return n.X.PosEnd() }

func (n *Annotation) Pos() token.Loc { return n.Loc }
func (n *Annotation) PosEnd() token.Loc {
	if n.Args != nil {
		return n.Args.PosEnd()
	}
	return n.Name.PosEnd()
}

func (n *Attribute) Pos() token.Loc { return n.Loc }
func (n *Attribute) PosEnd() token.Loc {
	if n.X != nil {
		return n.X.PosEnd()
	}
	return n.Name.PosEnd()
}

func (n *Try) Pos() token.Loc    { return n.X.Pos() }
func (n *Try) PosEnd() token.Loc { return n.Loc }

func (n *Unwrap) Pos() token.Loc    { return n.X.Pos() }
func (n *Unwrap) PosEnd() token.Loc { return n.Loc }

func (n *ArrayType) Pos() token.Loc    { return n.Open }
func (n *ArrayType) PosEnd() token.Loc { return n.X.PosEnd() }

func (n *Signature) Pos() token.Loc {
	if n.Loc.Line == 0 {
		return n.Params.Pos()
	}
	return n.Loc
}
func (n *Signature) PosEnd() token.Loc { return n.Result.PosEnd() }

func (n *MemberAccess) Pos() token.Loc    { return n.X.Pos() }
func (n *MemberAccess) PosEnd() token.Loc { return n.Selector.PosEnd() }

func (n *Star) Pos() token.Loc    { return n.Loc }
func (n *Star) PosEnd() token.Loc { return n.Loc }

func (n *Call) Pos() token.Loc    { return n.X.Pos() }
func (n *Call) PosEnd() token.Loc { return n.Args.PosEnd() }

func (n *Index) Pos() token.Loc    { return n.X.Pos() }
func (n *Index) PosEnd() token.Loc { return n.Args.PosEnd() }

func (n *List) Pos() token.Loc    { return n.Nodes[0].Pos() }
func (n *List) PosEnd() token.Loc { return n.Nodes[len(n.Nodes)-1].PosEnd() }

func (n *ParenList) Pos() token.Loc    { return n.Open }
func (n *ParenList) PosEnd() token.Loc { return n.Close }

func (n *CurlyList) Pos() token.Loc    { return n.Open }
func (n *CurlyList) PosEnd() token.Loc { return n.Close }

func (n *BracketList) Pos() token.Loc    { return n.Open }
func (n *BracketList) PosEnd() token.Loc { return n.Close }

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

func (n *UnaryOpr) Pos() token.Loc    { return n.Start }
func (n *UnaryOpr) PosEnd() token.Loc { return n.End }

func (n *BinaryOpr) Pos() token.Loc    { return n.Start }
func (n *BinaryOpr) PosEnd() token.Loc { return n.End }

func (n *UnaryOp) Pos() token.Loc    { return n.Opr.Pos() }
func (n *UnaryOp) PosEnd() token.Loc { return n.X.PosEnd() }

func (n *BinaryOp) Pos() token.Loc    { return n.X.Pos() }
func (n *BinaryOp) PosEnd() token.Loc { return n.Y.PosEnd() }

type (
	Decl interface {
		Node
		Ident() *Ident
		Doc() string
		Annotations() []*Annotation
	}

	ModuleDecl struct {
		Annots []*Annotation
		Name   *Ident
		Body   Node      // Can be either [CurlyList] or [ExprList]
		Loc    token.Loc // `module` token
	}

	// Declaration of variables and constants
	GenericDecl struct {
		Annots []*Annotation
		Field  *Field
		Loc    token.Loc // `const`, `var`, `val` token
		Kind   GenericDeclKind
	}

	FuncDecl struct {
		Annots    []*Annotation
		Name      *Ident
		Signature *Signature
		Body      Node
		Loc       token.Loc // `func` token
	}

	StructDecl struct {
		Annots []*Annotation
		Name   *Ident
		Fields *CurlyList
		Loc    token.Loc // `struct` token
	}

	EnumDecl struct {
		Annots []*Annotation
		Name   *Ident
		Body   *CurlyList
		Loc    token.Loc // `enum` token
	}

	TypeAliasDecl struct {
		Annots []*Annotation
		Name   *Ident
		Expr   Node
		Loc    token.Loc // `alias` token
	}
)

func (n *ModuleDecl) Pos() token.Loc             { return n.Loc }
func (n *ModuleDecl) PosEnd() token.Loc          { return n.Name.PosEnd() }
func (n *ModuleDecl) Ident() *Ident              { return n.Name }
func (*ModuleDecl) Doc() string                  { return "" }
func (n *ModuleDecl) Annotations() []*Annotation { return n.Annots }

func (n *GenericDecl) Pos() token.Loc    { return n.Loc }
func (n *GenericDecl) PosEnd() token.Loc { return n.Field.PosEnd() }
func (n *GenericDecl) Ident() *Ident {
	if len(n.Field.Names) > 0 {
		return n.Field.Names[0]
	}
	return nil
}
func (*GenericDecl) Doc() string                  { return "" }
func (n *GenericDecl) Annotations() []*Annotation { return n.Annots }

func (n *FuncDecl) Pos() token.Loc { return n.Loc }
func (n *FuncDecl) PosEnd() token.Loc {
	if n.Body != nil {
		return n.Body.PosEnd()
	}
	return n.Signature.PosEnd()
}
func (n *FuncDecl) Ident() *Ident              { return n.Name }
func (*FuncDecl) Doc() string                  { return "" }
func (n *FuncDecl) Annotations() []*Annotation { return n.Annots }

func (n *StructDecl) Pos() token.Loc             { return n.Loc }
func (n *StructDecl) PosEnd() token.Loc          { return n.Fields.PosEnd() }
func (n *StructDecl) Ident() *Ident              { return n.Name }
func (*StructDecl) Doc() string                  { return "" }
func (n *StructDecl) Annotations() []*Annotation { return n.Annots }

func (n *EnumDecl) Pos() token.Loc             { return n.Loc }
func (n *EnumDecl) PosEnd() token.Loc          { return n.Body.PosEnd() }
func (n *EnumDecl) Ident() *Ident              { return n.Name }
func (*EnumDecl) Doc() string                  { return "" }
func (n *EnumDecl) Annotations() []*Annotation { return n.Annots }

func (n *TypeAliasDecl) Pos() token.Loc             { return n.Loc }
func (n *TypeAliasDecl) PosEnd() token.Loc          { return n.Expr.PosEnd() }
func (n *TypeAliasDecl) Ident() *Ident              { return n.Name }
func (*TypeAliasDecl) Doc() string                  { return "" }
func (n *TypeAliasDecl) Annotations() []*Annotation { return n.Annots }

type (
	If struct {
		Cond Node
		Body Node
		Else *Else
		Loc  token.Loc // `if` token
	}

	Else struct {
		Body Node
		Loc  token.Loc // `else` token
	}

	While struct {
		Cond Node
		Body Node
		Loc  token.Loc // `while` token
	}

	Return struct {
		X   Node
		Loc token.Loc
	}

	Break struct {
		Label *Ident
		Loc   token.Loc
	}

	Continue struct {
		Label *Ident
		Loc   token.Loc
	}
)

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

func (n *Break) Pos() token.Loc    { return n.Loc }
func (n *Break) PosEnd() token.Loc { return n.Label.PosEnd() }

func (n *Continue) Pos() token.Loc    { return n.Loc }
func (n *Continue) PosEnd() token.Loc { return n.Label.PosEnd() }
