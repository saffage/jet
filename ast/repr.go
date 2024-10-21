package ast

import (
	"fmt"
	"strconv"
	"strings"
)

type representable interface {
	// Representation of the node tree. Result must be equal to the
	// code from which this tree can be parsed.
	Repr() string
}

func (stmts Stmts) Repr() string {
	return printList([]Node(stmts), "; ")
}

//------------------------------------------------
// Atoms
//------------------------------------------------

func (node *BadNode) Repr() string {
	return "__bad_node__"
}

func (node *Empty) Repr() string {
	return ";"
}

func (node *Lower) Repr() string {
	return node.Data
}

func (node *Upper) Repr() string {
	return node.Data
}

func (node *TypeVar) Repr() string {
	return node.Data
}

func (node *Underscore) Repr() string {
	return node.Data
}

func (node *Literal) Repr() string {
	switch node.Kind {
	case IntLiteral, FloatLiteral:
		return node.Data

	case StringLiteral:
		// TODO replace [strconv.Quote].
		return strconv.Quote(node.Data[1 : len(node.Data)-1])

	default:
		panic("unreachable")
	}
}

//------------------------------------------------
// Declaration
//------------------------------------------------

func (node *LetDecl) Repr() string {
	return fmt.Sprintf(
		"let %s = %s",
		node.Decl.Repr(),
		node.Value.Repr(),
	)
}

func (node *TypeDecl) Repr() string {
	buf := strings.Builder{}
	buf.WriteString("type " + node.Name.Repr())

	if node.Args != nil {
		buf.WriteString(node.Args.Repr())
	}

	if _, ok := node.Expr.(*Block); ok {
		buf.WriteString(" " + node.Expr.Repr())
	} else {
		buf.WriteString(" = " + node.Expr.Repr())
	}

	return buf.String()
}

func (node *Decl) Repr() string {
	switch node.Type.(type) {
	case *Signature, *Parens:
		if node.Name == nil {
			return node.Type.Repr()
		}
		return node.Name.Repr() + node.Type.Repr()

	default:
		if node.TypeTok.IsValid() {
			if node.Name == nil {
				panic("ast.(*Decl).Repr: type declaration should have name")
			}
			if node.Type == nil {
				return "type " + node.Name.Repr()
			}
			return "type " + node.Name.Repr() + " " + node.Type.Repr()
		}
		if node.Name == nil {
			if node.Type == nil {
				panic("ast.(*Decl).Repr: declaration must have name, type or both")
			}
			return node.Type.Repr()
		}
		if node.Type == nil {
			return node.Name.Repr()
		}
		return node.Name.Repr() + " " + node.Type.Repr()
	}
}

func (node *Variant) Repr() string {
	if node.Params != nil {
		return node.Name.Repr() + node.Params.Repr()
	}

	return node.Name.Repr()
}

//------------------------------------------------
// Composite nodes
//------------------------------------------------

func (node *Label) Repr() string {
	if node.Name != nil {
		return node.Name.Repr() + ": " + node.X.Repr()
	}
	return ":" + node.X.Repr()
}

func (node *Signature) Repr() string {
	if node.Result == nil {
		return node.Params.Repr()
	}
	return fmt.Sprintf("%s %s", node.Params.Repr(), node.Result.Repr())
}

func (node *Function) Repr() string {
	if node.Body != nil {
		return fmt.Sprintf("fn%s = %s", node.Params.Repr(), node.Body.Repr())
	}
	return fmt.Sprintf("fn%s", node.Params.Repr())
}

func (node *Call) Repr() string {
	return node.X.Repr() + node.Args.Repr()
}

func (node *Dot) Repr() string {
	return node.X.Repr() + "." + node.Y.Repr()
}

func (node *Op) Repr() string {
	if node.X != nil {
		if node.Y != nil {
			return fmt.Sprintf(
				"%s %s %s",
				node.X.Repr(),
				node.Kind.String(),
				node.Y.Repr(),
			)
		}
		return node.X.Repr() + node.Kind.String()
	}
	if node.Y != nil {
		return node.Kind.String() + node.Y.Repr()
	}
	return node.Kind.String()
}

//------------------------------------------------
// Lists
//------------------------------------------------

func (node *Block) Repr() string {
	if len(node.Nodes) == 0 {
		return "{}"
	}
	return fmt.Sprintf("{ %s }", printList(node.Nodes, "; "))
}

func (node *List) Repr() string {
	return fmt.Sprintf("[%s]", printList(node.Nodes, ", "))
}

func (node *Parens) Repr() string {
	return fmt.Sprintf("(%s)", printList(node.Nodes, ", "))
}

func printList[T Node](nodes []T, separator string) string {
	buf := strings.Builder{}
	for i, node := range nodes {
		if i > 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(node.Repr())
	}
	return buf.String()
}

//------------------------------------------------
// Language constructions
//------------------------------------------------

func (node *When) Repr() string {
	return fmt.Sprintf("when %s %s", node.Expr.Repr(), node.Body.Repr())
}

func (node *Extern) Repr() string {
	if node.Args != nil {
		return "extern" + node.Args.Repr()
	}
	return "extern"
}
