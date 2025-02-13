package ast

import "encoding/json"

//go:generate stringer -type=OperatorKind -linecomment -output=operator_kind_string.go
type OperatorKind byte

const (
	UnknownOperator OperatorKind = iota

	// Prefix.

	OperatorNot // !
	OperatorNeg // -

	// Infix.

	OperatorAssign    // =
	OperatorAdd       // +
	OperatorAddAssign // +=
	OperatorSub       // -
	OperatorSubAssign // -=
	OperatorMul       // *
	OperatorMulAssign // *=
	OperatorDiv       // /
	OperatorDivAssign // /=
	OperatorMod       // %
	OperatorModAssign // %=
	OperatorEq        // ==
	OperatorNe        // !=
	OperatorLt        // <
	OperatorLe        // <=
	OperatorGt        // >
	OperatorGe        // >=
	OperatorBitAnd    // &
	OperatorBitOr     // |

	// Postfix.
)

func (kind OperatorKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}
