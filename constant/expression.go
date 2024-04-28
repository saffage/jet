package constant

import (
	"fmt"
	"math/big"

	"github.com/saffage/jet/ast"
)

func NewExpression(node ast.Node) Value {
	return &expression{
		Node: node,
	}
}

func FromNode(node ast.Node) Value {
	if n, isLiteral := node.(*ast.Literal); isLiteral {
		switch n.Kind {
		case ast.IntLiteral:
			if value, ok := new(big.Int).SetString(n.Value, 0); ok {
				return NewInt(value)
			}

			// Unreachable?
			panic(fmt.Sprintf("invalid integer value for constant: '%s'", n.Value))

		case ast.FloatLiteral:
			if value, ok := new(big.Float).SetString(n.Value); ok {
				return NewFloat(value)
			}

			// Unreachable?
			panic(fmt.Sprintf("invalid float value for constant: '%s'", n.Value))

		case ast.StringLiteral:
			return NewString(n.Value)

		default:
			panic("unreachable")
		}
	}

	return NewExpression(node)
}

type expression struct {
	Node ast.Node
}

func (expr *expression) Kind() Kind {
	return Expression
}

func (expr *expression) String() string {
	return expr.Node.String()
}
