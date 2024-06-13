package ast

import "encoding/json"

//go:generate stringer -type=OperatorKind -linecomment -output=operator_kind_string.go
type OperatorKind byte

const (
	UnknownOperator OperatorKind = iota

	// Prefix.

	OperatorNot      // !
	OperatorNeg      // -
	OperatorAddrOf   // &
	OperatorStar     // *
	OperatorEllipsis // ...

	// Infix.

	OperatorAssign         // =
	OperatorAddAndAssign   // +=
	OperatorSubAndAssign   // -=
	OperatorMultAndAssign  // *=
	OperatorDivAndAssign   // /=
	OperatorModAndAssign   // %=
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

	// Postfix.

	// OperatorTry    // ?
	// OperatorUnwrap // !
)

func (kind OperatorKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}
