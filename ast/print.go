package ast

import (
	"fmt"
	"strings"
)

var Indent = "  "

func (n *BadNode) String() string {
	return "#error(\"bad node\")"
}

func (n *Empty) String() string {
	return ";"
}

func (n *Ident) String() string {
	return n.Name
}

func (n *Literal) String() string {
	switch n.Kind {
	case IntLiteral, FloatLiteral:
		return n.Value

	case StringLiteral:
		return "\"" + n.Value + "\""

	default:
		panic("unreachable")
	}
}

func (n *Comment) String() string {
	return "#" + n.Data
}

func (n *CommentGroup) String() string {
	panic("todo")
}

func (n *Ellipsis) String() string {
	return "..." + n.X.String()
}

func (n *ArrayType) String() string {
	if n.N != nil {
		return "[" + n.N.String() + "]" + n.X.String()
	}
	return "[]" + n.X.String()
}

func (n *Signature) String() string {
	buf := strings.Builder{}
	// buf.WriteString("func")
	buf.WriteString(n.Params.String())

	if n.Result != nil {
		buf.WriteByte(' ')
		buf.WriteString(n.Result.String())
	}

	return buf.String()
}

func (n *Annotation) String() string {
	buf := strings.Builder{}
	buf.WriteByte('@')
	buf.WriteString(n.Name.String())

	if n.Args != nil {
		buf.WriteString(n.Args.String())
	}

	return buf.String()
}

func (n *Attribute) String() string {
	buf := strings.Builder{}
	buf.WriteByte('#')
	buf.WriteString(n.Name.String())

	if n.X != nil {
		buf.WriteString(n.X.String())
	}

	return buf.String()
}

func (n *Try) String() string {
	return n.X.String() + "?"
}

func (n *Unwrap) String() string {
	return n.X.String() + "!"
}

func (n *MemberAccess) String() string {
	return n.X.String() + "." + n.Selector.String()
}

func (n *Star) String() string {
	return "*"
}

func (n *Call) String() string {
	return n.X.String() + n.Args.String()
}

func (n *Index) String() string {
	return n.X.String() + n.Args.String()
}

func printList(nodes []Node, separator rune) string {
	buf := strings.Builder{}

	for i, n := range nodes {
		if i > 0 {
			buf.WriteRune(separator)
			buf.WriteByte(' ')
		}

		_, err := buf.WriteString(n.String())
		if err != nil {
			panic(err)
		}
	}

	return buf.String()
}

func (n *List) String() string {
	return printList(n.Nodes, ';')
}

func (n *ParenList) String() string {
	return fmt.Sprintf("(%s)", printList(n.Nodes, ','))
}

func (n *CurlyList) String() string {
	return fmt.Sprintf("{ %s }", n.List.String())
}

func (n *BracketList) String() string {
	return fmt.Sprintf("[%s]", printList(n.Nodes, ','))
}

func (n *Field) String() string {
	buf := strings.Builder{}

	for i, name := range n.Names {
		if i > 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(name.String())
	}

	if n.Type != nil {
		buf.WriteByte(' ')
		buf.WriteString(n.Type.String())
	}

	if n.Value != nil {
		buf.WriteString(" = ")
		buf.WriteString(n.Value.String())
	}

	return buf.String()
}

func (n *UnaryOp) String() string {
	return n.OpKind.String() + n.X.String()
}

func (n *BinaryOp) String() string {
	return fmt.Sprintf("%s %s %s", n.X.String(), n.OpKind.String(), n.Y.String())
}

// TODO append annotations and documentation to declarations.

func (n *ModuleDecl) String() string {
	return fmt.Sprintf("module %s %s", n.Name.String(), n.Body.String())
}

func (n *GenericDecl) String() string {
	return fmt.Sprintf("%s %s", n.Kind.String(), n.Field.String())
}

func (n *FuncDecl) String() string {
	return fmt.Sprintf("func %s%s %s", n.Name.String(), n.Signature.String(), n.Body.String())
}

func (n *StructDecl) String() string {
	return fmt.Sprintf("struct %s %s", n.Name.String(), n.Fields.String())
}

func (n *EnumDecl) String() string {
	return fmt.Sprintf("enum %s %s", n.Name.String(), n.Body.String())
}

func (n *TypeAliasDecl) String() string {
	return fmt.Sprintf("alias %s = %s", n.Name.String(), n.Expr.String())
}

func (n *If) String() string {
	if n.Else != nil {
		return fmt.Sprintf("if %s %s %s", n.Cond.String(), n.Body.String(), n.Else.String())
	}

	return fmt.Sprintf("if %s %s", n.Cond.String(), n.Body.String())
}

func (n *Else) String() string {
	return fmt.Sprintf("else %s", n.Body.String())
}

func (n *While) String() string {
	return fmt.Sprintf("while %s %s", n.Cond.String(), n.Body.String())
}

func (n *Return) String() string {
	return fmt.Sprintf("return %s", n.X.String())
}

func (n *Break) String() string {
	if n.Label != nil {
		return fmt.Sprintf("break %s", n.Label.String())
	}

	return "break"
}

func (n *Continue) String() string {
	if n.Label != nil {
		return fmt.Sprintf("continue %s", n.Label.String())
	}

	return "continue"
}
