package ast

// TODO: fix a bug when 'UnmarshalYAML' method brokes marshaling.

func (n *BadNode) MarshalYAML() (any, error)    { return wrap(n, "BadNode"), nil }
func (n *Empty) MarshalYAML() (any, error)      { return wrap(n, "Empty"), nil }
func (n *Name) MarshalYAML() (any, error)       { return wrap(n, "Name"), nil }
func (n *Type) MarshalYAML() (any, error)       { return wrap(n, "Type"), nil }
func (n *Underscore) MarshalYAML() (any, error) { return wrap(n, "Underscore"), nil }
func (n *Literal) MarshalYAML() (any, error)    { return wrap(n, "Literal"), nil }

func (n *Comment) MarshalYAML() (any, error)       { return wrap(n, "Comment"), nil }
func (n *CommentGroup) MarshalYAML() (any, error)  { return wrap(n, "CommentGroup"), nil }
func (n *AttributeList) MarshalYAML() (any, error) { return wrap(n, "AttributeList"), nil }
func (n *LetDecl) MarshalYAML() (any, error)       { return wrap(n, "LetDecl"), nil }
func (n *TypeDecl) MarshalYAML() (any, error)      { return wrap(n, "TypeDecl"), nil }
func (n *Decl) MarshalYAML() (any, error)          { return wrap(n, "Decl"), nil }

func (n *Label) MarshalYAML() (any, error)      { return wrap(n, "Label"), nil }
func (n *ArrayType) MarshalYAML() (any, error)  { return wrap(n, "ArrayType"), nil }
func (n *StructType) MarshalYAML() (any, error) { return wrap(n, "StructType"), nil }
func (n *EnumType) MarshalYAML() (any, error)   { return wrap(n, "EnumType"), nil }
func (n *Signature) MarshalYAML() (any, error)  { return wrap(n, "Signature"), nil }
func (n *BuiltIn) MarshalYAML() (any, error)    { return wrap(n, "BuiltIn"), nil }
func (n *Call) MarshalYAML() (any, error)       { return wrap(n, "Call"), nil }
func (n *Index) MarshalYAML() (any, error)      { return wrap(n, "Index"), nil }
func (n *Function) MarshalYAML() (any, error)   { return wrap(n, "Function"), nil }
func (n *Dot) MarshalYAML() (any, error)        { return wrap(n, "Dot"), nil }
func (n *Op) MarshalYAML() (any, error)         { return wrap(n, "Op"), nil }

func (n *List) MarshalYAML() (any, error)        { return wrap(n, "List"), nil }
func (n *StmtList) MarshalYAML() (any, error)    { return wrap(n, "StmtList"), nil }
func (n *BracketList) MarshalYAML() (any, error) { return wrap(n, "BracketList"), nil }
func (n *ParenList) MarshalYAML() (any, error)   { return wrap(n, "ParenList"), nil }
func (n *CurlyList) MarshalYAML() (any, error)   { return wrap(n, "CurlyList"), nil }

func (n *If) MarshalYAML() (any, error)       { return wrap(n, "If"), nil }
func (n *Else) MarshalYAML() (any, error)     { return wrap(n, "Else"), nil }
func (n *While) MarshalYAML() (any, error)    { return wrap(n, "While"), nil }
func (n *For) MarshalYAML() (any, error)      { return wrap(n, "For"), nil }
func (n *When) MarshalYAML() (any, error)     { return wrap(n, "When"), nil }
func (n *Defer) MarshalYAML() (any, error)    { return wrap(n, "Defer"), nil }
func (n *Return) MarshalYAML() (any, error)   { return wrap(n, "Return"), nil }
func (n *Break) MarshalYAML() (any, error)    { return wrap(n, "Break"), nil }
func (n *Continue) MarshalYAML() (any, error) { return wrap(n, "Continue"), nil }
func (n *Import) MarshalYAML() (any, error)   { return wrap(n, "Import"), nil }

func wrap[T any](node *T, typename string) any {
	return struct {
		NodeKind string `yaml:"node_kind"`
		Node     T      `yaml:",inline"`
	}{typename, *node}
}
