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

func (node *Capitalized) Render(buf text.Writer) (rendered bool) {
	buf.WriteString(node.Data)
	return true
}

func (node *Placeholder) Render(buf text.Writer) (rendered bool) {
	buf.WriteString(node.Data)
	return true
}

func (node *TypeVariable) Render(buf text.Writer) (rendered bool) {
	buf.WriteByte('\'')
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

func (node *ValueDecl) Render(buf text.Writer) (rendered bool) {
	switch node.Kind {
	case ValueLet:
		buf.WriteString("let ")
	case ValueVal:
		buf.WriteString("val ")
	case ValueVar:
		buf.WriteString("var ")
	}

	render(node.Pattern, buf)
	buf.WriteString(" = ")
	render(node.Value, buf)
	return true
}

func (node *TypeAliasDecl) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("type ")
	render(node.Ident, buf)

	if node.Args != nil {
		renderParens(node.Args, buf)
	}

	buf.WriteString(" = ")
	render(node.Expr, buf)
	return true
}

func (node *TypeDecl) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("type ")
	render(node.Ident, buf)

	if node.Args != nil {
		renderParens(node.Args, buf)
	}

	buf.WriteByte(' ')
	render(node.Body, buf)
	return true
}

func (node *FieldDecl) Render(buf text.Writer) (rendered bool) {
	render(node.Label, buf)
	buf.WriteByte(' ')
	render(node.Type, buf)
	return true
}

func (node *VariantDecl) Render(buf text.Writer) (rendered bool) {
	render(node.Name, buf)

	if node.Params != nil {
		renderParens(node.Params, buf)
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
	renderParens(node.Params, buf)

	if node.Result != nil {
		buf.WriteByte(' ')
		render(node.Result, buf)
	}

	return true
}

func (node *FnType) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("fn")
	render(node.Signature, buf)
	return true
}

func (node *Fn) Render(buf text.Writer) (rendered bool) {
	render(node.Type, buf)
	buf.WriteByte(' ')
	render(node.Body, buf)
	return true
}

func (node *Call) Render(buf text.Writer) (rendered bool) {
	render(node.X, buf)
	renderParens(node.Args, buf)
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

func (node *When) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("when ")

	render(node.Cond, buf)

	buf.WriteByte(' ')

	render(node.Clauses, buf)
	return true
}

func (node *CaseClause) Render(buf text.Writer) (rendered bool) {
	render(node.Pattern, buf)

	buf.WriteString(" -> ")

	render(node.Expr, buf)
	return true
}

func (node *External) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("external")

	if node.Args != nil {
		renderParens(node.Args, buf)
	}
	return true
}

func renderParens(node *Parens, buf text.Writer) (rendered bool) {
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

//
//
//

func (node *BadNode) IsValid() bool {
	return node != nil
}

func (node *Lower) IsValid() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Capitalized) IsValid() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Placeholder) IsValid() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *TypeVariable) IsValid() bool {
	return node != nil && token.IsValidIdent(node.Data)
}

func (node *Literal) IsValid() bool {
	return node != nil && node.Value != "" && // TODO: check the value
		(IntLiteral <= node.Kind && node.Kind <= StringLiteral)
}

func (node *ValueDecl) IsValid() bool {
	return node != nil &&
		node.Pattern != nil &&
		node.Value != nil &&
		node.KeywordPos.IsValid() &&
		(node.Kind >= ValueLet && node.Kind <= ValueVar)
}

func (node *TypeAliasDecl) IsValid() bool {
	return node != nil && node.Ident != nil && node.Expr != nil
}

func (node *TypeDecl) IsValid() bool {
	return node != nil && node.Ident != nil && node.Body != nil
}

func (node *FieldDecl) IsValid() bool {
	return node != nil && node.Name != nil && node.Type != nil
}

func (node *VariantDecl) IsValid() bool {
	return node != nil && node.Name != nil
}

func (node *Label) IsValid() bool {
	return node != nil && node.X != nil
}

