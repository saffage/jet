package types

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/report"
)

// Represents a compile-time known value.
// Also can represent just a type in some cases.
type Value struct {
	T Type
	V constant.Value // Can be nil.
}

func (v *Value) Type() Type            { return v.T }
func (v *Value) Value() constant.Value { return v.V }
func (v *Value) Index() int            { return 0 } // implements [pattern.constructor].

func constantFromNode(node *ast.Literal) constant.Value {
	switch node.Kind {
	case ast.IntLiteral:
		value := node.Data

		if suffixIdx := strings.LastIndex(node.Data, "'"); suffixIdx != -1 {
			value = value[:suffixIdx]
		}

		if value, ok := big.NewInt(0).SetString(value, 0); ok {
			return constant.NewBigInt(value)
		}

		// Unreachable?
		panic(fmt.Sprintf("invalid integer value for constant: '%s'", value))

	case ast.FloatLiteral:
		value := node.Data

		if suffixIdx := strings.LastIndex(node.Data, "'"); suffixIdx != -1 {
			value = value[:suffixIdx]
		}

		if value, ok := big.NewFloat(0.0).SetString(value); ok {
			return constant.NewBigFloat(value)
		}

		// Unreachable?
		panic(fmt.Sprintf("invalid float value for constant: '%s'", node.Data))

	case ast.StringLiteral:
		start := strings.IndexAny(node.Data, "\"'")
		end := strings.LastIndexAny(node.Data, "\"'")
		value := node.Data[start+1 : end]

		return constant.NewString(value)

	default:
		panic("unreachable")
	}
}

func (check *checker) valueOfInternal(expr ast.Node) (*Value, error) {
	switch node := expr.(type) {
	case *ast.Literal:
		value := constantFromNode(node)

		return &Value{
			T: FromConstant(value),
			V: value,
		}, nil

	case *ast.Lower:
		if sym := check.symbolOf(node); sym != nil {
			// if _const, _ := sym.(*Const); _const != nil {
			// 	return _const.value
			// }
			return nil, nil
		}

		report.Debug("%s => %+v\n", check.env.name, check.env.symbols)
		return nil, &errorUndefinedIdent{node}

	case *ast.Upper:
		if sym := check.symbolOf(node); sym != nil {
			// if _const, _ := sym.(*Const); _const != nil {
			// 	return _const.value
			// }
			return nil, nil
		}

		if sym := check.typeSymbolOf(node); sym != nil {
			// if _const, _ := sym.(*Const); _const != nil {
			// 	return _const.value
			// }
			return nil, nil
		}

		report.Debug("%s => %+v\n", check.env.name, check.env.symbols)
		return nil, &errorUndefinedIdent{node}

	case *ast.Op:
		if node.X == nil {
			y, err := check.valueOf(node.Y)
			if err != nil || y == nil {
				return nil, err
			}

			ty := check.prefix(node, y.T)
			if ty == nil {
				return nil, nil
			}

			return &Value{
				T: ty,
				V: compileUnaryOp(y.V, node.Kind),
			}, nil
		}

		x, err := check.valueOf(node.X)

		if err != nil || x == nil {
			return nil, err
		}

		y, err := check.valueOf(node.Y)

		if err != nil || y == nil {
			return nil, err
		}

		t, err := check.infix(node, x.T, y.T)
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, nil
		}

		if x.V == nil || y.V == nil {
			return nil, nil
		}

		if x.V.Kind() == y.V.Kind() {
			return &Value{
				T: t,
				V: comptimeBinaryOp(x.V, y.V, node.Kind),
			}, nil
		} else {
			panic("not implemented")
		}
	}

	return nil, nil
}

// func (check *checker) resolveBuiltInCall(node *ast.BuiltIn, call *ast.Call) *TypedValue {
// 	idx := slices.IndexFunc(builtIns, func(b *BuiltIn) bool {
// 		return b.name == node.Data
// 	})
// 	if idx == -1 {
// 		check.errorf(node, "unknown built-in function '%s'", node.Repr())
// 		return nil
// 	}

// 	builtIn := builtIns[idx]

// 	tyArgList := check.typeOfParenList(call.Args)
// 	if tyArgList == nil {
// 		return nil
// 	}

// 	tyArgs, _ := tyArgList.(*Tuple)
// 	if tyArgs == nil {
// 		return nil
// 	}

// 	if idx, err := builtIn.t.CheckArgs(tyArgs); err != nil {
// 		n := ast.Node(call.Args)

// 		if idx < len(call.Args.Nodes) {
// 			n = call.Args.Nodes[idx]
// 		}

// 		check.errorf(n, err.Error())
// 		return nil
// 	}

// 	vArgs := make([]*TypedValue, tyArgs.Len())

// 	for i := range len(vArgs) {
// 		vArgs[i] = check.module.Types[call.Args.Nodes[i]]
// 	}

// 	value, err := builtIn.f(call.Args, vArgs)
// 	if err != nil {
// 		check.addError(err)
// 		return nil
// 	}
// 	if value == nil {
// 		return nil
// 	}

// 	return value
// }

func comptimeBinaryOp(x, y constant.Value, opKind ast.OperatorKind) constant.Value {
	assert(x.Kind() == y.Kind())

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

	default:
		panic(fmt.Sprintf("invalid unary operation: '%s'", opKind))
	}
}
