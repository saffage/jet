package ast

import (
	"fmt"

	"github.com/saffage/jet/internal/assert"
)

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

	case *Decl:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Type != nil || n.Value != nil)

		if n.Attrs != nil {
			v.WalkTopDown(n.Attrs)
		}

		v.WalkTopDown(n.Name)

		if n.Type != nil {
			v.WalkTopDown(n.Type)
		}

		if n.Value != nil {
			v.WalkTopDown(n.Value)
		}

	// case *Binding:
	// 	assert.Ok(n.Name != nil)
	// 	assert.Ok(n.Type != nil)

	// 	v.WalkTopDown(visit, n.Name)

	// 	if n.Type != nil {
	// 		v.WalkTopDown(visit, n.Type)
	// 	}

	// case *BindingWithValue:
	// 	assert.Ok(n.Binding.Name != nil)
	// 	assert.Ok(n.Binding.Type != nil || (n.Value != nil && n.Operator != nil))

	// 	v.WalkTopDown(visit, n.Binding.Name)

	// 	if n.Type != nil {
	// 		v.WalkTopDown(visit, n.Binding.Type)
	// 	}

	// 	if n.Value != nil {
	// 		v.WalkTopDown(visit, n.Operator)
	// 		v.WalkTopDown(visit, n.Value)
	// 	}

	case *BuiltInCall:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Args != nil)

		v.WalkTopDown(n.Name)
		v.WalkTopDown(n.Args)

	case *Call:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		v.WalkTopDown(n.X)
		v.walkList(n.Args.List)

	case *Index:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		v.WalkTopDown(n.X)
		v.walkList(n.Args.List)

	case *ArrayType:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		v.WalkTopDown(n.X)
		v.walkList(n.Args.List)

	case *StructType:
		for _, field := range n.Fields {
			assert.Ok(field != nil)

			v.WalkTopDown(field)
		}

	case *EnumType:
		for _, field := range n.Fields {
			assert.Ok(field != nil)

			v.WalkTopDown(field)
		}

	case *Signature:
		assert.Ok(n.Params != nil)

		v.walkList(n.Params.List)

		if n.Result != nil {
			v.WalkTopDown(n.Result)
		}

	case *Dot:
		assert.Ok(n.X != nil)
		assert.Ok(n.Y != nil)

		v.WalkTopDown(n.X)
		v.WalkTopDown(n.Y)

	case *Deref:
		assert.Ok(n.X != nil)

		v.WalkTopDown(n.X)

	// case *SafeMemberAccess:
	// 	assert.Ok(n.X != nil)
	// 	assert.Ok(n.Selector != nil)

	// 	v.WalkTopDown(visit, n.X)
	// 	v.WalkTopDown(visit, n.Selector)

	// case *PrefixOp:
	// 	assert.Ok(n.X != nil)

	// 	v.WalkTopDown(visit, n.X)

	case *Op:
		assert.Ok(n.X != nil)
		assert.Ok(n.Y != nil)

		if n.X != nil {
			v.WalkTopDown(n.X)
		}

		if n.Y != nil {
			v.WalkTopDown(n.Y)
		}

	case *BracketList:
		v.walkList(n.List)

	case *ParenList:
		v.walkList(n.List)

	case *CurlyList:
		v.walkStmtList(n.StmtList)

	case *If:
		assert.Ok(n.Cond != nil)
		assert.Ok(n.Body != nil)

		v.WalkTopDown(n.Cond)
		v.WalkTopDown(n.Body)

		if n.Else != nil {
			v.WalkTopDown(n.Else)
		}

	case *Else:
		assert.Ok(n.Body != nil)

		v.WalkTopDown(n.Body)

	case *StmtList:
		v.walkStmtList(n)

	case *List:
		v.walkList(n)

	case *AttributeList:
		assert.Ok(n.List != nil)

		v.walkList(n.List.List)

	case *While:
		assert.Ok(n.Cond != nil)
		assert.Ok(n.Body != nil)

		v.WalkTopDown(n.Cond)
		v.WalkTopDown(n.Body)

	case *For:
		assert.Ok(n.DeclList != nil)
		assert.Ok(len(n.DeclList.Nodes) > 0)
		assert.Ok(n.IterExpr != nil)
		assert.Ok(n.Body != nil)

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
		assert.Ok(n.Module != nil)

		v.WalkTopDown(n.Module)

	default:
		// Should not happen.
		panic(fmt.Sprintf("unknown node type '%T'", n))
	}

	v(nil)
}

func (v Visitor) walkStmtList(list *StmtList) {
	if list != nil {
		for _, node := range list.Nodes {
			v.WalkTopDown(node)
		}
	}
}

func (v Visitor) walkList(list *List) {
	if list != nil {
		for _, node := range list.Nodes {
			v.WalkTopDown(node)
		}
	}
}
