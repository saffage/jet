package ast

//go:generate stringer -type=LiteralKind,PrefixOpKind,InfixOpKind,PostfixOpKind,GenericDeclKind -output=kind_string.go -linecomment

type (
	LiteralKind     byte
	PrefixOpKind    byte
	InfixOpKind     byte
	PostfixOpKind   byte
	GenericDeclKind byte
)

const (
	UnknownLiteral LiteralKind = iota // unknown literal kind

	IntLiteral    // int
	FloatLiteral  // float
	StringLiteral // string
)

const (
	UnknownPrefix PrefixOpKind = iota // unknown prefix operation

	PrefixNot     // !
	PrefixNeg     // -
	PrefixAddr    // &
	PrefixMutAddr // &var
)

const (
	UnknownInfix InfixOpKind = iota // unknown infix operation

	InfixAdd    // +
	InfixSub    // -
	InfixMult   // *
	InfixDiv    // /
	InfixMod    // %
	InfixAssign // =
	InfixEq     // ==
	InfixNe     // !=
	InfixLt     // <
	InfixLe     // <=
	InfixGt     // >
	InfixGe     // >=
)

const (
	UnknownPostfix PostfixOpKind = iota // unknown postfix operation

	PostfixTry    // ?
	PostfixUnwrap // !
)

const (
	UnknownDecl GenericDeclKind = iota // unknown generic declaration kind

	ConstDecl // const
	VarDecl   // var
	ValDecl   // val
)
