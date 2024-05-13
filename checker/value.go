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
				Value: comptimeOp(x.Value, y.Value, node.Opr.Kind),
			}
		} else {
			panic("not implemented")
		}
	}

	return nil
}

func comptimeOp(x, y constant.Value, opKind ast.OperatorKind) constant.Value {
	assert.Ok(x.Kind() == y.Kind())

	switch opKind {
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

		case constant.String:
			panic("unreachable")

		default:
			panic("unreachable")
		}

	case ast.OperatorNe:
		result := constant.AsBool(comptimeOp(x, y, ast.OperatorEq))
		return constant.NewBool(!*result)

	default:
		panic("not implemented")
	}
}
