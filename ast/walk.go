package ast

// Applies some action to a node.
//
// The result must be the next action to be performed on each
// of the child nodes. If no action is required for nodes of
// this branch, the result must be nil.
type Visitor interface {
	Visit(Node) Visitor
}

// Top-down traversal. Visit a parent node before visiting its children.
//
// Each node is terminated by a call with 'nil' argument.
//
// Example:
//   - List(len: 3)
//   - - Ident
//   - - Ident(nil)
//   - - Ident
//   - - Ident(nil)
//   - - List(len: 0)
//   - - List(nil)
//   - List(nil)
func WalkTopDown(n Node, v Visitor) {
	if n == nil {
		panic("can't walk a nil node")
	}
	if v = v.Visit(n); v == nil {
		return
	}
	n.walk(v)
	v.Visit(nil)
}

func walkList(nodes []Node, v Visitor) {
	for _, node := range nodes {
		WalkTopDown(node, v)
	}
}

func assert(ok bool) {
	if !ok {
		panic("assertion failed")
	}
}

type walkable interface {
	walk(Visitor)
}

func (node *BadNode) walk(Visitor)    {}
func (node *Empty) walk(Visitor)      {}
func (node *Lower) walk(Visitor)      {}
func (node *Upper) walk(Visitor)      {}
func (node *TypeVar) walk(Visitor)    {}
func (node *Underscore) walk(Visitor) {}
func (node *Literal) walk(Visitor)    {}

func (node *AttributeList) walk(v Visitor) {
	assert(node.List != nil)

	walkList(node.List.Nodes, v)
}

func (node *LetDecl) walk(v Visitor) {
	assert(node.Decl.Name != nil)
	assert(node.Value != nil)

	if node.Attrs != nil {
		WalkTopDown(node.Attrs, v)
	}

	WalkTopDown(node.Decl.Name, v)

	if node.Decl.Type != nil {
		WalkTopDown(node.Decl.Type, v)
	}

	WalkTopDown(node.Value, v)
}

func (node *TypeDecl) walk(v Visitor) {
	assert(node.Name != nil)
	assert(node.Expr != nil)

	if node.Attrs != nil {
		WalkTopDown(node.Attrs, v)
	}

	WalkTopDown(node.Name, v)

	if node.Args != nil {
		walkList(node.Args.Nodes, v)
	}

	WalkTopDown(node.Expr, v)
}

func (node *Decl) walk(v Visitor) {
	assert(node.Name != nil)

	WalkTopDown(node.Name, v)

	if node.Type != nil {
		WalkTopDown(node.Type, v)
	}
}

func (node *Variant) walk(v Visitor) {
	assert(node.Name != nil)

	WalkTopDown(node.Name, v)

	if node.Params != nil {
		walkList(node.Params.Nodes, v)
	}
}

func (node *Label) walk(v Visitor) {
	assert(node.Name != nil)
	assert(node.X != nil)

	WalkTopDown(node.Name, v)
	WalkTopDown(node.X, v)
}

func (node *Signature) walk(v Visitor) {
	assert(node.Params != nil)

	walkList(node.Params.Nodes, v)

	if node.Result != nil {
		WalkTopDown(node.Result, v)
	}
}

func (node *Call) walk(v Visitor) {
	assert(node.X != nil)
	assert(node.Args != nil)

	WalkTopDown(node.X, v)
	walkList(node.Args.Nodes, v)
}

func (node *Dot) walk(v Visitor) {
	assert(node.X != nil)
	assert(node.Y != nil)

	WalkTopDown(node.X, v)
	WalkTopDown(node.Y, v)
}

func (node *Op) walk(v Visitor) {
	assert(node.X != nil)
	assert(node.Y != nil)

	if node.X != nil {
		WalkTopDown(node.X, v)
	}

	if node.Y != nil {
		WalkTopDown(node.Y, v)
	}
}

func (node *Stmts) walk(v Visitor) {
	walkList(node.Nodes, v)
}

func (node *Block) walk(v Visitor) {
	walkList(node.Stmts.Nodes, v)
}

func (node *List) walk(v Visitor) {
	walkList(node.Nodes, v)
}

func (node *Parens) walk(v Visitor) {
	walkList(node.Nodes, v)
}

func (node *When) walk(v Visitor) {
	assert(node.Expr != nil)
	assert(node.Body != nil)

	WalkTopDown(node.Expr, v)
	walkList(node.Body.Stmts.Nodes, v)
}

func (node *Extern) walk(v Visitor) {
	if node.Args != nil {
		walkList(node.Args.Nodes, v)
	}
}
