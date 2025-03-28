package ast

import (
	"strconv"
	"strings"
)

func Render(node Node) string {
	const initialBufferSize = 256

	buf := strings.Builder{}
	buf.Grow(initialBufferSize)

	node.Render(&buf)

	return buf.String()
}

//------------------------------------------------
// Atoms
//------------------------------------------------

func (node *BadNode) Render(buf *strings.Builder) {
	buf.WriteString("#[bad_node]#")
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

//------------------------------------------------
// Declaration
//------------------------------------------------

func (node *LetDecl) Render(buf *strings.Builder) {
	buf.WriteString("let ")

	node.Decl.Render(buf)

	buf.WriteString(" = ")

	node.Value.Render(buf)
}

func (node *TypeAlias) Render(buf *strings.Builder) {
	buf.WriteString("let ")

	node.Ident.Render(buf)

	if node.Args != nil {
		node.Args.Render(buf)
	}

	buf.WriteString(" = ")

	node.Expr.Render(buf)
}

func (node *TypeDef) Render(buf *strings.Builder) {
	buf.WriteString("let ")

	node.Ident.Render(buf)

	if node.Args != nil {
		node.Args.Render(buf)
	}

	buf.WriteByte(' ')

	node.Body.Render(buf)
}

func (node *Decl) Render(buf *strings.Builder) {
	if node.TypeTok.IsValid() {
		buf.WriteString("type ")
	}

	node.Ident.Render(buf)

	if node.Type != nil {
		buf.WriteByte(' ')

		node.Type.Render(buf)
	}

}

func (node *Variant) Render(buf *strings.Builder) {
	node.Name.Render(buf)

	if node.Params != nil {
		node.Params.Render(buf)
	}

}

//------------------------------------------------
// Composite nodes
//------------------------------------------------

func (node *Label) Render(buf *strings.Builder) {
	if node.Name != nil {
		node.Name.Render(buf)

		buf.WriteString(": ")
	} else {
		buf.WriteByte(':')
	}

	node.X.Render(buf)
}

func (node *Signature) Render(buf *strings.Builder) {
	node.Params.Render(buf)

	if node.Result != nil {
		buf.WriteByte(' ')

		node.Result.Render(buf)
	}

}

func (node *Function) Render(buf *strings.Builder) {
	buf.WriteString("fn")

	node.Signature.Render(buf)

	if node.Body != nil {
		buf.WriteString(" = ")

		node.Body.Render(buf)
	}

}

func (node *Call) Render(buf *strings.Builder) {
	node.X.Render(buf)
	node.Args.Render(buf)
}

func (node *Dot) Render(buf *strings.Builder) {
	node.X.Render(buf)

	buf.WriteByte('.')

	node.Y.Render(buf)
}

func (node *Op) Render(buf *strings.Builder) {
	if node.X != nil {
		node.X.Render(buf)
	}

	buf.WriteByte(' ')
	buf.WriteString(node.Kind.String())
	buf.WriteByte(' ')

	if node.Y != nil {
		node.Y.Render(buf)
	}

}

//------------------------------------------------
// Lists
//------------------------------------------------

func (node *List) Render(buf *strings.Builder) {
	buf.WriteByte('[')

	for i, node := range node.Nodes {
		if i > 0 {
			buf.WriteString(", ")
		}

		node.Render(buf)
	}

	buf.WriteByte(']')
}

func (stmts Stmts) Render(buf *strings.Builder) {
	for i, stmt := range stmts.Items {
		if i > 0 {
			buf.WriteString("; ")
		}

		stmt.Render(buf)
	}

}

func (node *Block) Render(buf *strings.Builder) {
	if len(node.Stmts.Items) == 0 {
		buf.WriteString("{}")
	} else {
		buf.WriteString("{ ")

		node.Stmts.Render(buf)

		buf.WriteString(" }")
	}

}

func (node *Parens) Render(buf *strings.Builder) {
	buf.WriteByte('(')

	for i, node := range node.Nodes {
		if i > 0 {
			buf.WriteString(", ")
		}

		node.Render(buf)
	}

	buf.WriteByte(')')
}

//------------------------------------------------
// Language constructions
//------------------------------------------------

func (node *When) Render(buf *strings.Builder) {
	buf.WriteString("when ")

	node.Expr.Render(buf)

	buf.WriteByte(' ')

	node.Body.Render(buf)
}

func (node *Case) Render(buf *strings.Builder) {
	node.Pattern.Render(buf)

	buf.WriteString(" -> ")

	node.Expr.Render(buf)
}

func (node *Spread) Render(buf *strings.Builder) {
	buf.WriteString("..")

	if node.Expr != nil {
		node.Expr.Render(buf)
	}
}

func (node *As) Render(buf *strings.Builder) {
	node.Expr.Render(buf)

	buf.WriteString(" as ")

	node.NewName.Render(buf)
}

func (node *Extern) Render(buf *strings.Builder) {
	buf.WriteString("extern")

	if node.Args != nil {
		node.Args.Render(buf)
	}
}
