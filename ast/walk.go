package ast

import (
	"fmt"

	"github.com/saffage/jet/internal/assert"
)

// Applies some action to a node.
//
// The result must be the next action to be performed on each
// of the child nodes. If no action is required and this branch
// should be dropped, nil should be returned.
type Visitor func(Node) Visitor

// Preorder\top-down traversal.
// Visit a parent node before visiting its children.
// Each node is terminated by a call with `nil` argument.
//
// Example:
//   - List (length = 3)
//   - - Ident
//   - - nil (Ident)
//   - - Ident
//   - - nil (Ident)
//   - - List (length = 0)
//   - - nil (List)
//   - nil (List)
func WalkTopDown(visit Visitor, tree Node) {
	if tree == nil {
		panic("can't walk a nil node")
	}

	visit = visit(tree)

	if visit == nil {
		return
	}

	switch n := tree.(type) {
	case *BadNode, *Empty, *Ident, *Literal, *PrefixOpr, *InfixOpr, *PostfixOpr, *Star, *Comment, *CommentGroup:
		// Nothing to walk

	case *Field:
		for _, name := range n.Names {
			assert.Ok(name != nil)
			WalkTopDown(visit, name)
		}

		if n.Type != nil {
			WalkTopDown(visit, n.Type)
		}

		if n.Value != nil {
			WalkTopDown(visit, n.Value)
		}

	case *Signature:
		walkList(visit, n.Params.List)

		if n.Result != nil {
			WalkTopDown(visit, n.Result)
		}

	case *Call:
		assert.Ok(n.X != nil)
		WalkTopDown(visit, n.X)
		walkList(visit, n.Args.List)

	case *Index:
		assert.Ok(n.X != nil)
		WalkTopDown(visit, n.X)

	case *ArrayType:
		assert.Ok(n.X != nil)
		WalkTopDown(visit, n.X)
		walkList(visit, n.Args.List)

	case *MemberAccess:
		assert.Ok(n.X != nil)
		WalkTopDown(visit, n.X)

		assert.Ok(n.Selector != nil)
		WalkTopDown(visit, n.Selector)

	case *PrefixOp:
		assert.Ok(n.X != nil)
		WalkTopDown(visit, n.X)

	case *InfixOp:
		assert.Ok(n.X != nil)
		assert.Ok(n.Y != nil)
		WalkTopDown(visit, n.X)
		WalkTopDown(visit, n.Y)

	case *PostfixOp:
		assert.Ok(n.X != nil)
		WalkTopDown(visit, n.X)

	case *List:
		walkList(visit, n)

	case *ParenList:
		walkList(visit, n.List)

	case *CurlyList:
		walkList(visit, n.List)

	case *BracketList:
		walkList(visit, n.List)

	case *AttributeList:
		assert.Ok(n.List != nil)
		walkList(visit, n.List.List)

	case *BuiltInCall:
		assert.Ok(n.Name != nil)
		WalkTopDown(visit, n.Name)
		WalkTopDown(visit, n.X)

	case *ModuleDecl:
		if n.Attrs != nil {
			WalkTopDown(visit, n.Attrs)
		}

		assert.Ok(n.Name != nil)
		WalkTopDown(visit, n.Name)

		switch b := n.Body.(type) {
		case *List:
			walkList(visit, b)

		case *CurlyList:
			walkList(visit, b.List)

		default:
			panic(fmt.Sprintf("unexpected node type '%T' for module body", n.Body))
		}

	case *GenericDecl:
		if n.Attrs != nil {
			WalkTopDown(visit, n.Attrs)
		}

		assert.Ok(n.Field != nil)
		WalkTopDown(visit, n.Field)

	case *FuncDecl:
		if n.Attrs != nil {
			WalkTopDown(visit, n.Attrs)
		}

		assert.Ok(n.Name != nil)
		WalkTopDown(visit, n.Name)

		assert.Ok(n.Signature != nil)
		WalkTopDown(visit, n.Signature)

		if n.Body != nil {
			WalkTopDown(visit, n.Body)
		}

	case *TypeAliasDecl:
		if n.Attrs != nil {
			WalkTopDown(visit, n.Attrs)
		}

		assert.Ok(n.Name != nil)
		WalkTopDown(visit, n.Name)

		assert.Ok(n.Expr != nil)
		WalkTopDown(visit, n.Expr)

	case *If:
		assert.Ok(n.Cond != nil)
		WalkTopDown(visit, n.Cond)

		assert.Ok(n.Body != nil)
		WalkTopDown(visit, n.Body)

		if n.Else != nil {
			WalkTopDown(visit, n.Else)
		}

	case *Else:
		assert.Ok(n.Body != nil)
		WalkTopDown(visit, n.Body)

	case *While:
		assert.Ok(n.Cond != nil)
		WalkTopDown(visit, n.Cond)

		assert.Ok(n.Body != nil)
		WalkTopDown(visit, n.Body)

	case *Return:
		if n.X != nil {
			WalkTopDown(visit, n.X)
		}

	case *Break:
		if n.Label != nil {
			WalkTopDown(visit, n.Label)
		}

	case *Continue:
		if n.Label != nil {
			WalkTopDown(visit, n.Label)
		}

	default:
		// Should not happen.
		panic(fmt.Sprintf("unknown node type '%T'", n))
	}

	visit(nil)
}

func walkList(visit Visitor, list *List) {
	if list != nil {
		for _, node := range list.Nodes {
			WalkTopDown(visit, node)
		}
	}
}
