package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (check *Checker) prefix(node *ast.PrefixOp, tOperand types.Type) types.Type {
	switch node.Opr.Kind {
	case ast.OperatorNeg:
		if p := types.AsPrimitive(tOperand); p != nil {
			switch p.Kind() {
			case types.UntypedInt, types.UntypedFloat, types.I32:
				return tOperand
			}
		}

	case ast.OperatorNot:
		if p := types.AsPrimitive(tOperand); p != nil {
			switch p.Kind() {
			case types.UntypedBool, types.Bool:
				return tOperand
			}
		}

	case ast.OperatorDeref:
		if ref := types.AsRef(tOperand); ref != nil {
			return ref.Base()
		}

		check.errorf(node.X, "expression is not a reference type")
		return nil

	case ast.OperatorAddr:
		if types.IsTypeDesc(tOperand) {
			t := types.NewRef(types.SkipTypeDesc(tOperand))
			return types.NewTypeDesc(t)
		}

		return types.NewRef(types.SkipUntyped(tOperand))

	case ast.OperatorMutAddr:
		panic("not implemented")

	default:
		panic(fmt.Sprintf("unknown prefix operator: '%s'", node.Opr.Kind))
	}

	check.errorf(
		node.Opr,
		"operator '%s' is not defined for the type (%s)",
		node.Opr.Kind,
		tOperand,
	)
	return nil
}

func (check *Checker) infix(node *ast.InfixOp, tOperandX, tOperandY types.Type) types.Type {
	if !tOperandX.Equals(tOperandY) {
		check.errorf(node, "type mismatch (%s and %s)", tOperandX, tOperandY)
		return nil
	}

	// Assignment operation doesn't have a value.
	if node.Opr.Kind == ast.OperatorAssign {
		return types.Unit
	}

	primitive := types.AsPrimitive(tOperandX)

	if primitive == nil {
		check.errorf(node, "only primitive types have operators")
		return nil
	}

	switch node.Opr.Kind {
	case ast.OperatorAdd,
		ast.OperatorSub,
		ast.OperatorMul,
		ast.OperatorDiv,
		ast.OperatorMod,
		ast.OperatorBitAnd,
		ast.OperatorBitOr,
		ast.OperatorBitXor,
		ast.OperatorBitShl,
		ast.OperatorBitShr:
		switch primitive.Kind() {
		case types.UntypedInt, types.UntypedFloat, types.I32:
			return tOperandX
		}

	case ast.OperatorEq,
		ast.OperatorNe,
		ast.OperatorLt,
		ast.OperatorLe,
		ast.OperatorGt,
		ast.OperatorGe:
		switch primitive.Kind() {
		case types.UntypedBool, types.UntypedInt, types.UntypedFloat:
			return types.Primitives[types.UntypedBool]

		case types.Bool, types.I32:
			return types.Primitives[types.Bool]
		}

	case ast.OperatorAnd, ast.OperatorOr:
		switch primitive.Kind() {
		case types.UntypedBool:
			return types.Primitives[types.UntypedBool]

		case types.Bool:
			return types.Primitives[types.Bool]
		}

	default:
		panic(fmt.Sprintf("unknown infix operator: '%s'", node.Opr.Kind))
	}

	check.errorf(
		node.Opr,
		"operator '%s' is not defined for the type (%s)",
		node.Opr.Kind,
		tOperandX,
	)
	return nil
}
