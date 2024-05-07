package checker

import (
	"fmt"
	"math/big"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
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

		if type_ == types.Primitives[types.UntypedString] {
			check.Data[node] = TypedValue{type_, value}
		}

		return &TypedValue{type_, value}

		// case *ast.Ident:
		// 	panic("constants are not implemented")

		// case *ast.PrefixOp, *ast.PostfixOp, *ast.InfixOp:
		// 	panic("not implemented")
	}

	return nil
}
