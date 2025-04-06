package ast

import (
	"strconv"
	"strings"

	"github.com/saffage/jet/token"
)

func Render(node Node) string {
	const initialBufferSize = 256

	buf := strings.Builder{}
	buf.Grow(initialBufferSize)
	render(node, &buf)

	return buf.String()
}

func render[T SomeNode](node T, buf *strings.Builder) {
	var zero T

	if node == zero {
		buf.WriteString("#[nil-ast-node]#")
	} else if !node.Renderable() {
		buf.WriteString("#[ill-formed-ast]#")
	} else {
		node.Render(buf)
	}
}

func (node *BadNode) Render(buf *strings.Builder) {
	buf.WriteString("#[bad-ast-node]#")
}

func (node *Lower) Render(buf *strings.Builder) {
	buf.WriteString(node.Data)
}

func (node *Upper) Render(buf *strings.Builder) {
	buf.WriteString(node.Data)
}

func (node *Placeholder) Render(buf *strings.Builder) {
	buf.WriteString(node.Data)
}

func (node *Literal) Render(buf *strings.Builder) {
	switch node.Kind {
	case IntLiteral, FloatLiteral:
		buf.WriteString(node.Value)

	case StringLiteral:
		// TODO replace [strconv.Quote] with own implementation.
		buf.WriteString(strconv.Quote(node.Value[1 : len(node.Value)-1]))

	default:
		panic("unreachable")
	}

}

func (node *LetDecl) Render(buf *strings.Builder) {
	buf.WriteString("let ")

	render(node.Decl, buf)

	buf.WriteString(" = ")

	render(node.Value, buf)
}

func (node *ValDecl) Render(buf *strings.Builder) {
	buf.WriteString("val ")

	render(node.Decl, buf)

	buf.WriteString(" = ")

	render(node.Value, buf)
}

func (node *VarDecl) Render(buf *strings.Builder) {
	buf.WriteString("var ")

	render(node.Decl, buf)

	buf.WriteString(" = ")

	render(node.Value, buf)
}

func (node *TypeAlias) Render(buf *strings.Builder) {
	buf.WriteString("type ")

	render(node.Ident, buf)

	if node.Args != nil {
		render(node.Args, buf)
	}

	buf.WriteString(" = ")

	render(node.Expr, buf)
}

func (node *TypeDef) Render(buf *strings.Builder) {
	buf.WriteString("type ")

	render(node.Ident, buf)

	if node.Args != nil {
		render(node.Args, buf)
	}

	buf.WriteByte(' ')

	render(node.Body, buf)
}

func (node *Decl) Render(buf *strings.Builder) {
	if node.TypeTok.IsValid() {
		buf.WriteString("type ")
	}

	render(node.Ident, buf)

	if node.Type != nil {
		buf.WriteByte(' ')

		render(node.Type, buf)
	}

}

func (node *Variant) Render(buf *strings.Builder) {
	render(node.Name, buf)

	if node.Params != nil {
		render(node.Params, buf)
	}

}

func (node *Label) Render(buf *strings.Builder) {
	if node.Name != nil {
		render(node.Name, buf)

		buf.WriteString(": ")
	} else {
		buf.WriteByte(':')
	}

	render(node.X, buf)
}

func (node *Signature) Render(buf *strings.Builder) {
	render(node.Params, buf)

	if node.Result != nil {
		buf.WriteByte(' ')

		render(node.Result, buf)
	}

}

func (node *Function) Render(buf *strings.Builder) {
	buf.WriteString("fn")

	render(node.Signature, buf)

	if node.Body != nil {
		buf.WriteString(" = ")

		render(node.Body, buf)
	}

}

func (node *Call) Render(buf *strings.Builder) {
	render(node.X, buf)
	render(node.Args, buf)
}

func (node *Dot) Render(buf *strings.Builder) {
	render(node.X, buf)

	buf.WriteByte('.')

	render(node.Y, buf)
}

func (node *Op) Render(buf *strings.Builder) {
	if node.X != nil {
		render(node.X, buf)
	}

	buf.WriteByte(' ')
	buf.WriteString(node.Kind.String())
	buf.WriteByte(' ')

	if node.Y != nil {
		render(node.Y, buf)
	}

}

