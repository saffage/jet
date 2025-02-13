package ast

// Walker interface implements tree walking algorithm.
type Walker interface {
	Walk(Node, Visitor)
}

// Applies some action, specified by implementor, to a node.
//
// Implementor may implement any of the interfaces below.
type Visitor interface {
	IsVisitor()
}

// Visitor interface for a specific node type.
//
// Concrete node visitor have higher precedence over generic one.
type (
	NodeVisitor  interface{ Visit(Node) }
	IdentVisitor interface{ VisitIdent(Ident) }

	BadNodeVisitor     interface{ VisitBadNode(*BadNode) }
	LowerVisitor       interface{ VisitLower(*Lower) }
	UpperVisitor       interface{ VisitUpper(*Upper) }
	PlaceholderVisitor interface{ VisitPlaceholder(*Placeholder) }
	LiteralVisitor     interface{ VisitLiteral(*Literal) }

	LetDeclVisitor   interface{ VisitLetDecl(*LetDecl) }
	TypeAliasVisitor interface{ VisitTypeAlias(*TypeAlias) }
	TypeDefVisitor   interface{ VisitTypeDef(*TypeDef) }
	DeclVisitor      interface{ VisitDecl(*Decl) }
	VariantVisitor   interface{ VisitVariant(*Variant) }

	LabelVisitor     interface{ VisitLabel(*Label) }
	SignatureVisitor interface{ VisitSignature(*Signature) }
	FunctionVisitor  interface{ VisitFunction(*Function) }
	CallVisitor      interface{ VisitCall(*Call) }
	DotVisitor       interface{ VisitDot(*Dot) }
	OpVisitor        interface{ VisitOp(*Op) }

	ListVisitor   interface{ VisitList(*List) }
	BlockVisitor  interface{ VisitBlock(*Block) }
	StmtsVisitor  interface{ VisitStmts(*Stmts) }
	ParensVisitor interface{ VisitParens(*Parens) }

	WhenVisitor   interface{ VisitWhen(*When) }
	CaseVisitor   interface{ VisitCase(*Case) }
	SpreadVisitor interface{ VisitSpread(*Spread) }
	AsVisitor     interface{ VisitAs(*As) }
	ExternVisitor interface{ VisitExtern(*Extern) }
)
