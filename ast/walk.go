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
	case *BadNode, *Empty, *Name, *Type, *Underscore, *Literal, *Comment, *CommentGroup:
		// Nothing to walk

	case *AttributeList:
		assert(n.List != nil)

		walkList(n.List.List, visitor)

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
		assert(n.Type != nil)
		assert(n.Expr != nil)

		if n.Attrs != nil {
			WalkTopDown(n.Attrs, visitor)
		}

		WalkTopDown(n.Type, visitor)

		if n.Args != nil {
			walkList(n.Args.List, visitor)
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

	case *ArrayType:
		assert(n.X != nil)
		assert(n.Args != nil)

		WalkTopDown(n.X, visitor)
		walkList(n.Args.List, visitor)

	case *StructType:
		for _, field := range n.Fields {
			assert(field != nil)

			WalkTopDown(field, visitor)
		}

	case *EnumType:
		for _, field := range n.Fields {
			assert(field != nil)

			WalkTopDown(field, visitor)
		}

	case *Signature:
		assert(n.Params != nil)

		walkList(n.Params.List, visitor)

		if n.Result != nil {
			WalkTopDown(n.Result, visitor)
		}

	case *BuiltIn:
		assert(n.Name != nil)

		WalkTopDown(n.Name, visitor)

	case *Call:
		assert(n.X != nil)
		assert(n.Args != nil)

		WalkTopDown(n.X, visitor)
		walkList(n.Args.List, visitor)

	case *Index:
		assert(n.X != nil)
		assert(n.Args != nil)

		WalkTopDown(n.X, visitor)
		walkList(n.Args.List, visitor)

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

	case *Deref:
		assert(n.X != nil)

		WalkTopDown(n.X, visitor)

	case *Op:
		assert(n.X != nil)
		assert(n.Y != nil)

		if n.X != nil {
			WalkTopDown(n.X, visitor)
		}

		if n.Y != nil {
			WalkTopDown(n.Y, visitor)
		}

	case *List:
		walkList(n, visitor)

	case *StmtList:
		walkStmtList(n, visitor)

	case *BracketList:
		walkList(n.List, visitor)

	case *ParenList:
		walkList(n.List, visitor)

	case *CurlyList:
		walkStmtList(n.StmtList, visitor)

	case *If:
		assert(n.Cond != nil)
		assert(n.Body != nil)

		WalkTopDown(n.Cond, visitor)
		WalkTopDown(n.Body, visitor)

		if n.Else != nil {
			WalkTopDown(n.Else, visitor)
		}

	case *Else:
		assert(n.Body != nil)

		WalkTopDown(n.Body, visitor)

	case *While:
		assert(n.Cond != nil)
		assert(n.Body != nil)

		WalkTopDown(n.Cond, visitor)
		WalkTopDown(n.Body, visitor)

	case *For:
		assert(n.Decls != nil)
		assert(len(n.Decls.Nodes) > 0)
		assert(n.IterExpr != nil)
		assert(n.Body != nil)

		walkList(n.Decls, visitor)
		WalkTopDown(n.IterExpr, visitor)
		WalkTopDown(n.Body, visitor)

	case *When:
		assert(n.Expr != nil)
		assert(n.Body != nil)

		WalkTopDown(n.Expr, visitor)
		walkStmtList(n.Body.StmtList, visitor)

	case *Defer:
		assert(n.X != nil)

		WalkTopDown(n.X, visitor)

	case *Return:
		if n.X != nil {
			WalkTopDown(n.X, visitor)
		}

	case *Break:
		if n.Label != nil {
			WalkTopDown(n.Label, visitor)
		}

	case *Continue:
		if n.Label != nil {
			WalkTopDown(n.Label, visitor)
		}

	case *Import:
		assert(n.Module != nil)

		WalkTopDown(n.Module, visitor)

	default:
		// Should not happen.
		panic(fmt.Sprintf("unknown node type '%T'", n))
	}

	visitor.Visit(nil)
}

func walkList(list *List, visitor Visitor) {
	if list != nil {
		for _, node := range list.Nodes {
			WalkTopDown(node, visitor)
		}
	}
}

func walkStmtList(list *StmtList, visitor Visitor) {
	if list != nil {
		for _, node := range list.Nodes {
			WalkTopDown(node, visitor)
		}
	}
}

func assert(ok bool) {
	if !ok {
		panic("assertion failed")
	}
}
