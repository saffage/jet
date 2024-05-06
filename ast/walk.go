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

	if v := visit(tree); v != nil {
		visit = v
	} else {
		return
	}

	switch n := tree.(type) {
	case *BadNode, *Empty, *Ident, *Literal, *Operator, *Comment, *CommentGroup:
		// Nothing to walk

	case *BindingWithValue:
		assert.Ok(n.Binding.Name != nil)
		assert.Ok(n.Binding.Type != nil || (n.Value != nil && n.Operator != nil))

		WalkTopDown(visit, n.Binding.Name)

		if n.Type != nil {
			WalkTopDown(visit, n.Binding.Type)
		}

		if n.Value != nil {
			WalkTopDown(visit, n.Operator)
			WalkTopDown(visit, n.Value)
		}

	case *Signature:
		assert.Ok(n.Params != nil)

		walkExprList(visit, n.Params.ExprList)

		if n.Result != nil {
			WalkTopDown(visit, n.Result)
		}

	case *Call:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		WalkTopDown(visit, n.X)
		walkExprList(visit, n.Args.ExprList)

	case *Index:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		WalkTopDown(visit, n.X)
		walkExprList(visit, n.Args.ExprList)

	case *ArrayType:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		WalkTopDown(visit, n.X)
		walkExprList(visit, n.Args.ExprList)

	case *MemberAccess:
		assert.Ok(n.X != nil)
		assert.Ok(n.Selector != nil)

		WalkTopDown(visit, n.X)
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

	case *ExprList:
		walkExprList(visit, n)

	case *BracketList:
		walkExprList(visit, n.ExprList)

	case *ParenList:
		walkExprList(visit, n.ExprList)

	case *CurlyList:
		walkList(visit, n.List)

	case *AttributeList:
		assert.Ok(n.List != nil)

		walkExprList(visit, n.List.ExprList)

	case *BuiltInCall:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Args != nil)

		WalkTopDown(visit, n.Name)
		WalkTopDown(visit, n.Args)

	case *ModuleDecl:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Body != nil)

		if n.Attrs != nil {
			WalkTopDown(visit, n.Attrs)
		}

		WalkTopDown(visit, n.Name)

		switch b := n.Body.(type) {
		case *List:
			walkList(visit, b)

		case *ExprList:
			walkExprList(visit, b)

		case *CurlyList:
			walkList(visit, b.List)

		default:
			panic(fmt.Sprintf("unexpected node type '%T' for module body", n.Body))
		}

	case *VarDecl:
		assert.Ok(n.Binding.Name != nil)
		assert.Ok(n.Binding.Type != nil || n.Value != nil)

		if n.Attrs != nil {
			WalkTopDown(visit, n.Attrs)
		}

		WalkTopDown(visit, n.Binding.Name)

		if n.Binding.Type != nil {
			WalkTopDown(visit, n.Binding.Type)
		}

		if n.Value != nil {
			WalkTopDown(visit, n.Binding.Type)
		}

	case *FuncDecl:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Signature != nil)

		if n.Attrs != nil {
			WalkTopDown(visit, n.Attrs)
		}

		WalkTopDown(visit, n.Name)
		WalkTopDown(visit, n.Signature)

		if n.Body != nil {
			WalkTopDown(visit, n.Body)
		}

	case *TypeAliasDecl:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Expr != nil)

		if n.Attrs != nil {
			WalkTopDown(visit, n.Attrs)
		}

		WalkTopDown(visit, n.Name)
		WalkTopDown(visit, n.Expr)

	case *If:
		assert.Ok(n.Cond != nil)
		assert.Ok(n.Body != nil)

		WalkTopDown(visit, n.Cond)
		WalkTopDown(visit, n.Body)

		if n.Else != nil {
			WalkTopDown(visit, n.Else)
		}

	case *Else:
		assert.Ok(n.Body != nil)

		WalkTopDown(visit, n.Body)

	case *While:
		assert.Ok(n.Cond != nil)
		assert.Ok(n.Body != nil)

		WalkTopDown(visit, n.Cond)
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

func walkExprList(visit Visitor, list *ExprList) {
	if list != nil {
		for _, node := range list.Exprs {
			WalkTopDown(visit, node)
		}
	}
}
