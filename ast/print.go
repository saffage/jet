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
	// return fmt.Sprintf("(loc: %s, name: %s)", color.CyanString(n.Start.String()), n.Name)
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
	return "##" + n.Data
}

func (n *CommentGroup) String() string {
	buf := strings.Builder{}

	for _, comment := range n.Comments {
		buf.WriteString(comment.String())
		buf.WriteByte('\n')
	}

	return buf.String()
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

	if n.Args != nil {
		if _, ok := n.Args.(*ParenList); !ok {
			buf.WriteByte(' ')
		}

		buf.WriteString(n.Args.String())
	}

	return buf.String()
}

func (n *MemberAccess) String() string {
	return n.X.String() + "." + n.Selector.String()
}

func (n *SafeMemberAccess) String() string {
	return n.X.String() + "?." + n.Selector.String()
}

func (n *Call) String() string {
	return n.X.String() + n.Args.String()
}

func (n *Index) String() string {
	return n.X.String() + n.Args.String()
}

func printList[T Node](nodes []T, separator rune) string {
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

func (n *ExprList) String() string {
	return printList(n.Exprs, ',')
}

func (n *ParenList) String() string {
	return fmt.Sprintf("(%s)", printList(n.Exprs, ','))
}

func (n *CurlyList) String() string {
	if len(n.List.Nodes) > 0 {
		return fmt.Sprintf("{ %s }", n.List.String())
	}

	return "{}"
}

func (n *BracketList) String() string {
	return fmt.Sprintf("[%s]", printList(n.Exprs, ','))
}

func (n *BindingWithValue) String() string {
	if n.Operator != nil {
		return fmt.Sprintf("%s %s %s", n.Binding.String(), n.Operator.String(), n.Value.String())
	}

	return n.Binding.String()
}

func (n *Binding) String() string {
	buf := strings.Builder{}
	buf.WriteString(n.Name.String())

	if n.Type != nil {
		buf.WriteByte(' ')
		buf.WriteString(n.Type.String())
	}

	return buf.String()
}

func (n *Operator) String() string {
	return n.Kind.String()
}

func (n *PrefixOp) String() string {
	return n.Opr.String() + n.X.String()
}

func (n *InfixOp) String() string {
	return fmt.Sprintf("(%s %s %s)", n.X.String(), n.Opr.String(), n.Y.String())
}

func (n *PostfixOp) String() string {
	return n.X.String() + n.Opr.String()
}

func (n *ModuleDecl) String() string {
	return fmt.Sprintf(
		"%s%smodule %s %s",
		optionalComment(n.CommentGroup),
		optionalAttributeList(n.Attrs),
		n.Name.String(),
		n.Body.String(),
	)
}

func (n *VarDecl) String() string {
	if n.Value != nil {
		return fmt.Sprintf(
			"%s%svar %s = %s",
			optionalComment(n.CommentGroup),
			optionalAttributeList(n.Attrs),
			n.Binding.String(),
			n.Value.String(),
		)
	}

	return fmt.Sprintf(
		"%s%svar %s",
		optionalComment(n.CommentGroup),
		optionalAttributeList(n.Attrs),
		n.Binding.String(),
	)
}

func (n *ConstDecl) String() string {
	return fmt.Sprintf(
		"%s%sconst %s",
		optionalComment(n.CommentGroup),
		optionalAttributeList(n.Attrs),
		n.Binding.String(),
	)
}

func (n *FuncDecl) String() string {
	if n.Body != nil {
		return fmt.Sprintf(
			"%s%sfunc %s%s %s",
			optionalComment(n.CommentGroup),
			optionalAttributeList(n.Attrs),
			n.Name.String(),
			n.Signature.String(),
			n.Body.String(),
		)
	}

	return fmt.Sprintf(
		"%s%sfunc %s%s",
		optionalComment(n.CommentGroup),
		optionalAttributeList(n.Attrs),
		n.Name.String(),
		n.Signature.String(),
	)
}

func (n *StructDecl) String() string {
	return fmt.Sprintf(
		"%sstruct %s %s",
		optionalAttributeList(n.Attrs),
		n.Name.String(),
		n.Body.String(),
	)
}

func (n *EnumDecl) String() string {
	return fmt.Sprintf(
		"%senum %s %s",
		optionalAttributeList(n.Attrs),
		n.Name.String(),
		n.Body.String(),
	)
}

func (n *TypeAliasDecl) String() string {
	return fmt.Sprintf(
		"%s%salias %s = %s",
		optionalComment(n.CommentGroup),
		optionalAttributeList(n.Attrs),
		n.Name.String(),
		n.Expr.String(),
	)
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

func (n *Import) String() string {
	return fmt.Sprintf("import %s", n.Module)
}

func optionalComment(commentGroup *CommentGroup) string {
	if commentGroup == nil {
		return ""
	}

	return commentGroup.String()
}

func optionalAttributeList(attrs *AttributeList) string {
	if attrs == nil {
		return ""
	}

	return attrs.String() + " "
}
