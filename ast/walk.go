package ast

// Walker interface implements tree walking algorithm.
type Walker interface {
	Walk(Node, Visitor)
}

// Visitor applies some action, specified by implementor, to a node.
//
// Implementor may implement any of the interfaces below.
type Visitor interface {
	IsVisitor()
}

// Visitor interface for a specific node type.
//
// Concrete node visitor have higher precedence over generic ones.
type (
	NodeVisitor  interface{ Visit(Node) }
	IdentVisitor interface{ VisitIdent(Ident) }

	BadNodeVisitor      interface{ VisitBadNode(*BadNode) }
	LowerVisitor        interface{ VisitLower(*Lower) }
	UpperVisitor        interface{ VisitUpper(*Capitalized) }
	PlaceholderVisitor  interface{ VisitPlaceholder(*Placeholder) }
	TypeVariableVisitor interface{ VisitTypeVariable(*TypeVariable) }
	LiteralVisitor      interface{ VisitLiteral(*Literal) }

	BindingVisitor   interface{ VisitBinding(*ValueDecl) }
	TypeAliasVisitor interface{ VisitTypeAlias(*TypeAliasDecl) }
	TypeDefVisitor   interface{ VisitTypeDef(*TypeDecl) }
	FieldVisitor     interface{ VisitField(*FieldDecl) }
	VariantVisitor   interface{ VisitVariant(*VariantDecl) }

	LabelVisitor     interface{ VisitLabel(*Label) }
	SignatureVisitor interface{ VisitSignature(*Signature) }
	FnTypeVisitor    interface{ VisitFnType(*FnType) }
	FnVisitor        interface{ VisitFn(*Fn) }
	CallVisitor      interface{ VisitCall(*Call) }
	DotVisitor       interface{ VisitDot(*Dot) }
	OpVisitor        interface{ VisitOp(*Op) }

	ListVisitor  interface{ VisitList(*List) }
	BlockVisitor interface{ VisitBlock(*Block) }
	StmtsVisitor interface{ VisitStmts(*Stmts) }

	WhenVisitor       interface{ VisitWhen(*When) }
	CaseClauseVisitor interface{ VisitCaseClause(*CaseClause) }
	ExternalVisitor   interface{ VisitExternal(*External) }

	PatternInvalidVisitor     interface{ VisitPatternInvalid(*PatternInvalid) }
	PatternLiteralVisitor     interface{ VisitPatternLiteral(*PatternLiteral) }
	PatternBindingVisitor     interface{ VisitPatternBinding(*PatternBinding) }
	PatternPlaceholderVisitor interface{ VisitPatternPlaceholder(*PatternPlaceholder) }
	PatternRebindingVisitor   interface{ VisitPatternRebinding(*PatternRebinding) }
	PatternListVisitor        interface{ VisitPatternList(*PatternList) }
	PatternRangeVisitor       interface{ VisitPatternRange(*PatternRange) }
	PatternVariantVisitor     interface{ VisitPatternVariant(*PatternVariant) }
	PatternTypeTestVisitor    interface{ VisitPatternTypeTest(*PatternTypeTest) }
	PatternAlternativeVisitor interface{ VisitPatternAlternative(*PatternAlternative) }
)
