//go:build ignore

package ast

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func (n *BadNode) UnmarshalYAML(v *yaml.Node) error    { return check(n, "BadNode", v) }
func (n *Empty) UnmarshalYAML(v *yaml.Node) error      { return check(n, "Empty", v) }
func (n *Ident) UnmarshalYAML(v *yaml.Node) error      { return check(n, "Ident", v) }
func (n *Type) UnmarshalYAML(v *yaml.Node) error       { return check(n, "Type", v) }
func (n *TypeVar) UnmarshalYAML(v *yaml.Node) error    { return check(n, "TypeVar", v) }
func (n *Underscore) UnmarshalYAML(v *yaml.Node) error { return check(n, "Underscore", v) }
func (n *Literal) UnmarshalYAML(v *yaml.Node) error    { return check(n, "Literal", v) }

func (n *Comment) UnmarshalYAML(v *yaml.Node) error       { return check(n, "Comment", v) }
func (n *CommentGroup) UnmarshalYAML(v *yaml.Node) error  { return check(n, "CommentGroup", v) }
func (n *AttributeList) UnmarshalYAML(v *yaml.Node) error { return check(n, "AttributeList", v) }
func (n *Decl) UnmarshalYAML(v *yaml.Node) error          { return check(n, "Decl", v) }

func (n *ArrayType) UnmarshalYAML(v *yaml.Node) error  { return check(n, "ArrayType", v) }
func (n *StructType) UnmarshalYAML(v *yaml.Node) error { return check(n, "StructType", v) }
func (n *EnumType) UnmarshalYAML(v *yaml.Node) error   { return check(n, "EnumType", v) }
func (n *Signature) UnmarshalYAML(v *yaml.Node) error  { return check(n, "Signature", v) }
func (n *BuiltIn) UnmarshalYAML(v *yaml.Node) error    { return check(n, "BuiltIn", v) }
func (n *Call) UnmarshalYAML(v *yaml.Node) error       { return check(n, "Call", v) }
func (n *Index) UnmarshalYAML(v *yaml.Node) error      { return check(n, "Index", v) }
func (n *Function) UnmarshalYAML(v *yaml.Node) error   { return check(n, "Function", v) }
func (n *Dot) UnmarshalYAML(v *yaml.Node) error        { return check(n, "Dot", v) }
func (n *Deref) UnmarshalYAML(v *yaml.Node) error      { return check(n, "Deref", v) }
func (n *Op) UnmarshalYAML(v *yaml.Node) error         { return check(n, "Op", v) }

func (n *List) UnmarshalYAML(v *yaml.Node) error        { return check(n, "List", v) }
func (n *StmtList) UnmarshalYAML(v *yaml.Node) error    { return check(n, "StmtList", v) }
func (n *BracketList) UnmarshalYAML(v *yaml.Node) error { return check(n, "BracketList", v) }
func (n *ParenList) UnmarshalYAML(v *yaml.Node) error   { return check(n, "ParenList", v) }
func (n *CurlyList) UnmarshalYAML(v *yaml.Node) error   { return check(n, "CurlyList", v) }

func (n *If) UnmarshalYAML(v *yaml.Node) error       { return check(n, "If", v) }
func (n *Else) UnmarshalYAML(v *yaml.Node) error     { return check(n, "Else", v) }
func (n *While) UnmarshalYAML(v *yaml.Node) error    { return check(n, "While", v) }
func (n *For) UnmarshalYAML(v *yaml.Node) error      { return check(n, "For", v) }
func (n *When) UnmarshalYAML(v *yaml.Node) error     { return check(n, "When", v) }
func (n *Extern) UnmarshalYAML(v *yaml.Node) error   { return check(n, "Extern", v) }
func (n *Defer) UnmarshalYAML(v *yaml.Node) error    { return check(n, "Defer", v) }
func (n *Return) UnmarshalYAML(v *yaml.Node) error   { return check(n, "Return", v) }
func (n *Break) UnmarshalYAML(v *yaml.Node) error    { return check(n, "Break", v) }
func (n *Continue) UnmarshalYAML(v *yaml.Node) error { return check(n, "Continue", v) }
func (n *Import) UnmarshalYAML(v *yaml.Node) error   { return check(n, "Import", v) }

func check[T any](node *T, typename string, value *yaml.Node) error {
	var w struct {
		NodeKind string `yaml:"node_kind"`
	}
	if err := value.Decode(&w); err != nil {
		return err
	}
	if w.NodeKind != typename {
		return fmt.Errorf("invalid 'node_kind' value: '%s', expected '%s'", w.NodeKind, typename)
	}
	var n struct {
		Node T `yaml:",inline"`
	}
	if err := value.Decode(&n); err != nil {
		return err
	}
	*node = n.Node
	return nil
}