func (node *Signature) IsValid() bool {
	return node != nil && node.Params != nil
}

func (node *FnType) IsValid() bool {
	return node != nil && node.Signature.IsValid()
}

func (node *Fn) IsValid() bool {
	return node != nil && node.Type.IsValid() && node.Body.IsValid()
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

func (node *List) IsValid() bool {
	return node != nil
}

func (node *When) IsValid() bool {
	return node != nil && node.Cond != nil && node.Clauses != nil
}

func (node *CaseClause) IsValid() bool {
	return node != nil && node.Pattern != nil && node.Expr != nil
}

func (node *External) IsValid() bool {
	return node != nil
}

//
//
//

func (pattern *PatternInvalid) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("#[bad-pattern]#")
	return true
}

func (pattern *PatternLiteral) Render(buf text.Writer) (rendered bool) {
	render(pattern.Value, buf)
	return true
}

func (pattern *PatternBinding) Render(buf text.Writer) (rendered bool) {
	render(pattern.Name, buf)
	return true
}

func (pattern *PatternPlaceholder) Render(buf text.Writer) (rendered bool) {
	render(pattern.Name, buf)
	return true
}

func (pattern *PatternRebinding) Render(buf text.Writer) (rendered bool) {
	render(pattern.X, buf)
	buf.WriteString(" as ")
	render(pattern.Name, buf)
	return true
}

func (pattern *PatternList) Render(buf text.Writer) (rendered bool) {
	buf.WriteByte('[')

	if len(pattern.Items) != 0 {
		render(pattern.Items[0], buf)

		for _, item := range pattern.Items[1:] {
			buf.WriteString(", ")
			render(item, buf)
		}
	}

	buf.WriteByte(']')
	return true
}

func (pattern *PatternRange) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("...")

	if pattern.Name != nil {
		render(pattern.Name, buf)
	}

	return true
}

func (pattern *PatternVariant) Render(buf text.Writer) (rendered bool) {
	render(pattern.Name, buf)

	if pattern.Parens.IsValid() {
		buf.WriteByte('(')

		if len(pattern.Values) != 0 {
			render(pattern.Values[0], buf)

			for _, item := range pattern.Values[1:] {
				buf.WriteString(", ")
				render(item, buf)
			}
		}

		buf.WriteByte(')')
	}
	return true
}

func (pattern *PatternLabeled) Render(buf text.Writer) (rendered bool) {
	if pattern.Label != nil {
		render(pattern.Label, buf)
		buf.WriteString(": ")
	} else {
		buf.WriteByte(':')
	}

	render(pattern.X, buf)
	return true
}

func (pattern *PatternTypeTest) Render(buf text.Writer) (rendered bool) {
	render(pattern.X, buf)
	buf.WriteByte(' ')
	render(pattern.Type, buf)
	return true
}

func (pattern *PatternAlternative) Render(buf text.Writer) (rendered bool) {
	if len(pattern.Items) != 0 {
		render(pattern.Items[0], buf)

		for _, item := range pattern.Items[1:] {
			buf.WriteString(" | ")
			render(item, buf)
		}
	}
	return true
}

func (pattern *PatternInvalid) IsValid() bool {
	return false
}

func (pattern *PatternLiteral) IsValid() bool {
	return false
}

func (pattern *PatternBinding) IsValid() bool {
	return false
}

func (pattern *PatternPlaceholder) IsValid() bool {
	return false
}

func (pattern *PatternRebinding) IsValid() bool {
	return false
}

func (pattern *PatternList) IsValid() bool {
	return false
}

func (pattern *PatternRange) IsValid() bool {
	return false
}

func (pattern *PatternVariant) IsValid() bool {
	return false
}

func (pattern *PatternLabeled) IsValid() bool {
	return false
}

func (pattern *PatternTypeTest) IsValid() bool {
	return false
}

func (pattern *PatternAlternative) IsValid() bool {
	return false
}
