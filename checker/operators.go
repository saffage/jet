package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

func (check *Checker) prefix(node *ast.PrefixOp, tOperand types.Type) types.Type {
	switch node.Opr.Kind {
	case ast.OperatorNot:
		if p := types.AsPrimitive(tOperand); p != nil {
			switch p.Kind() {
			case types.KindUntypedBool, types.KindBool:
				return tOperand
			}
		}

	case ast.OperatorNeg:
		if p := types.AsPrimitive(tOperand); p != nil {
			switch p.Kind() {
			case types.KindUntypedInt, types.KindUntypedFloat, types.KindI32:
				return tOperand
			}
		}

	case ast.OperatorAddr:
		if types.IsTypeDesc(tOperand) {
			t := types.NewRef(types.SkipTypeDesc(tOperand))
			return types.NewTypeDesc(t)
		}

		switch operand := node.X.(type) {
		case *ast.Ident:
			if sym, _ := check.symbolOf(operand).(*Var); sym != nil {
				return types.NewRef(tOperand)
			}

		case *ast.MemberAccess:
			if types.IsStruct(check.typeOf(operand.X)) {
				return types.NewRef(tOperand)
			}

		case *ast.InfixOp:
			if operand.Opr.Kind == ast.OperatorDeref {
				return types.NewRef(tOperand)
			}
		}

		check.errorf(node.X, "expression is not an addressable location.")
		return nil

	case ast.OperatorDeref:
		if ref := types.AsRef(tOperand); ref != nil {
			return ref.Base()
		}

		check.errorf(node.X, "expression is not a reference type")
		return nil

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
	if !tOperandY.Equals(tOperandX) {
		check.errorf(node, "type mismatch (%s and %s)", tOperandX, tOperandY)
		return nil
	}

	// Assignment operation doesn't have a value.
	if node.Opr.Kind == ast.OperatorAssign {
		check.assignable(node.X)
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
		ast.OperatorDiv:
		switch primitive.Kind() {
		case types.KindUntypedInt, types.KindUntypedFloat, types.KindI32, types.KindU8:
			return tOperandX
		}

	case ast.OperatorMod,
		ast.OperatorBitAnd,
		ast.OperatorBitOr,
		ast.OperatorBitXor,
		ast.OperatorBitShl,
		ast.OperatorBitShr:
		switch primitive.Kind() {
		case types.KindUntypedInt, types.KindI32, types.KindU8:
			return tOperandX
		}

	case ast.OperatorEq,
		ast.OperatorNe,
		ast.OperatorLt,
		ast.OperatorLe,
		ast.OperatorGt,
		ast.OperatorGe:
		switch primitive.Kind() {
		case types.KindUntypedBool, types.KindUntypedInt, types.KindUntypedFloat:
			return types.UntypedBool

		case types.KindBool, types.KindI32, types.KindU8:
			return types.Bool
		}

	case ast.OperatorAnd, ast.OperatorOr:
		switch primitive.Kind() {
		case types.KindUntypedBool:
			return types.UntypedBool

		case types.KindBool:
			return types.Bool
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

func (check *Checker) assignable(node ast.Node) bool {
	// TODO allow assignment to a references.
	switch operand := node.(type) {
	case *ast.Ident:
		if operand != nil {
			varSym, ok := check.symbolOf(operand).(*Var)
			if !ok || varSym == nil {
				check.errorf(operand, "identifier is not a variable")
				return false
			}

			report.TaggedDebugf("checker", "assign '%s' at '%s'", varSym.name, operand)
			check.newUse(operand, varSym)
			return true
		}

	case *ast.MemberAccess:
		if operand != nil {
			fieldIdent, _ := operand.Selector.(*ast.Ident)
			if fieldIdent == nil {
				break
			}

			fieldSym, ok := check.symbolOf(fieldIdent).(*Var)
			if !ok || fieldSym == nil {
				check.errorf(fieldIdent, "identifier is not a variable")
				return false
			}

			check.newUse(fieldIdent, fieldSym)
			return check.assignable(operand.X)
		}

	case *ast.Index:
		if operand != nil {
			operandName, _ := operand.X.(*ast.Ident)
			if operandName == nil {
				check.errorf(operand.X, "expected identifier")
				return false
			}

			varSym, ok := check.symbolOf(operandName).(*Var)
			if !ok || varSym == nil {
				check.errorf(operand, "identifier is not a variable")
				return false
			}

			report.TaggedDebugf("checker", "assign '%s' at '%s'", varSym.name, operand)
			check.newUse(operandName, varSym)
			return true
		}
	}

	check.errorf(node, "expression is not assignable")
	return false
}
