package ast

import (
	"strconv"
	"strings"
)

func Render(node Node) string {
	buf := strings.Builder{}
	err := node.Render(&buf)

	if err != nil {
		buf.WriteString("<invalid-node>")
	}

	return buf.String()
}

//------------------------------------------------
// Atoms
//------------------------------------------------

func (node *BadNode) Render(buf *strings.Builder) error {
	buf.WriteString("#[bad_node]#")
	return nil
}

func (node *Lower) Render(buf *strings.Builder) error {
	buf.WriteString(node.Data)
	return nil
}

func (node *Upper) Render(buf *strings.Builder) error {
	buf.WriteString(node.Data)
	return nil
}

func (node *Placeholder) Render(buf *strings.Builder) error {
	buf.WriteString(node.Data)
	return nil
}

func (node *Literal) Render(buf *strings.Builder) error {
	switch node.Kind {
	case IntLiteral, FloatLiteral:
		buf.WriteString(node.Value)

	case StringLiteral:
		// TODO replace [strconv.Quote] with own implementation.
		buf.WriteString(strconv.Quote(node.Value[1 : len(node.Value)-1]))

	default:
		panic("unreachable")
	}

	return nil
}

//------------------------------------------------
// Declaration
//------------------------------------------------

func (node *LetDecl) Render(buf *strings.Builder) error {
	buf.WriteString("let ")

	node.Decl.Render(buf)

	buf.WriteString(" = ")

	node.Value.Render(buf)
	return nil
}

func (node *TypeAlias) Render(buf *strings.Builder) error {
	buf.WriteString("let ")

	node.Ident.Render(buf)

	if node.Args != nil {
		node.Args.Render(buf)
	}

	buf.WriteString(" = ")

	node.Expr.Render(buf)
	return nil
}

func (node *TypeDef) Render(buf *strings.Builder) error {
	buf.WriteString("let ")

	node.Ident.Render(buf)

	if node.Args != nil {
		node.Args.Render(buf)
	}

	buf.WriteByte(' ')

	node.Body.Render(buf)
	return nil
}

func (node *Decl) Render(buf *strings.Builder) error {
	if node.TypeTok.IsValid() {
		buf.WriteString("type ")
	}

	node.Ident.Render(buf)

	if node.Type != nil {
		buf.WriteByte(' ')

		node.Type.Render(buf)
	}

	return nil
}

func (node *Variant) Render(buf *strings.Builder) error {
	node.Name.Render(buf)

	if node.Params != nil {
		node.Params.Render(buf)
	}

	return nil
}

//------------------------------------------------
// Composite nodes
//------------------------------------------------

func (node *Label) Render(buf *strings.Builder) error {
	if node.Name != nil {
		node.Name.Render(buf)

		buf.WriteString(": ")
	} else {
		buf.WriteByte(':')
	}

	node.X.Render(buf)
	return nil
}

func (node *Signature) Render(buf *strings.Builder) error {
	node.Params.Render(buf)

	if node.Result != nil {
		buf.WriteByte(' ')

		node.Result.Render(buf)
	}

	return nil
}

func (node *Function) Render(buf *strings.Builder) error {
	buf.WriteString("fn")

	node.Signature.Render(buf)

	if node.Body != nil {
		buf.WriteString(" = ")

		node.Body.Render(buf)
	}

	return nil
}

func (node *Call) Render(buf *strings.Builder) error {
	node.X.Render(buf)
	node.Args.Render(buf)
	return nil
}

func (node *Dot) Render(buf *strings.Builder) error {
	node.X.Render(buf)

	buf.WriteByte('.')

	node.Y.Render(buf)
	return nil
}

func (node *Op) Render(buf *strings.Builder) error {
	if node.X != nil {
		node.X.Render(buf)
	}

	buf.WriteByte(' ')
	buf.WriteString(node.Kind.String())
	buf.WriteByte(' ')

	if node.Y != nil {
		node.Y.Render(buf)
	}

	return nil
}

//------------------------------------------------
// Lists
//------------------------------------------------

func (node *List) Render(buf *strings.Builder) error {
	buf.WriteByte('[')

	for i, node := range node.Nodes {
		if i > 0 {
			buf.WriteString(", ")
		}

		node.Render(buf)
	}

	buf.WriteByte(']')
	return nil
}

func (stmts Stmts) Render(buf *strings.Builder) error {
	for i, stmt := range stmts.Items {
		if i > 0 {
			buf.WriteString("; ")
		}

		stmt.Render(buf)
	}

	return nil
}

func (node *Block) Render(buf *strings.Builder) error {
	if len(node.Stmts.Items) == 0 {
		buf.WriteString("{}")
	} else {
		buf.WriteString("{ ")

		node.Stmts.Render(buf)

		buf.WriteString(" }")
	}

	return nil
}

func (node *Parens) Render(buf *strings.Builder) error {
	buf.WriteByte('(')

	for i, node := range node.Nodes {
		if i > 0 {
			buf.WriteString(", ")
		}

		node.Render(buf)
	}

	buf.WriteByte(')')
	return nil
}

//------------------------------------------------
// Language constructions
//------------------------------------------------

func (node *When) Render(buf *strings.Builder) error {
	buf.WriteString("when ")

	node.Expr.Render(buf)

	buf.WriteByte(' ')

	node.Body.Render(buf)
	return nil
}

func (node *Case) Render(buf *strings.Builder) error {
	node.Pattern.Render(buf)

	buf.WriteString(" -> ")

	node.Expr.Render(buf)
	return nil
}

func (node *Spread) Render(buf *strings.Builder) error {
	buf.WriteString("..")

	if node.Expr != nil {
		node.Expr.Render(buf)
	}
	return nil
}

func (node *As) Render(buf *strings.Builder) error {
	node.Expr.Render(buf)

	buf.WriteString(" as ")

	node.NewName.Render(buf)
	return nil
}

func (node *Extern) Render(buf *strings.Builder) error {
	buf.WriteString("extern")

	if node.Args != nil {
		node.Args.Render(buf)
	}
	return nil
}
