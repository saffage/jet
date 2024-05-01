package ast

import (
	"fmt"
	"strconv"
	"strings"
)

func (n *BadNode) String() string {
	return "@error(\"bad node\")"
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
		// TODO replace [strconv.Quote].
		return strconv.Quote(n.Value)

	default:
		panic("unreachable")
	}
}

func (n *Comment) String() string {
	return n.Data
}

func (n *CommentGroup) String() string {
	panic("todo")
}

func (n *ArrayType) String() string {
	return n.Args.String() + n.X.String()
}

func (n *Signature) String() string {
	buf := strings.Builder{}

	if n.Loc.Line > 0 {
		buf.WriteString("func")
	}

	buf.WriteString(n.Params.String())

	if n.Result != nil {
		buf.WriteByte(' ')
		buf.WriteString(n.Result.String())
	}

	return buf.String()
}

func (n *AttributeList) String() string {
	return "@" + n.List.String()
}

func (n *BuiltInCall) String() string {
	buf := strings.Builder{}
	buf.WriteByte('@')
	buf.WriteString(n.Name.String())

	if n.X != nil {
		if _, ok := n.X.(*ParenList); !ok {
			buf.WriteByte(' ')
		}

		buf.WriteString(n.X.String())
	}

	return buf.String()
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

		buf.WriteString(n.String())
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
	if len(n.List.Nodes) > 0 {
		return fmt.Sprintf("{ %s }", n.List.String())
	}

	return "{}"
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

func (n *PrefixOpr) String() string {
	return n.Kind.String()
}

func (n *InfixOpr) String() string {
	return n.Kind.String()
}

func (n *PostfixOpr) String() string {
	return n.Kind.String()
}

func (n *PrefixOp) String() string {
	return n.Opr.String() + n.X.String()
}

func (n *InfixOp) String() string {
	return fmt.Sprintf("%s %s %s", n.X.String(), n.Opr.String(), n.Y.String())
}

func (n *PostfixOp) String() string {
	return n.X.String() + n.Opr.String()
}

// TODO append documentation to the declarations.

func (n *ModuleDecl) String() string {
	return fmt.Sprintf("%smodule %s %s", optionalAttributeList(n.Attrs), n.Name.String(), n.Body.String())
}

func (n *GenericDecl) String() string {
	return fmt.Sprintf("%s%s %s", optionalAttributeList(n.Attrs), n.Kind.String(), n.Field.String())
}

func (n *FuncDecl) String() string {
	return fmt.Sprintf("%sfunc %s%s %s", optionalAttributeList(n.Attrs), n.Name.String(), n.Signature.String(), n.Body.String())
}

func (n *TypeAliasDecl) String() string {
	return fmt.Sprintf("%salias %s = %s", optionalAttributeList(n.Attrs), n.Name.String(), n.Expr.String())
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

func optionalAttributeList(attrs *AttributeList) string {
	if attrs == nil {
		return ""
	}

	return attrs.String() + " "
}
