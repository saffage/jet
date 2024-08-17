package ast

import "encoding/json"

//go:generate stringer -type=OperatorKind -linecomment -output=operator_kind_string.go
type OperatorKind byte

const (
	UnknownOperator OperatorKind = iota

	// Prefix.

	OperatorNot       // !
	OperatorNeg       // -
	OperatorAddrOf    // &
	OperatorMutAddrOf // &mut
	OperatorPtr       // *
	OperatorMutPtr    // *mut
	OperatorEllipsis  // ...

	// Infix.

	OperatorAssign         // =
	OperatorAddAssign      // +=
	OperatorSubAssign      // -=
	OperatorMultAssign     // *=
	OperatorDivAssign      // /=
	OperatorModAssign      // %=
	OperatorBitAndAssign   // &=
	OperatorBitOrAssign    // |=
	OperatorBitXorAssign   // ^=
	OperatorBitShlAssign   // <<=
	OperatorBitShrAssign   // >>=
	OperatorAdd            // +
	OperatorSub            // -
	OperatorMul            // *
	OperatorDiv            // /
	OperatorMod            // %
	OperatorEq             // ==
	OperatorNe             // !=
	OperatorLt             // <
	OperatorLe             // <=
	OperatorGt             // >
	OperatorGe             // >=
	OperatorBitAnd         // &
	OperatorBitOr          // |
	OperatorBitXor         // ^
	OperatorBitShl         // <<
	OperatorBitShr         // >>
	OperatorAnd            // and
	OperatorOr             // or
	OperatorAs             // as
	OperatorRangeInclusive // ..
	OperatorRangeExclusive // ..<
	OperatorFatArrow       // =>

	// Postfix.

	// OperatorTry    // ?
	// OperatorUnwrap // !
)

func (kind OperatorKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}
