package ast

import (
	"strconv"
	"strings"

	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

func Render(node Node) string {
	const initialBufferSize = 256

	buf := strings.Builder{}
	buf.Grow(initialBufferSize)
	render(node, &buf)

	return buf.String()
}

func render[T SomeNode](node T, buf text.Writer) {
	var zero T

	if node == zero {
		buf.WriteString("#[nil-ast-node]#")
	} else if !node.IsValid() {
		buf.WriteString("#[ill-formed-ast]#")
	} else {
		node.Render(buf)
	}
}

func (node *BadNode) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("#[bad-ast-node]#")
	return true
}

func (node *Lower) Render(buf text.Writer) (rendered bool) {
	buf.WriteString(node.Data)
	return true
}

func (node *Upper) Render(buf text.Writer) (rendered bool) {
	buf.WriteString(node.Data)
	return true
}

func (node *Placeholder) Render(buf text.Writer) (rendered bool) {
	buf.WriteString(node.Data)
	return true
}

func (node *Literal) Render(buf text.Writer) (rendered bool) {
	switch node.Kind {
	case IntLiteral, FloatLiteral:
		buf.WriteString(node.Value)

	case StringLiteral:
		// TODO replace [strconv.Quote] with own implementation.
		buf.WriteString(strconv.Quote(node.Value[1 : len(node.Value)-1]))

	default:
		panic("unreachable")
	}

	return true
}

func (node *LetDecl) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("let ")

	render(node.Decl, buf)

	buf.WriteString(" = ")

	render(node.Value, buf)
	return true
}

func (node *ValDecl) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("val ")

	render(node.Decl, buf)

	buf.WriteString(" = ")

	render(node.Value, buf)
	return true
}

func (node *VarDecl) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("var ")

	render(node.Decl, buf)

	buf.WriteString(" = ")

	render(node.Value, buf)
	return true
}

func (node *TypeAlias) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("type ")

	render(node.Ident, buf)

	if node.Args != nil {
		render(node.Args, buf)
	}

	buf.WriteString(" = ")

	render(node.Expr, buf)
	return true
}

func (node *TypeDef) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("type ")

	render(node.Ident, buf)

	if node.Args != nil {
		render(node.Args, buf)
	}

	buf.WriteByte(' ')

	render(node.Body, buf)
	return true
}

func (node *Decl) Render(buf text.Writer) (rendered bool) {
	if node.TypeTok.IsValid() {
		buf.WriteString("type ")
	}

	render(node.Ident, buf)

	if node.Type != nil {
		buf.WriteByte(' ')

		render(node.Type, buf)
	}

	return true
}

func (node *Variant) Render(buf text.Writer) (rendered bool) {
	render(node.Name, buf)

	if node.Params != nil {
		render(node.Params, buf)
	}

	return true
}

func (node *Label) Render(buf text.Writer) (rendered bool) {
	if node.Name != nil {
		render(node.Name, buf)

		buf.WriteString(": ")
	} else {
		buf.WriteByte(':')
	}

	render(node.X, buf)
	return true
}

func (node *Signature) Render(buf text.Writer) (rendered bool) {
	render(node.Params, buf)

	if node.Result != nil {
		buf.WriteByte(' ')

		render(node.Result, buf)
	}

	return true
}

func (node *Function) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("fn")

	render(node.Signature, buf)

	if node.Body != nil {
		buf.WriteString(" = ")

		render(node.Body, buf)
	}

	return true
}

func (node *Call) Render(buf text.Writer) (rendered bool) {
	render(node.X, buf)
	render(node.Args, buf)
	return true
}

func (node *Dot) Render(buf text.Writer) (rendered bool) {
	render(node.X, buf)

	buf.WriteByte('.')

	render(node.Y, buf)
	return true
}

func (node *Op) Render(buf text.Writer) (rendered bool) {
	if node.X != nil {
		render(node.X, buf)
	}

	buf.WriteByte(' ')
	buf.WriteString(node.Kind.String())
	buf.WriteByte(' ')

	if node.Y != nil {
		render(node.Y, buf)
	}

	return true
}

