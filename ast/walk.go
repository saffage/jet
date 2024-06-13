package ast

import "fmt"

// Applies some action to a node.
//
// The result must be the next action to be performed on each
// of the child nodes. If no action is required for nodes of
// this branch, the result must be nil.
type Visitor func(Node) Visitor

// Preorder\top-down traversal.
// Visit a parent node before visiting its children.
// Each node is terminated by a call with `nil` argument.
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
func (v Visitor) WalkTopDown(tree Node) {
	if tree == nil {
		panic("can't walk a nil node")
	}

	if visitor := v(tree); visitor != nil {
		v = visitor
	} else {
		return
	}

	switch n := tree.(type) {
	case *BadNode, *Empty, *Ident, *Literal, *Comment, *CommentGroup:
		// Nothing to walk

	case *AttributeList:
		assert(n.List != nil)

		v.walkList(n.List.List)

	case *Decl:
		assert(n.Ident != nil)
		assert(n.Type != nil || n.Value != nil)

		if n.Attrs != nil {
			v.WalkTopDown(n.Attrs)
		}

		v.WalkTopDown(n.Ident)

		if n.Type != nil {
			v.WalkTopDown(n.Type)
		}

		if n.Value != nil {
			v.WalkTopDown(n.Value)
		}

	case *ArrayType:
		assert(n.X != nil)
		assert(n.Args != nil)

		v.WalkTopDown(n.X)
		v.walkList(n.Args.List)

	case *StructType:
		for _, field := range n.Fields {
			assert(field != nil)

			v.WalkTopDown(field)
		}

	case *EnumType:
		for _, field := range n.Fields {
			assert(field != nil)

			v.WalkTopDown(field)
		}

	case *Signature:
		assert(n.Params != nil)

		v.walkList(n.Params.List)

		if n.Result != nil {
			v.WalkTopDown(n.Result)
		}

	case *BuiltIn:
		assert(n.Ident != nil)

		v.WalkTopDown(n.Ident)

	case *Call:
		assert(n.X != nil)
		assert(n.Args != nil)

		v.WalkTopDown(n.X)
		v.walkList(n.Args.List)

	case *Index:
		assert(n.X != nil)
		assert(n.Args != nil)

		v.WalkTopDown(n.X)
		v.walkList(n.Args.List)

	case *Function:
		assert(n.Signature != nil)
		assert(n.Body != nil)

		v.WalkTopDown(n.Signature)
		v.WalkTopDown(n.Body)

	case *Dot:
		assert(n.X != nil)
		assert(n.Y != nil)

		v.WalkTopDown(n.X)
		v.WalkTopDown(n.Y)

	case *Deref:
		assert(n.X != nil)

		v.WalkTopDown(n.X)

	case *Op:
		assert(n.X != nil)
		assert(n.Y != nil)

		if n.X != nil {
			v.WalkTopDown(n.X)
		}

		if n.Y != nil {
			v.WalkTopDown(n.Y)
		}

	case *List:
		v.walkList(n)

	case *StmtList:
		v.walkStmtList(n)

	case *BracketList:
		v.walkList(n.List)

	case *ParenList:
		v.walkList(n.List)

	case *CurlyList:
		v.walkStmtList(n.StmtList)

	case *If:
		assert(n.Cond != nil)
		assert(n.Body != nil)

		v.WalkTopDown(n.Cond)
		v.WalkTopDown(n.Body)

		if n.Else != nil {
			v.WalkTopDown(n.Else)
		}

	case *Else:
		assert(n.Body != nil)

		v.WalkTopDown(n.Body)

	case *While:
		assert(n.Cond != nil)
		assert(n.Body != nil)

		v.WalkTopDown(n.Cond)
		v.WalkTopDown(n.Body)

	case *For:
		assert(n.DeclList != nil)
		assert(len(n.DeclList.Nodes) > 0)
		assert(n.IterExpr != nil)
		assert(n.Body != nil)

		v.walkList(n.DeclList)
		v.WalkTopDown(n.IterExpr)
		v.WalkTopDown(n.Body)

	case *Return:
		if n.X != nil {
			v.WalkTopDown(n.X)
		}

	case *Break:
		if n.Label != nil {
			v.WalkTopDown(n.Label)
		}

	case *Continue:
		if n.Label != nil {
			v.WalkTopDown(n.Label)
		}

	case *Import:
		assert(n.Module != nil)

		v.WalkTopDown(n.Module)

	default:
		// Should not happen.
		panic(fmt.Sprintf("unknown node type '%T'", n))
	}

	v(nil)
}

func (v Visitor) walkList(list *List) {
	if list != nil {
		for _, node := range list.Nodes {
			v.WalkTopDown(node)
		}
	}
}

func (v Visitor) walkStmtList(list *StmtList) {
	if list != nil {
		for _, node := range list.Nodes {
			v.WalkTopDown(node)
		}
	}
}

func assert(ok bool, message ...any) {
	if !ok {
		if len(message) > 0 {
			panic("assertion failed: " + fmt.Sprint(message...))
		}
		panic("assertion failed")
	}
}
