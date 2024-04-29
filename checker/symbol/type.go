package symbol

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/constant"
)

type nilTypeError struct {
	sym Symbol
	use *ast.Ident
}

func (err *nilTypeError) Error() string {
	return fmt.Sprintf("type of the `%s` symbol is not yet resolved", err.sym.Name())
}

// Meaning:
//   - `type_` - type of the expression.
//   - `required` - symbol that is required for this expression but have no type.
//   - `where` - identifier that refers to the symbol (declared or not),
//     wich type is required for inderring type of the expression.
//
// `type_` returns:
//   - nil - expression has no type.
//   - [types.Unknown] - expression should have type but doesn't (symbol is deffered or not defined).
func TypeOf(scope Scope, expr ast.Node) (types.Type, error) {
	switch node := ast.UnwrapParenExpr(expr).(type) {
	case *ast.Ident:
		if sym := scope.Resolve(node.Name); sym != nil {
			// symbol is defined
			if type_ := scope.Resolve(node.Name).Type(); type_ != nil {
				// symbol have a type (resolved)
				return type_, nil
			}
			// symbol is not yet resolver
			return types.Unknown{}, &nilTypeError{sym, node}
		}

		return nil, NewErrorf(node, "identifier `%s` is undefined", node.Name)

	case *ast.Literal:
		switch node.Kind {
		case ast.IntLiteral:
			return types.UntypedInt{}, nil

		case ast.FloatLiteral:
			return types.UntypedFloat{}, nil

		case ast.StringLiteral:
			return types.UntypedString{}, nil

		default:
			panic(fmt.Sprintf("unhandled literal kind: '%s'", node.Kind.String()))
		}

	case *ast.PrefixOp:
		switch node.Opr.Kind {
		case ast.PrefixNeg:
			type_, err := TypeOf(scope, node.X)
			if err != nil {
				return types.Unknown{}, err
			}

			switch type_.(type) {
			case types.UntypedInt, types.UntypedFloat, types.I32:
				return type_, nil

			default:
				panic(NewErrorf(
					node.Opr,
					"operator '%s' is not defined for the type '%s'",
					node.Opr.Kind.String(),
					type_.String(),
				))
			}

		case ast.PrefixNot:
			type_, err := TypeOf(scope, node.X)
			if err != nil {
				return types.Unknown{}, err
			}

			switch type_.(type) {
			case types.UntypedBool, types.Bool:
				return type_, nil

			default:
				panic(NewErrorf(
					node.X,
					"operator '%s' is not defined for the type '%s'",
					node.Opr.Kind.String(),
					type_.String(),
				))
			}

		case ast.PrefixAddr, ast.PrefixMutAddr:
			panic("not implemented")

		default:
			panic("unreachable")
		}

	case *ast.InfixOp:
		x_type, err := TypeOf(scope, node.X)
		if err != nil {
			return types.Unknown{}, err
		}

		y_type, err := TypeOf(scope, node.Y)
		if err != nil {
			return types.Unknown{}, err
		}

		if !x_type.Equals(y_type) {
			panic(NewErrorf(node, "type mismatch ('%s' and '%s')", x_type, y_type))
		}

		switch node.Opr.Kind {
		case ast.InfixAdd, ast.InfixSub, ast.InfixMult, ast.InfixDiv, ast.InfixMod:
			switch x_type.(type) {
			case types.UntypedInt, types.UntypedFloat, types.I32:
				return x_type, nil
			}

		case ast.InfixEq, ast.InfixNe, ast.InfixLt, ast.InfixLe, ast.InfixGt, ast.InfixGe:
			switch x_type.(type) {
			case types.UntypedBool, types.UntypedInt, types.UntypedFloat:
				return types.UntypedBool{}, nil

			case types.Bool, types.I32:
				return types.Bool{}, nil
			}
		}

		panic(NewErrorf(
			node.Opr,
			"operator '%s' is not defined for the type '%s'",
			node.Opr.Kind.String(),
			x_type.String(),
		))

	case *ast.BuiltInCall:
		builtin, ok := scope.Resolve("@" + node.Name.Name).(*Builtin)
		if !ok || builtin == nil {
			panic(NewErrorf(node.Name, "unknown builtin '@%s'", node.Name.Name))
		}

		result := builtin.fn(builtin, scope, node)

		switch x := result.(type) {
		case constant.Value:
			return types.FromConstantValue(x), nil

		case types.Type:
			return x, nil

		case error:
			return types.Unknown{}, x
		}

		panic("todo")

	case *ast.CurlyList:
		listScope := NewLocalScope(scope)
		defer listScope.Free()

		ast.WalkTopDown(listScope.Visit, node.List)

		if listScope.evalType == nil {
			return types.Unknown{}, nil
		}

		return listScope.evalType, nil
	}

	return nil, NewError(expr, "expression has no type")
}