func (node *List) Render(buf text.Writer) (rendered bool) {
	buf.WriteByte('[')

	for i, node := range node.Nodes {
		if i > 0 {
			buf.WriteString(", ")
		}

		render(node, buf)
	}

	buf.WriteByte(']')
	return true
}

func (stmts Stmts) Render(buf text.Writer) (rendered bool) {
	for i, stmt := range stmts.Items {
		if i > 0 {
			buf.WriteString("; ")
		}

		render(stmt, buf)
	}

	return true
}

func (node *Block) Render(buf text.Writer) (rendered bool) {
	if len(node.Stmts.Items) == 0 {
		buf.WriteString("{}")
	} else {
		buf.WriteString("{ ")

		render(node.Stmts, buf)

		buf.WriteString(" }")
	}

	return true
}

func (node *Parens) Render(buf text.Writer) (rendered bool) {
	buf.WriteByte('(')

	for i, node := range node.Nodes {
		if i > 0 {
			buf.WriteString(", ")
		}

		render(node, buf)
	}

	buf.WriteByte(')')
	return true
}

func (node *When) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("when ")

	render(node.Expr, buf)

	buf.WriteByte(' ')

	render(node.Body, buf)
	return true
}

func (node *Case) Render(buf text.Writer) (rendered bool) {
	render(node.Pattern, buf)

	buf.WriteString(" -> ")

	render(node.Expr, buf)
	return true
}

func (node *Spread) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("..")

	if node.Expr != nil {
		render(node.Expr, buf)
	}
	return true
}

func (node *As) Render(buf text.Writer) (rendered bool) {
	render(node.Expr, buf)

	buf.WriteString(" as ")

	render(node.NewName, buf)
	return true
}

func (node *External) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("external")

	if node.Args != nil {
		render(node.Args, buf)
	}
	return true
}

//
//
//

func (node *BadNode) IsValid() bool {
	return node != nil
}

func (node *Lower) IsValid() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Upper) IsValid() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Placeholder) IsValid() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Literal) IsValid() bool {
	return node != nil && node.Value != "" && // TODO: check the value
		(IntLiteral <= node.Kind && node.Kind <= StringLiteral)
}

func (node *LetDecl) IsValid() bool {
	return node != nil && node.Decl != nil && node.Value != nil
}

func (node *ValDecl) IsValid() bool {
	return node != nil && node.Decl != nil && node.Value != nil
}

func (node *VarDecl) IsValid() bool {
	return node != nil && node.Decl != nil && node.Value != nil
}

func (node *TypeAlias) IsValid() bool {
	return node != nil && node.Ident != nil && node.Expr != nil
}

func (node *TypeDef) IsValid() bool {
	return node != nil && node.Ident != nil && node.Body != nil
}

func (node *Decl) IsValid() bool {
	return node != nil && node.Ident != nil
}

func (node *Variant) IsValid() bool {
	return node != nil && node.Name != nil
}

func (node *Label) IsValid() bool {
	return node != nil && node.X != nil
}

func (node *Signature) IsValid() bool {
	return node != nil && node.Params != nil
}

func (node *Function) IsValid() bool {
	return node != nil && node.Signature != nil && node.Body != nil
}

func (node *Call) IsValid() bool {
	return node != nil && node.X != nil && node.Args != nil
}

func (node *Dot) IsValid() bool {
	return node != nil && node.X != nil && node.Y != nil
}

func (node *Op) IsValid() bool {
	return node != nil &&
		(OperatorNot <= node.Kind && node.Kind <= OperatorBitOr)
}

func (stmts Stmts) IsValid() bool {
	return true
}

func (node *Block) IsValid() bool {
	return node != nil && node.Stmts != nil
}

func (node *Parens) IsValid() bool {
	return node != nil
}

func (node *List) IsValid() bool {
	return node != nil
}

func (node *When) IsValid() bool {
	return node != nil && node.Expr != nil && node.Body != nil
}

func (node *Case) IsValid() bool {
	return node != nil && node.Pattern != nil && node.Expr != nil
}

func (node *Spread) IsValid() bool {
	return node != nil
}

func (node *As) IsValid() bool {
	return node != nil && node.Expr != nil && node.NewName != nil
}

func (node *External) IsValid() bool {
	return node != nil
}
