package checker

import (
	"fmt"
	"math/big"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/types"
)

// Represents a compile-time known value.
// Also can represent a type in some situations.
type TypedValue struct {
	Type  types.Type
	Value constant.Value // Can be nil.
}

func constantFromNode(node *ast.Literal) constant.Value {
	if node == nil {
		panic("unnreachable")
	}

	switch node.Kind {
	case ast.IntLiteral:
		if value, ok := big.NewInt(0).SetString(node.Value, 0); ok {
			return constant.NewBigInt(value)
		}

		// Unreachable?
		panic(fmt.Sprintf("invalid integer value for constant: '%s'", node.Value))

	case ast.FloatLiteral:
		if value, ok := big.NewFloat(0.0).SetString(node.Value); ok {
			return constant.NewBigFloat(value)
		}

		// Unreachable?
		panic(fmt.Sprintf("invalid float value for constant: '%s'", node.Value))

	case ast.StringLiteral:
		return constant.NewString(node.Value)

	default:
		panic("unreachable")
	}
}

func (check *Checker) valueOfInternal(expr ast.Node) *TypedValue {
	switch node := expr.(type) {
	case *ast.Literal:
		value := constantFromNode(node)
		type_ := types.FromConstant(value)

		if type_ == types.UntypedString {
			check.module.Data[node] = &TypedValue{type_, value}
		}

		return &TypedValue{type_, value}

	case *ast.Ident:
		if _const, _ := check.symbolOf(node).(*Const); _const != nil {
			return _const.value
		}

	case *ast.InfixOp:
		x := check.valueOf(node.X)
		y := check.valueOf(node.Y)

		if x == nil || y == nil {
			return nil
		}

		t := check.infix(node, x.Type, y.Type)
		if t == nil {
			return nil
		}

		if x.Value.Kind() == y.Value.Kind() {
			return &TypedValue{
				Type:  t,
				Value: comptimeBinaryOp(x.Value, y.Value, node.Opr.Kind),
			}
		} else {
			panic("not implemented")
		}
	}

	return nil
}

func comptimeBinaryOp(x, y constant.Value, opKind ast.OperatorKind) constant.Value {
	assert.Ok(x.Kind() == y.Kind())

	switch opKind {
	case ast.OperatorAdd:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBigInt(new(big.Int).Add(x, y))

		case constant.Float:
			x, y := constant.AsFloat(x), constant.AsFloat(y)
			return constant.NewBigFloat(new(big.Float).Add(x, y))

		default:
			panic("unreachable")
		}

	case ast.OperatorSub:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBigInt(new(big.Int).Sub(x, y))

		case constant.Float:
			x, y := constant.AsFloat(x), constant.AsFloat(y)
			return constant.NewBigFloat(new(big.Float).Sub(x, y))

		default:
			panic("unreachable")
		}

	case ast.OperatorMul:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBigInt(new(big.Int).Mul(x, y))

		case constant.Float:
			x, y := constant.AsFloat(x), constant.AsFloat(y)
			return constant.NewBigFloat(new(big.Float).Mul(x, y))

		default:
			panic("unreachable")
		}

	case ast.OperatorDiv:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBigInt(new(big.Int).Div(x, y))

		case constant.Float:
			x, y := constant.AsFloat(x), constant.AsFloat(y)
			return constant.NewBigFloat(new(big.Float).Quo(x, y))

		default:
			panic("unreachable")
		}

	case ast.OperatorMod:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBigInt(new(big.Int).Mod(x, y))

		default:
			panic("unreachable")
		}

	case ast.OperatorBitAnd:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBigInt(new(big.Int).And(x, y))

		default:
			panic("unreachable")
		}

	case ast.OperatorBitOr:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBigInt(new(big.Int).Or(x, y))

		default:
			panic("unreachable")
		}

	case ast.OperatorBitXor:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBigInt(new(big.Int).Xor(x, y))

		default:
			panic("unreachable")
		}

	case ast.OperatorBitShl:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			if y.Sign() == -1 {
				// error: shift count cannot be negative
				return nil
			}
			if y.BitLen() > 32 {
				// error: value is too big
				return nil
			}
			n := uint(y.Int64())
			return constant.NewBigInt(new(big.Int).Lsh(x, n))

		default:
			panic("unreachable")
		}

	case ast.OperatorBitShr:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			if y.Sign() == -1 {
				// error: shift count cannot be negative
				return nil
			}
			if y.BitLen() > 32 {
				// error: value is too big
				return nil
			}
			n := uint(y.Int64())
			return constant.NewBigInt(new(big.Int).Rsh(x, n))

		default:
			panic("unreachable")
		}

	case ast.OperatorAnd:
		switch x.Kind() {
		case constant.Bool:
			x, y := constant.AsBool(x), constant.AsBool(y)
			return constant.NewBool(*x && *y)

		default:
			panic("unreachable")
		}

	case ast.OperatorOr:
		switch x.Kind() {
		case constant.Bool:
			x, y := constant.AsBool(x), constant.AsBool(y)
			return constant.NewBool(*x || *y)

		default:
			panic("unreachable")
		}

	case ast.OperatorEq:
		switch x.Kind() {
		case constant.Bool:
			x, y := constant.AsBool(x), constant.AsBool(y)
			return constant.NewBool(*x == *y)

		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBool(x.Cmp(y) == 0)

		case constant.Float:
			x, y := constant.AsFloat(x), constant.AsFloat(y)
			return constant.NewBool(x.Cmp(y) == 0)

		default:
			panic("unreachable")
		}

	case ast.OperatorLt:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBool(x.Cmp(y) == -1)

		case constant.Float:
			x, y := constant.AsFloat(x), constant.AsFloat(y)
			return constant.NewBool(x.Cmp(y) == -1)

		default:
			panic("unreachable")
		}

	case ast.OperatorGt:
		switch x.Kind() {
		case constant.Int:
			x, y := constant.AsInt(x), constant.AsInt(y)
			return constant.NewBool(x.Cmp(y) == 1)

		case constant.Float:
			x, y := constant.AsFloat(x), constant.AsFloat(y)
			return constant.NewBool(x.Cmp(y) == 1)

		default:
			panic("unreachable")
		}

	case ast.OperatorNe:
		result := constant.AsBool(comptimeBinaryOp(x, y, ast.OperatorEq))
		return constant.NewBool(!*result)

	case ast.OperatorLe:
		result := constant.AsBool(comptimeBinaryOp(x, y, ast.OperatorGt))
		return constant.NewBool(!*result)

	case ast.OperatorGe:
		result := constant.AsBool(comptimeBinaryOp(x, y, ast.OperatorLt))
		return constant.NewBool(!*result)

	default:
		panic(fmt.Sprintf("invalid binary operation: '%s'", opKind))
	}
}

func compileUnaryOp(x constant.Value, opKind ast.OperatorKind) constant.Value {
	switch opKind {
	case ast.OperatorNot:
		switch x.Kind() {
		case constant.Bool:
			x := constant.AsBool(x)
			return constant.NewBool(!*x)

		default:
			panic("unreachable")
		}

	case ast.OperatorNeg:
		switch x.Kind() {
		case constant.Int:
			x := constant.AsInt(x)
			return constant.NewBigInt(new(big.Int).Neg(x))

		case constant.Float:
			x := constant.AsFloat(x)
			return constant.NewBigFloat(new(big.Float).Neg(x))

		default:
			panic("unreachable")
		}

	case ast.OperatorAddr, ast.OperatorDeref:
		panic("pointer operation cannot be evaluated at compile-time")

	default:
		panic(fmt.Sprintf("invalid unary operation: '%s'", opKind))
	}
}
