package ast

import "fmt"

type TopDownWalker struct{}

var _ Walker = TopDownWalker{}

// WalkTopDown performs top-down tree traversal.
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

	case *Binding:
		switch v := v.(type) {
		case BindingVisitor:
			v.VisitBinding(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Binding(n, v)
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

	case *Field:
		switch v := v.(type) {
		case FieldVisitor:
			v.VisitField(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Field(n, v)
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

	case *Fn:
		switch v := v.(type) {
		case FnVisitor:
			v.VisitFn(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.Fn(n, v)
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

	case *When:
		switch v := v.(type) {
		case WhenVisitor:
			v.VisitWhen(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.When(n, v)
		}

	case *CaseClause:
		switch v := v.(type) {
		case CaseClauseVisitor:
			v.VisitCaseClause(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.CaseClause(n, v)
		}

	case *External:
		switch v := v.(type) {
		case ExternalVisitor:
			v.VisitExternal(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.External(n, v)
		}

	case *PatternLiteral:
		switch v := v.(type) {
		case PatternLiteralVisitor:
			v.VisitPatternLiteral(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternLiteral(n, v)
		}

	case *PatternBinding:
		switch v := v.(type) {
		case PatternBindingVisitor:
			v.VisitPatternBinding(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternBinding(n, v)
		}

	case *PatternPlaceholder:
		switch v := v.(type) {
		case PatternPlaceholderVisitor:
			v.VisitPatternPlaceholder(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternPlaceholder(n, v)
		}

	case *PatternRebinding:
		switch v := v.(type) {
		case PatternRebindingVisitor:
			v.VisitPatternRebinding(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternRebinding(n, v)
		}

	case *PatternList:
		switch v := v.(type) {
		case PatternListVisitor:
			v.VisitPatternList(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternList(n, v)
		}

	case *PatternRange:
		switch v := v.(type) {
		case PatternRangeVisitor:
			v.VisitPatternRange(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternRange(n, v)
		}

	case *PatternVariant:
		switch v := v.(type) {
		case PatternVariantVisitor:
			v.VisitPatternVariant(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternVariant(n, v)
		}

	case *PatternTypeTest:
		switch v := v.(type) {
		case PatternTypeTestVisitor:
			v.VisitPatternTypeTest(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternTypeTest(n, v)
		}

	case *PatternAlternative:
		switch v := v.(type) {
		case PatternAlternativeVisitor:
			v.VisitPatternAlternative(n)
		case NodeVisitor:
			v.Visit(n)
		default:
			walk.PatternAlternative(n, v)
		}

	default:
		panic(fmt.Sprintf("ast.(TopDownWalker).Walk: unimplemented for %T", n))
	}
}

func (walk TopDownWalker) Binding(node *Binding, v Visitor) {
	WalkTopDown(node.Pattern, v)
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

func (walk TopDownWalker) Field(node *Field, v Visitor) {
	if node.Label != nil {
		WalkTopDown(node.Label, v)
	}

	WalkTopDown(node.Name, v)
	WalkTopDown(node.Type, v)
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

func (walk TopDownWalker) Fn(node *Fn, v Visitor) {
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

func (walk TopDownWalker) When(node *When, v Visitor) {
	WalkTopDown(node.Cond, v)

	for _, node := range node.Clauses.Stmts.Items {
		WalkTopDown(node, v)
	}
}

func (walk TopDownWalker) CaseClause(node *CaseClause, v Visitor) {
	WalkTopDown(node.Pattern, v)
	WalkTopDown(node.Expr, v)
}

func (walk TopDownWalker) External(node *External, v Visitor) {
	if node.Args != nil {
		for _, node := range node.Args.Nodes {
			WalkTopDown(node, v)
		}
	}
}

func (walk TopDownWalker) PatternLiteral(node *PatternLiteral, v Visitor) {
	WalkTopDown(node.Value, v)
}

func (walk TopDownWalker) PatternBinding(node *PatternBinding, v Visitor) {
	WalkTopDown(node.Name, v)
}

func (walk TopDownWalker) PatternPlaceholder(node *PatternPlaceholder, v Visitor) {
	WalkTopDown(node.Name, v)
}

func (walk TopDownWalker) PatternRebinding(node *PatternRebinding, v Visitor) {
	WalkTopDown(node.X, v)
	WalkTopDown(node.Name, v)
}

func (walk TopDownWalker) PatternList(node *PatternList, v Visitor) {
	for _, item := range node.Items {
		WalkTopDown(item, v)
	}
}

func (walk TopDownWalker) PatternRange(node *PatternRange, v Visitor) {
	WalkTopDown(node.Name, v)
}

func (walk TopDownWalker) PatternVariant(node *PatternVariant, v Visitor) {
	WalkTopDown(node.Name, v)

	for _, value := range node.Values {
		WalkTopDown(value, v)
	}
}

func (walk TopDownWalker) PatternTypeTest(node *PatternTypeTest, v Visitor) {
	WalkTopDown(node.X, v)
	WalkTopDown(node.Type, v)
}

func (walk TopDownWalker) PatternAlternative(node *PatternAlternative, v Visitor) {
	for _, item := range node.Items {
		WalkTopDown(item, v)
	}
}

func (walk TopDownWalker) parens(node *Parens, v Visitor) {
	for _, node := range node.Nodes {
		WalkTopDown(node, v)
	}
}
