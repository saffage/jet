package types

import (
	"errors"
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

func (check *checker) prefix(node *ast.Op, tOperand Type) Type {
	switch node.Kind {
	case ast.OperatorNot:
		if tOperand.Equal(BoolType) {
			return tOperand
		}

	case ast.OperatorNeg:
		if tOperand.Equal(IntType) {
			return tOperand
		}

	default:
		panic(fmt.Sprintf("unknown prefix operator: '%s'", node.Kind))
	}

	check.internalErrorf(
		node,
		"operator %s is not defined for the type `%s`",
		node.Kind,
		tOperand,
	)
	return nil
}

func (check *checker) infix(node *ast.Op, x, y Type) (Type, error) {
	// if node.Kind == ast.OperatorAs {
	// 	return check.infixAs(node, tOperandX, tOperandY)
	// }

	// Assignment operation doesn't have a value.
	if node.Kind == ast.OperatorAssign {
		errs := []error{}

		if !check.assignable(node.X) {
			errs = append(errs, internalErrorf(node.X, "expression cannot be assigned to"))
		}

		// TODO invalid type will be inferred if one of them is untyped
		if !y.Equal(x) && !IntoTyped(y).Equal(IntoTyped(x)) {
			errs = append(errs, &errorTypeMismatch{node.Y, node.X, y, x})
			// errorf(node, "type mismatch (%s and %s)", x, y)
		}

		// check.setType(node.Y, x)
		return NoneType, errors.Join(errs...)
	}

	for _, opTypes := range operatorTypes[node.Kind] {
		if x.Equal(opTypes.x) && y.Equal(opTypes.y) {
			_, isAssignOp := operatorTypesAssign[node.Kind]
			if isAssignOp && !check.assignable(node.X) {
				check.error(&errorNotAssignable{node.X})
				// check.errorf(node.X, "expression cannot be assigned to")
			}

			return opTypes.result, nil
		}
	}

	// TODO: add help message for possible operator types
	return nil, internalErrorf(
		node,
		"type mismatch for operator `%s`, got `%s` and `%s`",
		node.Kind,
		x,
		y,
	)
}

// func (check *checker) infixAs(node *ast.Op, _, tyY Type) Type {
// 	typedesc := As[*TypeDesc](tyY)
// 	if typedesc == nil {
// 		check.errorf(node.Y, "expected type, got '%s' instead", tyY)
// 		return nil
// 	}

// 	return SkipTypeDesc(typedesc)
// }

func (check *checker) postfix(node *ast.Op, tyOperand Type) Type {
	panic("unreachable")

	// switch node.Kind {
	// case ast.OperatorAddrOf:
	// 	switch operand := node.X.(type) {
	// 	case *ast.Ident:
	// 		if sym, _ := check.symbolOf(operand).(*Var); sym != nil {
	// 			return NewRef(tyOperand)
	// 		}

	// 	case *ast.Dot:
	// 		if IsStruct(check.typeOf(operand.X)) {
	// 			return NewRef(tyOperand)
	// 		}

	// 	// case *ast.SafeMemberAccess:
	// 	// 	if ptr := AsRef(check.typeOf(operand.X)); ptr != nil && IsStruct(ptr.Base()) {
	// 	// 		return NewRef(tyOperand)
	// 	// 	}

	// 	case *ast.Index:
	// 		if tArray := AsArray(check.typeOf(operand.X)); tArray != nil {
	// 			return NewRef(tArray.ElemType())
	// 		}

	// 	case *ast.Op:
	// 		if operand.Kind == ast.OperatorStar {
	// 			return NewRef(tyOperand)
	// 		}
	// 	}

	// 	check.errorf(node.X, "expression is not an addressable location")
	// 	return nil

	// case ast.OperatorStar:
	// 	if IsTypeDesc(tyOperand) {
	// 		t := NewRef(SkipTypeDesc(tyOperand))
	// 		return NewTypeDesc(t)
	// 	}

	// 	if ref := AsRef(tyOperand); ref != nil {
	// 		return ref.Base()
	// 	}

	// 	check.errorf(node.X, "expression is not a reference type")
	// 	return nil

	// default:
	// 	panic(fmt.Sprintf("unknown prefix operator: '%s'", node.Kind))
	// }
}

func (check *checker) assignable(node ast.Node) bool {
	switch operand := node.(type) {
	case *ast.Lower:
		if operand != nil {
			varSym, ok := check.symbolOf(operand).(*Binding)
			if !ok || varSym == nil {
				check.internalErrorf(operand, "identifier is not a variable")
				return false
			}

			report.DebugX("checker", "assign '%s' at '%s'", varSym.Name(), operand)
			check.newUse(operand, varSym)
			return true
		}

	case *ast.Dot:
		if operand != nil {
			// fieldIdent, _ := operand.Selector.(*ast.Ident)
			// if fieldIdent == nil {
			// 	break
			// }

			// fieldSym, ok := check.symbolOf(fieldIdent).(*Var)
			// if !ok || fieldSym == nil {
			// 	check.errorf(fieldIdent, "identifier is not a variable")
			// 	return false
			// }

			// check.newUse(fieldIdent, fieldSym)
			return check.assignable(operand.X)
		}
	}

	return false
}

type operandTypes struct{ x, y, result Type }

// NOTE assignment operator is checked before.
//
// NOTE additional checks for in-place assignment are located in [checker.matchOpTypes].
var operatorTypes = map[ast.OperatorKind][]operandTypes{
	ast.OperatorNot:          {{nil, BoolType, BoolType}},
	ast.OperatorNeg:          {{nil, IntType, IntType}, {nil, FloatType, FloatType}},
	ast.OperatorAdd:          {{IntType, IntType, IntType}, {FloatType, FloatType, FloatType}},
	ast.OperatorSub:          {{IntType, IntType, IntType}, {FloatType, FloatType, FloatType}},
	ast.OperatorMul:          {{IntType, IntType, IntType}, {FloatType, FloatType, FloatType}},
	ast.OperatorDiv:          {{IntType, IntType, IntType}, {FloatType, FloatType, FloatType}},
	ast.OperatorAddAssign:    {{IntType, IntType, NoneType}, {FloatType, FloatType, NoneType}},
	ast.OperatorSubAssign:    {{IntType, IntType, NoneType}, {FloatType, FloatType, NoneType}},
	ast.OperatorMulAssign:    {{IntType, IntType, NoneType}, {FloatType, FloatType, NoneType}},
	ast.OperatorDivAssign:    {{IntType, IntType, NoneType}, {FloatType, FloatType, NoneType}},
	ast.OperatorMod:          {{IntType, IntType, IntType}},
	ast.OperatorBitAnd:       {{IntType, IntType, IntType}},
	ast.OperatorBitOr:        {{IntType, IntType, IntType}},
	ast.OperatorBitXor:       {{IntType, IntType, IntType}},
	ast.OperatorBitShl:       {{IntType, IntType, IntType}},
	ast.OperatorBitShr:       {{IntType, IntType, IntType}},
	ast.OperatorModAssign:    {{IntType, IntType, NoneType}},
	ast.OperatorBitAndAssign: {{IntType, IntType, NoneType}},
	ast.OperatorBitOrAssign:  {{IntType, IntType, NoneType}},
	ast.OperatorBitXorAssign: {{IntType, IntType, NoneType}},
	ast.OperatorBitShlAssign: {{IntType, IntType, NoneType}},
	ast.OperatorBitShrAssign: {{IntType, IntType, NoneType}},
	ast.OperatorEq:           {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorNe:           {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorLt:           {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorLe:           {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorGt:           {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorGe:           {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorAnd:          {{BoolType, BoolType, BoolType}},
	ast.OperatorOr:           {{BoolType, BoolType, BoolType}},
}

var operatorTypesAssign = map[ast.OperatorKind]struct{}{
	ast.OperatorModAssign:    {},
	ast.OperatorBitAndAssign: {},
	ast.OperatorBitOrAssign:  {},
	ast.OperatorBitXorAssign: {},
	ast.OperatorBitShlAssign: {},
	ast.OperatorBitShrAssign: {},
}
