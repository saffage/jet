package ast

import "fmt"

type TopDownWalker struct{}

var _ Walker = TopDownWalker{}

// Top-down traversal. Visit a parent node before visiting its children.
func WalkTopDown(node Node, v Visitor) {
	if node == nil {
		panic("ast.WalkTopDown: ill-formed ast, node is nil")
	}

	if v == nil {
		panic("ast.WalkTopDown: visitor is nil")
	}

	TopDownWalker{}.Walk(node, v)
}

func (walk TopDownWalker) Walk(n Node, v Visitor) {
	switch n := n.(type) {
	case nil:
		panic("ast.(TopDownWalker).Walk: nil node")

	case *BadNode:
		switch v := v.(type) {
		case BadNodeVisitor:
			v.VisitBadNode(n)
		case NodeVisitor:
			v.Visit(n)
		}

	case *Lower:
		switch v := v.(type) {
		case LowerVisitor:
			v.VisitLower(n)
		case IdentVisitor:
			v.VisitIdent(n)
		case NodeVisitor:
			v.Visit(n)
		}

	case *Upper:
		switch v := v.(type) {
		case UpperVisitor:
			v.VisitUpper(n)
		case IdentVisitor:
			v.VisitIdent(n)
		case NodeVisitor:
			v.Visit(n)
		}

	case *Placeholder:
		switch v := v.(type) {
		case PlaceholderVisitor:
			v.VisitPlaceholder(n)
		case IdentVisitor:
			v.VisitIdent(n)
		case NodeVisitor:
			v.Visit(n)
		}

	case *Literal:
		switch v := v.(type) {
		case LiteralVisitor:
			v.VisitLiteral(n)
		case NodeVisitor:
			v.Visit(n)
		}

	case *LetDecl:
		switch v := v.(type) {
		case LetDeclVisitor:
			v.VisitLetDecl(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.LetDecl(n, v)
		}

	case *TypeAlias:
		switch v := v.(type) {
		case TypeAliasVisitor:
			v.VisitTypeAlias(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.TypeAlias(n, v)
		}

	case *TypeDef:
		switch v := v.(type) {
		case TypeDefVisitor:
			v.VisitTypeDef(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.TypeDef(n, v)
		}

	case *Decl:
		switch v := v.(type) {
		case DeclVisitor:
			v.VisitDecl(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Decl(n, v)
		}

	case *Variant:
		switch v := v.(type) {
		case VariantVisitor:
			v.VisitVariant(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Variant(n, v)
		}

	case *Label:
		switch v := v.(type) {
		case LabelVisitor:
			v.VisitLabel(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Label(n, v)
		}

	case *Signature:
		switch v := v.(type) {
		case SignatureVisitor:
			v.VisitSignature(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Signature(n, v)
		}

	case *Function:
		switch v := v.(type) {
		case FunctionVisitor:
			v.VisitFunction(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Function(n, v)
		}

	case *Call:
		switch v := v.(type) {
		case CallVisitor:
			v.VisitCall(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Call(n, v)
		}

	case *Dot:
		switch v := v.(type) {
		case DotVisitor:
			v.VisitDot(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Dot(n, v)
		}

	case *Op:
		switch v := v.(type) {
		case OpVisitor:
			v.VisitOp(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Op(n, v)
		}

	case *List:
		switch v := v.(type) {
		case ListVisitor:
			v.VisitList(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.List(n, v)
		}

	case *Stmts:
		switch v := v.(type) {
		case StmtsVisitor:
			v.VisitStmts(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Stmts(n, v)
		}

	case *Block:
		switch v := v.(type) {
		case BlockVisitor:
			v.VisitBlock(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Block(n, v)
		}

	case *Parens:
		switch v := v.(type) {
		case ParensVisitor:
			v.VisitParens(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Parens(n, v)
		}

	case *When:
		switch v := v.(type) {
		case WhenVisitor:
			v.VisitWhen(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.When(n, v)
		}

	case *Case:
		switch v := v.(type) {
		case CaseVisitor:
			v.VisitCase(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Case(n, v)
		}

	case *Spread:
		switch v := v.(type) {
		case SpreadVisitor:
			v.VisitSpread(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Spread(n, v)
		}

	case *As:
		switch v := v.(type) {
		case AsVisitor:
			v.VisitAs(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.As(n, v)
		}

	case *Extern:
		switch v := v.(type) {
		case ExternVisitor:
			v.VisitExtern(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Extern(n, v)
		}

	default:
		panic(fmt.Sprintf("ast.(TopDownWalker).Walk: unimplemented for %T", n))
	}
}

func (walk TopDownWalker) LetDecl(node *LetDecl, v Visitor) {
	WalkTopDown(node.Decl.Ident, v)

	if node.Decl.Type != nil {
		WalkTopDown(node.Decl.Type, v)
	}

	WalkTopDown(node.Value, v)
}

func (walk TopDownWalker) TypeAlias(node *TypeAlias, v Visitor) {
	WalkTopDown(node.Ident, v)

	if node.Args != nil {
		for _, node := range node.Args.Nodes {
			WalkTopDown(node, v)
		}
	}

	WalkTopDown(node.Expr, v)
}

func (walk TopDownWalker) TypeDef(node *TypeDef, v Visitor) {
	WalkTopDown(node.Ident, v)

	if node.Args != nil {
		for _, node := range node.Args.Nodes {
			WalkTopDown(node, v)
		}
	}

	WalkTopDown(node.Body, v)
}

func (walk TopDownWalker) Decl(node *Decl, v Visitor) {
	WalkTopDown(node.Ident, v)

	if node.Type != nil {
		WalkTopDown(node.Type, v)
	}
}

func (walk TopDownWalker) Variant(node *Variant, v Visitor) {
	WalkTopDown(node.Name, v)

	if node.Params != nil {
		for _, node := range node.Params.Nodes {
			WalkTopDown(node, v)
		}
	}
}

func (walk TopDownWalker) Label(node *Label, v Visitor) {
	WalkTopDown(node.Name, v)
	WalkTopDown(node.X, v)
}

func (walk TopDownWalker) Signature(node *Signature, v Visitor) {
	for _, node := range node.Params.Nodes {
		WalkTopDown(node, v)
	}

	if node.Result != nil {
		WalkTopDown(node.Result, v)
	}
}

func (walk TopDownWalker) Function(node *Function, v Visitor) {
	WalkTopDown(node.Signature, v)

	if node.Body != nil {
		WalkTopDown(node.Body, v)
	}
}

func (walk TopDownWalker) Call(node *Call, v Visitor) {
	WalkTopDown(node.X, v)

	for _, node := range node.Args.Nodes {
		WalkTopDown(node, v)
	}
}

func (walk TopDownWalker) Dot(node *Dot, v Visitor) {
	WalkTopDown(node.X, v)
	WalkTopDown(node.Y, v)
}

func (walk TopDownWalker) Op(node *Op, v Visitor) {
	if node.X != nil {
		WalkTopDown(node.X, v)
	}

	if node.Y != nil {
		WalkTopDown(node.Y, v)
	}
}

func (walk TopDownWalker) List(node *List, v Visitor) {
	for _, node := range node.Nodes {
		WalkTopDown(node, v)
	}
}

func (walk TopDownWalker) Stmts(stmts *Stmts, v Visitor) {
	for _, node := range stmts.Items {
		walk.Walk(node, v)
	}
}

func (walk TopDownWalker) Block(node *Block, v Visitor) {
	for _, node := range node.Stmts.Items {
		WalkTopDown(node, v)
	}
}

func (walk TopDownWalker) Parens(node *Parens, v Visitor) {
	for _, node := range node.Nodes {
		WalkTopDown(node, v)
	}
}

func (walk TopDownWalker) When(node *When, v Visitor) {
	WalkTopDown(node.Expr, v)

	for _, node := range node.Body.Stmts.Items {
		WalkTopDown(node, v)
	}
}

func (walk TopDownWalker) Case(node *Case, v Visitor) {
	WalkTopDown(node.Pattern, v)
	WalkTopDown(node.Expr, v)
}

func (walk TopDownWalker) Spread(node *Spread, v Visitor) {
	if node.Expr != nil {
		WalkTopDown(node.Expr, v)
	}
}

func (walk TopDownWalker) As(node *As, v Visitor) {
	WalkTopDown(node.Lhs, v)
	WalkTopDown(node.Rhs, v)
}

func (walk TopDownWalker) Extern(node *Extern, v Visitor) {
	if node.Args != nil {
		for _, node := range node.Args.Nodes {
			WalkTopDown(node, v)
		}
	}
}
