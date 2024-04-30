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
			visit(name)
		}

		if n.Type != nil {
			visit(n.Type)
		}

		if n.Value != nil {
			visit(n.Value)
		}

	case *Signature:
		walkList(visit, n.Params.List)

		if n.Result != nil {
			visit(n.Result)
		}

	case *Call:
		assert.Ok(n.X != nil)
		visit(n.X)
		walkList(visit, n.Args.List)

	case *Index:
		assert.Ok(n.X != nil)
		visit(n.X)

	case *ArrayType:
		assert.Ok(n.X != nil)
		visit(n.X)
		walkList(visit, n.Args.List)

	case *MemberAccess:
		assert.Ok(n.X != nil)
		visit(n.X)

		assert.Ok(n.Selector != nil)
		visit(n.Selector)

	case *PrefixOp:
		assert.Ok(n.X != nil)
		visit(n.X)

	case *InfixOp:
		assert.Ok(n.X != nil)
		assert.Ok(n.Y != nil)
		visit(n.X)
		visit(n.Y)

	case *PostfixOp:
		assert.Ok(n.X != nil)
		visit(n.X)

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
		visit(n.Name)
		visit(n.X)

	case *ModuleDecl:
		if n.Attrs != nil {
			visit(n.Attrs)
		}

		assert.Ok(n.Name != nil)
		visit(n.Name)

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
			visit(n.Attrs)
		}

		assert.Ok(n.Field != nil)
		visit(n.Field)

	case *FuncDecl:
		if n.Attrs != nil {
			visit(n.Attrs)
		}

		assert.Ok(n.Name != nil)
		visit(n.Name)

		assert.Ok(n.Signature != nil)
		visit(n.Signature)

		if n.Body != nil {
			visit(n.Body)
		}

	case *TypeAliasDecl:
		if n.Attrs != nil {
			visit(n.Attrs)
		}

		assert.Ok(n.Name != nil)
		visit(n.Name)

		assert.Ok(n.Expr != nil)
		visit(n.Expr)

	case *If:
		assert.Ok(n.Cond != nil)
		visit(n.Cond)

		assert.Ok(n.Body != nil)
		visit(n.Body)

		if n.Else != nil {
			visit(n.Else)
		}

	case *Else:
		assert.Ok(n.Body != nil)
		visit(n.Body)

	case *While:
		assert.Ok(n.Cond != nil)
		visit(n.Cond)

		assert.Ok(n.Body != nil)
		visit(n.Body)

	case *Return:
		if n.X != nil {
			visit(n.X)
		}

	case *Break:
		if n.Label != nil {
			visit(n.Label)
		}

	case *Continue:
		if n.Label != nil {
			visit(n.Label)
		}

	default:
		// NOTE should not happen.
		panic(fmt.Sprintf("unknown node type '%T'", n))
	}

	visit(nil)
}

func walkList(visit Visitor, list *List) {
	for _, node := range list.Nodes {
		visit(node)
	}
}
