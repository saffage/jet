package ast

import (
	"reflect"
	"strings"
)

// TODO: fix a bug when 'UnmarshalYAML' method broke marshaling.

func (node *BadNode) MarshalYAML() (any, error)    { return wrap(node), nil }
func (node *Empty) MarshalYAML() (any, error)      { return wrap(node), nil }
func (node *Lower) MarshalYAML() (any, error)      { return wrap(node), nil }
func (node *Upper) MarshalYAML() (any, error)      { return wrap(node), nil }
func (node *TypeVar) MarshalYAML() (any, error)    { return wrap(node), nil }
func (node *Underscore) MarshalYAML() (any, error) { return wrap(node), nil }
func (node *Literal) MarshalYAML() (any, error)    { return wrap(node), nil }

func (node *AttributeList) MarshalYAML() (any, error) { return wrap(node), nil }
func (node *LetDecl) MarshalYAML() (any, error)       { return wrap(node), nil }
func (node *TypeDecl) MarshalYAML() (any, error)      { return wrap(node), nil }
func (node *Decl) MarshalYAML() (any, error)          { return wrap(node), nil }
func (node *Variant) MarshalYAML() (any, error)       { return wrap(node), nil }

func (node *Label) MarshalYAML() (any, error)      { return wrap(node), nil }
func (node *Signature) MarshalYAML() (any, error)  { return wrap(node), nil }
func (node *Call) MarshalYAML() (any, error)       { return wrap(node), nil }
func (node *Dot) MarshalYAML() (any, error)        { return wrap(node), nil }
func (node *Op) MarshalYAML() (any, error)         { return wrap(node), nil }

func (node *List) MarshalYAML() (any, error)   { return wrap(node), nil }
func (node *Block) MarshalYAML() (any, error)  { return wrap(node), nil }
func (node *Parens) MarshalYAML() (any, error) { return wrap(node), nil }
func (node *Stmts) MarshalYAML() (any, error)  { return wrap(node), nil }

func (node *When) MarshalYAML() (any, error)     { return wrap(node), nil }
func (node *Extern) MarshalYAML() (any, error)   { return wrap(node), nil }

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
