package ast

import (
	"fmt"

	"github.com/saffage/jet/internal/assert"
)

type Visitor interface {
	// Applies some action to `node`. The result is the next action
	// to be performed on each of the children, or nil if no action
	// is needed and this branch should be dropped.
	Visit(node Node) Visitor
}

type Walker interface {
	// Preorder traversal or top-down traversal.
	// Visit a parent node before visiting its children.
	// Each node is terminated by a call with argument nil.
	//
	// Example:
	// 	- List (length = 3)
	//  - - Ident
	//  - - nil (Ident)
	//  - - Ident
	//  - - nil (Ident)
	//  - - List (length = 0)
	//  - - nil (List)
	//  - nil (List)
	Walk(node Node)
}

type stackEntry struct {
	Node
	Visitor
}

type defaultWalker struct {
	visitor Visitor
	stack   []stackEntry
}

func NewWalker(visitor Visitor) Walker {
	return &defaultWalker{visitor, []stackEntry{}}
}

func (w *defaultWalker) Walk(node Node) {
	if visitor := w.visitor.Visit(node); visitor == nil {
		return
	}

	if node == nil {
		return
	}

	w.stack = append(w.stack, stackEntry{node, w.visitor})

	switch n := node.(type) {
	case *BadNode, *Empty, *Ident, *Literal:
		// Nothing to walk

	case *ParenExpr:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *Ellipsis:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *Ref:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *ArrayType:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

		if n.N != nil {
			w.Walk(n.N)
		}

	case *Signature:
		walkList(w, n.Params.List)

		if n.Result != nil {
			w.Walk(n.Result)
		}

	case *MemberAccess:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

		assert.Ok(n.Y != nil)
		w.Walk(n.Y)

	case *Star:
		// Nothing to walk

	case *Try:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *Unwrap:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *Call:
		assert.Ok(n.X != nil)
		w.Walk(n.X)
		walkList(w, n.Args.List)

	case *Index:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *List:
		walkList(w, n)

	case *ParenList:
		walkList(w, n.List)

	case *CurlyList:
		walkList(w, n.List)

	case *BracketList:
		walkList(w, n.List)

	case *Field:
		for _, name := range n.Names {
			assert.Ok(name != nil)
			w.Walk(name)
		}

		if n.Type != nil {
			w.Walk(n.Type)
		}

		if n.Value != nil {
			w.Walk(n.Value)
		}

	case *UnaryOp:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *BinaryOp:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *ModuleDecl:
		assert.Ok(n.Name != nil)
		w.Walk(n.Name)

		switch b := n.Body.(type) {
		case *List:
			walkList(w, b)

		case *CurlyList:
			walkList(w, b.List)

		default:
			panic(fmt.Sprintf("ast.(*defaultWalker).Walk: unexpected node type '%T' for ModuleDecl.Body field", n.Body))
		}

	case *GenericDecl:
		assert.Ok(n.Field != nil)
		w.Walk(n.Field)

	case *FuncDecl:
		assert.Ok(n.Name != nil)
		w.Walk(n.Name)

		assert.Ok(n.Signature != nil)
		w.Walk(n.Signature)

		if n.Body != nil {
			w.Walk(n.Body)
		}

	case *StructDecl:
		assert.Ok(n.Name != nil)
		w.Walk(n.Name)
		walkList(w, n.Fields.List)

	case *EnumDecl:
		assert.Ok(n.Name != nil)
		w.Walk(n.Name)
		walkList(w, n.Body.List)

	case *AliasDecl:
		assert.Ok(n.Name != nil)
		w.Walk(n.Name)

		assert.Ok(n.Expr != nil)
		w.Walk(n.Expr)

	case *If:
		assert.Ok(n.Cond != nil)
		w.Walk(n.Cond)

		assert.Ok(n.Body != nil)
		w.Walk(n.Body)

		if n.Else != nil {
			assert.Ok(n.Else != nil)
			w.Walk(n.Else)
		}

	case *Else:
		assert.Ok(n.Body != nil)
		w.Walk(n.Body)

	case *While:
		assert.Ok(n.Cond != nil)
		w.Walk(n.Cond)

		assert.Ok(n.Body != nil)
		w.Walk(n.Body)

	case *Return:
		assert.Ok(n.X != nil)
		w.Walk(n.X)

	case *Break:
		if n.Label != nil {
			w.Walk(n.Label)
		}

	case *Continue:
		if n.Label != nil {
			w.Walk(n.Label)
		}

	default:
		panic(fmt.Sprintf("ast.(*defaultWalker).Walk: unknown node type '%T'", n))
	}

	w.stack = w.stack[:len(w.stack)-1]
	w.Walk(nil)
}

func walkList(w Walker, list *List) {
	for _, node := range list.Nodes {
		w.Walk(node)
	}
}
