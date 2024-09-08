package ast

import (
	"reflect"
	"strings"
)

// TODO: fix a bug when 'UnmarshalYAML' method brokes marshaling.

func (n *BadNode) MarshalYAML() (any, error)    { return wrap(n), nil }
func (n *Empty) MarshalYAML() (any, error)      { return wrap(n), nil }
func (n *Name) MarshalYAML() (any, error)       { return wrap(n), nil }
func (n *Type) MarshalYAML() (any, error)       { return wrap(n), nil }
func (n *Underscore) MarshalYAML() (any, error) { return wrap(n), nil }
func (n *Literal) MarshalYAML() (any, error)    { return wrap(n), nil }

func (n *Comment) MarshalYAML() (any, error)       { return wrap(n), nil }
func (n *CommentGroup) MarshalYAML() (any, error)  { return wrap(n), nil }
func (n *AttributeList) MarshalYAML() (any, error) { return wrap(n), nil }
func (n *LetDecl) MarshalYAML() (any, error)       { return wrap(n), nil }
func (n *TypeDecl) MarshalYAML() (any, error)      { return wrap(n), nil }
func (n *Decl) MarshalYAML() (any, error)          { return wrap(n), nil }

func (n *Label) MarshalYAML() (any, error)      { return wrap(n), nil }
func (n *ArrayType) MarshalYAML() (any, error)  { return wrap(n), nil }
func (n *StructType) MarshalYAML() (any, error) { return wrap(n), nil }
func (n *EnumType) MarshalYAML() (any, error)   { return wrap(n), nil }
func (n *Signature) MarshalYAML() (any, error)  { return wrap(n), nil }
func (n *BuiltIn) MarshalYAML() (any, error)    { return wrap(n), nil }
func (n *Call) MarshalYAML() (any, error)       { return wrap(n), nil }
func (n *Index) MarshalYAML() (any, error)      { return wrap(n), nil }
func (n *Function) MarshalYAML() (any, error)   { return wrap(n), nil }
func (n *Dot) MarshalYAML() (any, error)        { return wrap(n), nil }
func (n *Deref) MarshalYAML() (any, error)      { return wrap(n), nil }
func (n *Op) MarshalYAML() (any, error)         { return wrap(n), nil }

func (n *List) MarshalYAML() (any, error)   { return wrap(n), nil }
func (n *Block) MarshalYAML() (any, error)  { return wrap(n), nil }
func (n *Parens) MarshalYAML() (any, error) { return wrap(n), nil }
func (n *Stmts) MarshalYAML() (any, error)  { return wrap(n), nil }

func (n *If) MarshalYAML() (any, error)       { return wrap(n), nil }
func (n *Else) MarshalYAML() (any, error)     { return wrap(n), nil }
func (n *When) MarshalYAML() (any, error)     { return wrap(n), nil }
func (n *Defer) MarshalYAML() (any, error)    { return wrap(n), nil }
func (n *Return) MarshalYAML() (any, error)   { return wrap(n), nil }
func (n *Break) MarshalYAML() (any, error)    { return wrap(n), nil }
func (n *Continue) MarshalYAML() (any, error) { return wrap(n), nil }
func (n *Import) MarshalYAML() (any, error)   { return wrap(n), nil }

func wrap[T any](node *T) any {
	return struct {
		NodeKind string `yaml:"node_kind"`
		Node     T      `yaml:",inline"`
	}{typename[T](), *node}
}

func typename[T any]() string {
	name := reflect.TypeFor[T]().String()
	if idx := strings.IndexByte(name, '.'); idx >= 0 {
		name = name[idx+1:]
	}
	if idx := strings.IndexByte(name, '['); idx >= 0 {
		name = name[:idx]
	}
	return name
}
