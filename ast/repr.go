package ast

import (
	"fmt"
	"strconv"
	"strings"
)

//------------------------------------------------
// Atoms
//------------------------------------------------

func (n *BadNode) Repr() string {
	return "$bad_node"
}

func (n *Empty) Repr() string {
	return ";"
}

func (n *Ident) Repr() string {
	return n.Name
}

func (n *Literal) Repr() string {
	switch n.Kind {
	case IntLiteral, FloatLiteral:
		return n.Value

	case StringLiteral:
		// TODO replace [strconv.Quote].
		return strconv.Quote(n.Value)

	default:
		panic("unreachable")
	}
}

//------------------------------------------------
// Declaration
//------------------------------------------------

func (decl *Decl) Repr() string {
	buf := strings.Builder{}

	if decl.Attrs != nil {
		buf.WriteString(decl.Attrs.Repr())
		buf.WriteByte(' ')
	}

	if decl.Mut.IsValid() {
		buf.WriteString("mut ")
	}

	buf.WriteString(decl.Ident.Repr())

	c := ':'
	if decl.IsVar {
		c = '='
	}

	if decl.Type != nil {
		buf.WriteString(fmt.Sprintf(": %s", decl.Type.Repr()))

		if decl.Value != nil {
			buf.WriteString(fmt.Sprintf(" %c %s", c, decl.Value.Repr()))
		}
	} else if decl.Value != nil {
		buf.WriteString(fmt.Sprintf(" :%c %s", c, decl.Value.Repr()))
	}

	return buf.String()
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

func (n *ArrayType) Repr() string {
	return n.Args.Repr() + n.X.Repr()
}

func (n *StructType) Repr() string {
	buf := strings.Builder{}
	buf.WriteString("struct{")

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
	buf.WriteString("enum{")

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
	return "$" + n.Ident.Repr()
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
	buf := strings.Builder{}

	if n.X != nil {
		buf.WriteString(n.X.Repr())
	}

	buf.WriteString(n.Kind.String())

	if n.Y != nil {
		buf.WriteString(n.Y.Repr())
	}

	return buf.String()
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
	return fmt.Sprintf("for %s in %s %s", n.DeclList.Repr(), n.IterExpr.Repr(), n.Body.Repr())
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