func (node *List) Render(buf *strings.Builder) {
	buf.WriteByte('[')

	for i, node := range node.Nodes {
		if i > 0 {
			buf.WriteString(", ")
		}

		render(node, buf)
	}

	buf.WriteByte(']')
}

func (stmts Stmts) Render(buf *strings.Builder) {
	for i, stmt := range stmts.Items {
		if i > 0 {
			buf.WriteString("; ")
		}

		render(stmt, buf)
	}

}

func (node *Block) Render(buf *strings.Builder) {
	if len(node.Stmts.Items) == 0 {
		buf.WriteString("{}")
	} else {
		buf.WriteString("{ ")

		render(node.Stmts, buf)

		buf.WriteString(" }")
	}

}

func (node *Parens) Render(buf *strings.Builder) {
	buf.WriteByte('(')

	for i, node := range node.Nodes {
		if i > 0 {
			buf.WriteString(", ")
		}

		render(node, buf)
	}

	buf.WriteByte(')')
}

func (node *When) Render(buf *strings.Builder) {
	buf.WriteString("when ")

	render(node.Expr, buf)

	buf.WriteByte(' ')

	render(node.Body, buf)
}

func (node *Case) Render(buf *strings.Builder) {
	render(node.Pattern, buf)

	buf.WriteString(" -> ")

	render(node.Expr, buf)
}

func (node *Spread) Render(buf *strings.Builder) {
	buf.WriteString("..")

	if node.Expr != nil {
		render(node.Expr, buf)
	}
}

func (node *As) Render(buf *strings.Builder) {
	render(node.Expr, buf)

	buf.WriteString(" as ")

	render(node.NewName, buf)
}

func (node *External) Render(buf *strings.Builder) {
	buf.WriteString("external")

	if node.Args != nil {
		render(node.Args, buf)
	}
}

//
//
//

func (node *BadNode) Renderable() bool {
	return node != nil
}

func (node *Lower) Renderable() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Upper) Renderable() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Placeholder) Renderable() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Literal) Renderable() bool {
	return node != nil && node.Value != "" && // TODO: check the value
		(IntLiteral <= node.Kind && node.Kind <= StringLiteral)
}

func (node *LetDecl) Renderable() bool {
	return node != nil && node.Decl != nil && node.Value != nil
}

func (node *ValDecl) Renderable() bool {
	return node != nil && node.Decl != nil && node.Value != nil
}

func (node *VarDecl) Renderable() bool {
	return node != nil && node.Decl != nil && node.Value != nil
}

func (node *TypeAlias) Renderable() bool {
	return node != nil && node.Ident != nil && node.Expr != nil
}

func (node *TypeDef) Renderable() bool {
	return node != nil && node.Ident != nil && node.Body != nil
}

func (node *Decl) Renderable() bool {
	return node != nil && node.Ident != nil
}

func (node *Variant) Renderable() bool {
	return node != nil && node.Name != nil
}

func (node *Label) Renderable() bool {
	return node != nil && node.X != nil
}

func (node *Signature) Renderable() bool {
	return node != nil && node.Params != nil
}

func (node *Function) Renderable() bool {
	return node != nil && node.Signature != nil && node.Body != nil
}

func (node *Call) Renderable() bool {
	return node != nil && node.X != nil && node.Args != nil
}

func (node *Dot) Renderable() bool {
	return node != nil && node.X != nil && node.Y != nil
}

func (node *Op) Renderable() bool {
	return node != nil &&
		(OperatorNot <= node.Kind && node.Kind <= OperatorBitOr)
}

func (stmts Stmts) Renderable() bool {
	return true
}

func (node *Block) Renderable() bool {
	return node != nil && node.Stmts != nil
}

func (node *Parens) Renderable() bool {
	return node != nil
}

func (node *List) Renderable() bool {
	return node != nil
}

func (node *When) Renderable() bool {
	return node != nil && node.Expr != nil && node.Body != nil
}

func (node *Case) Renderable() bool {
	return node != nil && node.Pattern != nil && node.Expr != nil
}

func (node *Spread) Renderable() bool {
	return node != nil
}

func (node *As) Renderable() bool {
	return node != nil && node.Expr != nil && node.NewName != nil
}

func (node *External) Renderable() bool {
	return node != nil
}
