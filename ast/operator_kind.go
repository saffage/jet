package ast

import "encoding/json"

//go:generate stringer -type=OperatorKind -linecomment -output=operator_kind_string.go
type OperatorKind byte

const (
	UnknownOperator OperatorKind = iota

	// Prefix.

	OperatorNot   // !
	OperatorNeg   // -
	OperatorAddr  // &
	OperatorDeref // *

	// Infix.

	OperatorAssign // =
	OperatorAdd    // +
	OperatorSub    // -
	OperatorMul    // *
	OperatorDiv    // /
	OperatorMod    // %
	OperatorEq     // ==
	OperatorNe     // !=
	OperatorLt     // <
	OperatorLe     // <=
	OperatorGt     // >
	OperatorGe     // >=
	OperatorBitAnd // &
	OperatorBitOr  // |
	OperatorBitXor // ^
	OperatorBitShl // <<
	OperatorBitShr // >>
	OperatorAnd    // and
	OperatorOr     // or

	// Postfix.

	// OperatorTry    // ?
	// OperatorUnwrap // !
)

func (kind OperatorKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}
