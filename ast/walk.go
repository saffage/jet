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
type Visitor func(Node) (Visitor, error)

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
func WalkTopDown(visit Visitor, tree Node) error {
	if tree == nil {
		panic("can't walk a nil node")
	}

	if v, err := visit(tree); err != nil {
		return err
	} else if v == nil {
		return nil
	} else {
		visit = v
	}

	switch n := tree.(type) {
	case *BadNode, *Empty, *Ident, *Literal, *PrefixOpr, *InfixOpr, *PostfixOpr, *Star, *Comment, *CommentGroup:
		// Nothing to walk

	case *Field:
		for _, name := range n.Names {
			assert.Ok(name != nil)

			if err := WalkTopDown(visit, name); err != nil {
				return err
			}
		}

		if n.Type != nil {
			if err := WalkTopDown(visit, n.Type); err != nil {
				return err
			}
		}

		if n.Value != nil {
			if err := WalkTopDown(visit, n.Value); err != nil {
				return err
			}
		}

	case *Signature:
		assert.Ok(n.Params != nil)

		if err := walkList(visit, n.Params.List); err != nil {
			return err
		}

		if n.Result != nil {
			if err := WalkTopDown(visit, n.Result); err != nil {
				return err
			}
		}

	case *Call:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		if err := WalkTopDown(visit, n.X); err != nil {
			return err
		}

		if err := walkList(visit, n.Args.List); err != nil {
			return err
		}

	case *Index:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		if err := WalkTopDown(visit, n.X); err != nil {
			return err
		}

		if err := walkList(visit, n.Args.List); err != nil {
			return err
		}

	case *ArrayType:
		assert.Ok(n.X != nil)
		assert.Ok(n.Args != nil)

		if err := WalkTopDown(visit, n.X); err != nil {
			return err
		}

		if err := walkList(visit, n.Args.List); err != nil {
			return err
		}

	case *MemberAccess:
		assert.Ok(n.X != nil)
		assert.Ok(n.Selector != nil)

		if err := WalkTopDown(visit, n.X); err != nil {
			return err
		}

		if err := WalkTopDown(visit, n.Selector); err != nil {
			return err
		}

	case *PrefixOp:
		assert.Ok(n.X != nil)

		if err := WalkTopDown(visit, n.X); err != nil {
			return err
		}

	case *InfixOp:
		assert.Ok(n.X != nil)
		assert.Ok(n.Y != nil)

		if err := WalkTopDown(visit, n.X); err != nil {
			return err
		}

		if err := WalkTopDown(visit, n.Y); err != nil {
			return err
		}

	case *PostfixOp:
		assert.Ok(n.X != nil)

		if err := WalkTopDown(visit, n.X); err != nil {
			return err
		}

	case *List:
		if err := walkList(visit, n); err != nil {
			return err
		}

	case *ParenList:
		if err := walkList(visit, n.List); err != nil {
			return err
		}

	case *CurlyList:
		if err := walkList(visit, n.List); err != nil {
			return err
		}

	case *BracketList:
		if err := walkList(visit, n.List); err != nil {
			return err
		}

	case *AttributeList:
		assert.Ok(n.List != nil)

		if err := walkList(visit, n.List.List); err != nil {
			return err
		}

	case *BuiltInCall:
		assert.Ok(n.Name != nil)
		assert.Ok(n.X != nil)

		if err := WalkTopDown(visit, n.Name); err != nil {
			return err
		}

		if err := WalkTopDown(visit, n.X); err != nil {
			return err
		}

	case *ModuleDecl:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Body != nil)

		if n.Attrs != nil {
			if err := WalkTopDown(visit, n.Attrs); err != nil {
				return err
			}
		}

		if err := WalkTopDown(visit, n.Name); err != nil {
			return err
		}

		switch b := n.Body.(type) {
		case *List:
			if err := walkList(visit, b); err != nil {
				return err
			}

		case *CurlyList:
			if err := walkList(visit, b.List); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unexpected node type '%T' for module body", n.Body)
		}

	case *GenericDecl:
		assert.Ok(n.Field != nil)

		if n.Attrs != nil {
			if err := WalkTopDown(visit, n.Attrs); err != nil {
				return err
			}
		}

		if err := WalkTopDown(visit, n.Field); err != nil {
			return err
		}

	case *FuncDecl:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Signature != nil)

		if n.Attrs != nil {
			if err := WalkTopDown(visit, n.Attrs); err != nil {
				return err
			}
		}

		if err := WalkTopDown(visit, n.Name); err != nil {
			return err
		}

		if err := WalkTopDown(visit, n.Signature); err != nil {
			return err
		}

		if n.Body != nil {
			if err := WalkTopDown(visit, n.Body); err != nil {
				return err
			}
		}

	case *TypeAliasDecl:
		assert.Ok(n.Name != nil)
		assert.Ok(n.Expr != nil)

		if n.Attrs != nil {
			if err := WalkTopDown(visit, n.Attrs); err != nil {
				return err
			}
		}

		if err := WalkTopDown(visit, n.Name); err != nil {
			return err
		}

		if err := WalkTopDown(visit, n.Expr); err != nil {
			return err
		}

	case *If:
		assert.Ok(n.Cond != nil)
		assert.Ok(n.Body != nil)

		if err := WalkTopDown(visit, n.Cond); err != nil {
			return err
		}

		if err := WalkTopDown(visit, n.Body); err != nil {
			return err
		}

		if n.Else != nil {
			if err := WalkTopDown(visit, n.Else); err != nil {
				return err
			}
		}

	case *Else:
		assert.Ok(n.Body != nil)

		if err := WalkTopDown(visit, n.Body); err != nil {
			return err
		}

	case *While:
		assert.Ok(n.Cond != nil)
		assert.Ok(n.Body != nil)

		if err := WalkTopDown(visit, n.Cond); err != nil {
			return err
		}

		if err := WalkTopDown(visit, n.Body); err != nil {
			return err
		}

	case *Return:
		if n.X != nil {
			if err := WalkTopDown(visit, n.X); err != nil {
				return err
			}
		}

	case *Break:
		if n.Label != nil {
			if err := WalkTopDown(visit, n.Label); err != nil {
				return err
			}
		}

	case *Continue:
		if n.Label != nil {
			if err := WalkTopDown(visit, n.Label); err != nil {
				return err
			}
		}

	default:
		// Should not happen.
		panic(fmt.Sprintf("unknown node type '%T'", n))
	}

	_, err := visit(nil)
	return err
}

func walkList(visit Visitor, list *List) error {
	if list != nil {
		for _, node := range list.Nodes {
			if err := WalkTopDown(visit, node); err != nil {
				return err
			}
		}
	}

	return nil
}
