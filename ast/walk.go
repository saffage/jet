package ast

import "fmt"

// Applies some action to a node.
//
// The result must be the next action to be performed on each
// of the child nodes. If no action is required for nodes of
// this branch, the result must be nil.
type Visitor interface {
	Visit(Node) Visitor
}

// Preorder\top-down traversal.
// Visit a parent node before visiting its children.
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
func WalkTopDown(tree Node, visitor Visitor) {
	if tree == nil {
		panic("can't walk a nil node")
	}

	if visitor = visitor.Visit(tree); visitor == nil {
		return
	}

	switch n := tree.(type) {
	case *BadNode, *Empty, *Name, *Type, *Underscore, *Literal:
		// Nothing to walk

	case *AttributeList:
		assert(n.List != nil)

		walkList(n.List.Nodes, visitor)

	case *LetDecl:
		assert(n.Decl.Name != nil)
		assert(n.Value != nil)

		if n.Attrs != nil {
			WalkTopDown(n.Attrs, visitor)
		}

		WalkTopDown(n.Decl.Name, visitor)

		if n.Decl.Type != nil {
			WalkTopDown(n.Decl.Type, visitor)
		}

		WalkTopDown(n.Value, visitor)

	case *TypeDecl:
		assert(n.Name != nil)
		assert(n.Expr != nil)

		if n.Attrs != nil {
			WalkTopDown(n.Attrs, visitor)
		}

		WalkTopDown(n.Name, visitor)

		if n.Args != nil {
			walkList(n.Args.Nodes, visitor)
		}

		WalkTopDown(n.Expr, visitor)

	case *Decl:
		assert(n.Name != nil)

		WalkTopDown(n.Name, visitor)

		if n.Type != nil {
			WalkTopDown(n.Type, visitor)
		}

	case *Label:
		assert(n.Label != nil)
		assert(n.X != nil)

		WalkTopDown(n.Label, visitor)
		WalkTopDown(n.X, visitor)

	case *Signature:
		assert(n.Params != nil)

		walkList(n.Params.Nodes, visitor)

		if n.Result != nil {
			WalkTopDown(n.Result, visitor)
		}

	case *Call:
		assert(n.X != nil)
		assert(n.Args != nil)

		WalkTopDown(n.X, visitor)
		walkList(n.Args.Nodes, visitor)

	case *Index:
		assert(n.X != nil)
		assert(n.Args != nil)

		WalkTopDown(n.X, visitor)
		walkList(n.Args.Nodes, visitor)

	case *Function:
		assert(n.Signature != nil)
		assert(n.Body != nil)

		WalkTopDown(n.Signature, visitor)
		WalkTopDown(n.Body, visitor)

	case *Dot:
		assert(n.X != nil)
		assert(n.Y != nil)

		WalkTopDown(n.X, visitor)
		WalkTopDown(n.Y, visitor)

	case *Op:
		assert(n.X != nil)
		assert(n.Y != nil)

		if n.X != nil {
			WalkTopDown(n.X, visitor)
		}

		if n.Y != nil {
			WalkTopDown(n.Y, visitor)
		}

	case *Stmts:
		walkList(n.Nodes, visitor)

	case *Block:
		walkList(n.Stmts.Nodes, visitor)

	case *List:
		walkList(n.Nodes, visitor)

	case *Parens:
		walkList(n.Nodes, visitor)

	case *When:
		assert(n.Expr != nil)
		assert(n.Body != nil)

		WalkTopDown(n.Expr, visitor)
		walkList(n.Body.Stmts.Nodes, visitor)

	default:
		// Should not happen.
		panic(fmt.Sprintf("unknown node type '%T'", n))
	}

	visitor.Visit(nil)
}

func walkList(nodes []Node, visitor Visitor) {
	for _, node := range nodes {
		WalkTopDown(node, visitor)
	}
}

func assert(ok bool) {
	if !ok {
		panic("assertion failed")
	}
}
