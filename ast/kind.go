package ast

//go:generate stringer -type=LiteralKind,UnaryOpKind,BinaryOpKind,GenericDeclKind -output=kind_string.go -linecomment

type (
	LiteralKind     byte
	UnaryOpKind     byte
	BinaryOpKind    byte
	GenericDeclKind byte
)

const (
	UnknownLiteral LiteralKind = iota // unknown literal

	IntLiteral    // int
	FloatLiteral  // float
	StringLiteral // string
)

const (
	UnknownUnaryOp UnaryOpKind = iota // unknown unary operation

	UnaryNot     // !
	UnaryNeg     // -
	UnaryAddr    // &
	UnaryMutAddr // &var
)

const (
	UnknownBinaryOp BinaryOpKind = iota // unknown binary operation

	BinaryAdd    // +
	BinarySub    // -
	BinaryMult   // *
	BinaryDiv    // /
	BinaryMod    // %
	BinaryAssign // =
	BinaryEq     // ==
	BinaryNe     // !=
	BinaryLt     // <
	BinaryLe     // <=
	BinaryGt     // >
	BinaryGe     // >=
)

const (
	UnknownGenericDeclKind GenericDeclKind = iota // unknown generic declaration kind

	ConstDecl // const
	VarDecl   // var
	ValDecl   // val
)
