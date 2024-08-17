package ast

import (
	"fmt"
	"strconv"
	"strings"
)

type Representable interface {
	// Representation of the node tree. Result must be equal to the
	// code from which this tree can be parsed.
	Repr() string
}

//------------------------------------------------
// Atoms
//------------------------------------------------

func (n *BadNode) Repr() string {
	return "$bad_node"
}

func (n *Empty) Repr() string {
	return ";"
}

func (n *Name) Repr() string {
	return n.Data
}

func (n *Type) Repr() string {
	return n.Data
}

func (n *Underscore) Repr() string {
	return n.Data
}

func (n *Literal) Repr() string {
	switch n.Kind {
	case IntLiteral, FloatLiteral:
		return n.Data

	case StringLiteral:
		// TODO replace [strconv.Quote].
		return strconv.Quote(n.Data)

	default:
		panic("unreachable")
	}
}

//------------------------------------------------
// Declaration
//------------------------------------------------

func (n *LetDecl) Repr() string {
	if n.Attrs != nil {
		return fmt.Sprintf(
			"%s let %s = %s",
			n.Attrs.Repr(),
			n.Decl.Repr(),
			n.Value.Repr(),
		)
	}

	return fmt.Sprintf(
		"let %s = %s",
		n.Decl.Repr(),
		n.Value.Repr(),
	)
}

func (n *TypeDecl) Repr() string {
	buf := strings.Builder{}

	if n.Attrs != nil {
		buf.WriteString(n.Attrs.Repr())
		buf.WriteByte(' ')
	}

	buf.WriteString("type " + n.Type.Repr())

	if n.Args != nil {
		buf.WriteString(n.Args.Repr())
	}

	buf.WriteString(" = " + n.Expr.Repr())
	return buf.String()
}

func (n *Decl) Repr() string {
	if n.Type != nil {
		return n.Name.Repr() + " " + n.Type.Repr()
	}

	return n.Name.Repr()
}

func (n *AttributeList) Repr() string {
	return "@" + n.List.Repr()
}

func (n *Comment) Repr() string {
	return "##" + n.Value
}

func (n *CommentGroup) Repr() string {
	buf := strings.Builder{}

	for _, comment := range n.Comments {
		buf.WriteString(comment.Repr())
		buf.WriteByte('\n')
	}

	return buf.String()
}

//------------------------------------------------
// Composite nodes
//------------------------------------------------

func (n *Label) Repr() string {
	return n.Label.Repr() + ": " + n.X.Repr()
}

func (n *ArrayType) Repr() string {
	return n.Args.Repr() + n.X.Repr()
}

func (n *StructType) Repr() string {
	buf := strings.Builder{}
	buf.WriteString("struct {")

	for i, field := range n.Fields {
		if i != 0 {
			buf.WriteByte(';')
		}

		buf.WriteByte(' ')
		buf.WriteString(field.Repr())
	}

	buf.WriteString(" }")
	return buf.String()
}

func (n *EnumType) Repr() string {
	buf := strings.Builder{}
	buf.WriteString("enum {")

	for i, field := range n.Fields {
		if i != 0 {
			buf.WriteByte(';')
		}

		buf.WriteByte(' ')
		buf.WriteString(field.Repr())
	}

	buf.WriteString(" }")
	return buf.String()
}

func (n *Signature) Repr() string {
	if n.Result == nil {
		return fmt.Sprintf("%s -> ()", n.Params.Repr())
	}

	return fmt.Sprintf("%s -> %s", n.Params.Repr(), n.Result.Repr())
}

func (n *BuiltIn) Repr() string {
	return "$" + n.Name.Repr()
}

func (n *Call) Repr() string {
	return n.X.Repr() + n.Args.Repr()
}

func (n *Index) Repr() string {
	return n.X.Repr() + n.Args.Repr()
}

func (n *Function) Repr() string {
	if n.Signature.Result == nil {
		return fmt.Sprintf("%s %s", n.Signature.Params.Repr(), n.Body.Repr())
	}

	return fmt.Sprintf("%s %s", n.Signature.Repr(), n.Body.Repr())
}

func (n *Dot) Repr() string {
	return n.X.Repr() + "." + n.Y.Repr()
}

func (n *Deref) Repr() string {
	return n.X.Repr() + ".*"
}

func (n *Op) Repr() string {
	if n.X != nil {
		if n.Y != nil {
			return fmt.Sprintf(
				"%s %s %s",
				n.X.Repr(),
				n.Kind.String(),
				n.Y.Repr(),
			)
		}
		return n.X.Repr() + n.Kind.String()
	}
	if n.Y != nil {
		return n.Kind.String() + n.Y.Repr()
	}
	return n.Kind.String()
}

//------------------------------------------------
// Lists
//------------------------------------------------

func (n *List) Repr() string {
	return printList(n.Nodes, ',')
}

func (n *StmtList) Repr() string {
	return printList(n.Nodes, ';')
}

func (n *BracketList) Repr() string {
	return fmt.Sprintf("[%s]", printList(n.Nodes, ','))
}

func (n *ParenList) Repr() string {
	return fmt.Sprintf("(%s)", printList(n.Nodes, ','))
}

func (n *CurlyList) Repr() string {
	if len(n.StmtList.Nodes) > 0 {
		return fmt.Sprintf("{ %s }", n.StmtList.Repr())
	}

	return "{}"
}

func printList[T Node](nodes []T, separator rune) string {
	buf := strings.Builder{}

	for i, n := range nodes {
		if i > 0 {
			buf.WriteRune(separator)
			buf.WriteByte(' ')
		}

		buf.WriteString(n.Repr())
	}

	return buf.String()
}

//------------------------------------------------
// Language constructions
//------------------------------------------------

func (n *If) Repr() string {
	if n.Else != nil {
		return fmt.Sprintf("if %s %s %s", n.Cond.Repr(), n.Body.Repr(), n.Else.Repr())
	}

	return fmt.Sprintf("if %s %s", n.Cond.Repr(), n.Body.Repr())
}

func (n *Else) Repr() string {
	return fmt.Sprintf("else %s", n.Body.Repr())
}

func (n *While) Repr() string {
	return fmt.Sprintf("while %s %s", n.Cond.Repr(), n.Body.Repr())
}

func (n *For) Repr() string {
	return fmt.Sprintf("for %s in %s %s", n.Decls.Repr(), n.IterExpr.Repr(), n.Body.Repr())
}

func (n *When) Repr() string {
	return fmt.Sprintf("when %s %s", n.Expr.Repr(), n.Body.Repr())
}

func (n *Defer) Repr() string {
	return fmt.Sprintf("defer %s", n.X.Repr())
}

func (n *Return) Repr() string {
	if n.X != nil {
		return fmt.Sprintf("return %s", n.X.Repr())
	}

	return "return"
}

func (n *Break) Repr() string {
	if n.Label != nil {
		return fmt.Sprintf("break %s", n.Label.Repr())
	}

	return "break"
}

func (n *Continue) Repr() string {
	if n.Label != nil {
		return fmt.Sprintf("continue %s", n.Label.Repr())
	}

	return "continue"
}

func (n *Import) Repr() string {
	return fmt.Sprintf("import %s", n.Module)
}
