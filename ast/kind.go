package ast

//go:generate stringer -type=LiteralKind,UnaryOpKind,BinaryOpKind,GenericDeclKind -output=kind_string.go -linecomment

type LiteralKind byte

const (
	UnknownLiteral LiteralKind = iota // unknown literal

	IntLiteral    // int
	FloatLiteral  // float
	StringLiteral // string
)

type UnaryOpKind byte

const (
	UnknownUnaryOp UnaryOpKind = iota // unknown unary operation

	UnaryNot     // !
	UnaryNeg     // -
	UnaryAddr    // &
	UnaryMutAddr // &var
)

type BinaryOpKind byte

const (
	UnknownBinaryOp BinaryOpKind = iota // unknown binary operation

	BinaryAdd    // +
	BinarySub    // -
	BinaryMult   // *
	BinaryDiv    // /
	BinaryMod    // %
	BinaryAssign // =
)

type GenericDeclKind byte

const (
	UnknownGenericDeclKind GenericDeclKind = iota // unknown generic declaration kind

	ConstDecl // const
	VarDecl   // var
	ValDecl   // val
)
