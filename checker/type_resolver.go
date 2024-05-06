package checker

import (
	"fmt"
	"math"
	"strconv"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/types"
)

// Return type is never nil, if no error.
func (scope *Scope) TypeOf(expr ast.Node) (types.Type, error) {
	switch node := expr.(type) {
	case nil:
		panic("got nil not for expr")

	case ast.Decl:
		panic("declaration must be handled somewhere else")

	case *ast.BadNode,
		*ast.Comment,
		*ast.CommentGroup,
		*ast.Else,
		*ast.List,
		*ast.ExprList,
		*ast.AttributeList:
		// *ast.Signature:
		panic("ill-formed AST")

	case *ast.Empty:
		return types.Unit, nil

	case *ast.Ident:
		return typeCheckIdent(node, scope)

	case *ast.Literal:
		return typeCheckLiteral(node)

	// case *ast.Operator:
	// 	panic("not implemented")

	case *ast.BuiltInCall:
		return typeCheckBuiltInCall(node, scope)

	case *ast.Call:
		return typeCheckCall(node, scope)

	case *ast.Index:
		return typeCheckIndex(node, scope)

	case *ast.ArrayType:
		return typeCheckArrayType(node, scope)

	case *ast.Signature:
		return typeCheckSignature(node, scope)

	case *ast.PrefixOp:
		return typeCheckPrefixOp(node, scope)

	case *ast.InfixOp:
		return typeCheckInfixOp(node, scope)

	case *ast.PostfixOp:
		return typeCheckPostfixOp(node, scope)

	case *ast.BracketList:
		return typeCheckBracketList(node, scope)

	case *ast.ParenList:
		return typeCheckParenList(node, scope)

	case *ast.CurlyList:
		return typeCheckCurlyList(scope, node)

	case *ast.If:
		return typeCheckIf(node, scope)

	case *ast.While:
		return nil, typeCheckWhile(node, scope)

	// case *ast.Return, *ast.Break, *ast.Continue:
	// 	panic("not implemented")

	default:
		panic(fmt.Sprintf("type checking of %T is not implemented", expr))
	}
}

// For the `_` identifier the result is (nil, nil). This is the
// only 1 way to get this result.
func (scope *Scope) ValueOf(expr ast.Node) (*Value, error) {
	switch node := expr.(type) {
	case *ast.Literal:
		value := constant.FromNode(node)
		type_ := types.FromConstant(value)
		return &Value{
			Type:  type_,
			Value: value,
		}, nil

	case *ast.Ident:
		if node.Name == "_" {
			return nil, nil
		}

		if sym, _ := scope.Lookup(node.Name); sym != nil {
			panic("constants are not implemented")
		}

		return nil, NewErrorf(node, "identifier `%s` is undefined", node.Name)

	case *ast.PrefixOp, *ast.PostfixOp, *ast.InfixOp:
		panic("not implemented")
	}

	return nil, NewError(expr, "expression is not a constant value")
}

func (scope *Scope) SymbolOf(ident *ast.Ident) Symbol {
	if sym, _ := scope.Lookup(ident.Name); sym != nil {
		return sym
	}

	return nil
}

func typeCheckIdent(node *ast.Ident, scope *Scope) (types.Type, error) {
	if sym := scope.SymbolOf(node); sym != nil {
		if sym.Type() == nil {
			return nil, NewErrorf(node, "expression `%s` has no type", node.Name)
		}

		return sym.Type(), nil
	}

	return nil, NewErrorf(node, "identifier `%s` is undefined", node.Name)
}

func typeCheckLiteral(node *ast.Literal) (types.Type, error) {
	switch node.Kind {
	case ast.IntLiteral:
		return types.Primitives[types.UntypedInt], nil

	case ast.FloatLiteral:
		return types.Primitives[types.UntypedFloat], nil

	case ast.StringLiteral:
		return types.Primitives[types.UntypedString], nil

	default:
		panic(fmt.Sprintf("unhandled literal kind: '%s'", node.Kind.String()))
	}
}

func typeCheckBuiltInCall(node *ast.BuiltInCall, scope *Scope) (types.Type, error) {
	var builtIn *BuiltIn

	for _, b := range builtIns {
		if b.name == node.Name.Name {
			builtIn = b
		}
	}

	if builtIn == nil {
		return nil, NewErrorf(node.Name, "unknown built-in function '@%s'", node.Name.Name)
	}

	args, ok := node.Args.(*ast.ParenList)
	if !ok {
		return nil, NewError(node.Args, "block as built-in function argument is not yet supported")
	}

	argTypes, err := scope.TypeOf(args)
	if err != nil {
		return nil, err
	}

	if idx, err := builtIn.t.CheckArgs(argTypes.(*types.Tuple)); err != nil {
		n := ast.Node(args)

		if idx < len(args.Exprs) {
			n = args.Exprs[idx]
		}

		return nil, NewErrorf(n, err.Error())
	}

	value, err := builtIn.f(args, scope)
	if err != nil {
		return nil, err
	}

	if value != nil {
		return value.Type, nil
	}

	return types.Unit, nil
}

func typeCheckCall(node *ast.Call, scope *Scope) (types.Type, error) {
	t, err := scope.TypeOf(node.X)
	if err != nil {
		return nil, err
	}

	fn, ok := t.Underlying().(*types.Func)
	if !ok {
		return nil, NewError(node.X, "expression is not a function")
	}

	argTypes, err := scope.TypeOf(node.Args)
	if err != nil {
		return nil, err
	}

	if idx, err := fn.CheckArgs(argTypes.(*types.Tuple)); err != nil {
		n := ast.Node(node.Args)

		if idx < len(node.Args.Exprs) {
			n = node.Args.Exprs[idx]
		}

		return nil, NewErrorf(n, err.Error())
	}

	return fn.Result(), nil
}

func typeCheckIndex(node *ast.Index, scope *Scope) (types.Type, error) {
	t, err := scope.TypeOf(node.X)
	if err != nil {
		return nil, err
	}

	if len(node.Args.Exprs) != 1 {
		return nil, NewErrorf(node.Args.ExprList, "expected 1 argument")
	}

	i, err := scope.TypeOf(node.Args.Exprs[0])
	if err != nil {
		return nil, err
	}

	if array := types.AsArray(t); array != nil {
		if !types.Primitives[types.I32].Equals(i) {
			return nil, NewErrorf(node.Args.Exprs[0], "expected type 'i32' for index, got '%s' instead", i)
		}

		return array.ElemType(), nil
	} else if tuple := types.AsTuple(t); tuple != nil {

		index := uint64(0)

		if lit, _ := node.Args.Exprs[0].(*ast.Literal); lit != nil && lit.Kind == ast.IntLiteral {
			n, err := strconv.ParseInt(lit.Value, 0, 64)
			if err != nil {
				panic(err)
			}

			if n < 0 || n > int64(tuple.Len())-1 {
				return nil, NewErrorf(node.Args.Exprs[0], "index must be in range 0..%d", tuple.Len()-1)
			}

			index = uint64(n)
		} else {
			return nil, NewError(node.Args.Exprs[0], "expected integer literal")
		}

		return tuple.Types()[index], nil
	} else {
		return nil, NewError(node.X, "expression is not an array or tuple")
	}
}

func typeCheckArrayType(node *ast.ArrayType, scope *Scope) (types.Type, error) {
	if len(node.Args.Exprs) == 0 {
		return nil, NewError(node.Args, "slices are not implemented")
	}

	if len(node.Args.Exprs) > 1 {
		return nil, NewError(node.Args, "expected 1 argument")
	}

	value, err := scope.ValueOf(node.Args.Exprs[0])
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, NewError(node.Args.Exprs[0], "array size cannot be infered")
	}

	intValue := constant.AsInt(value.Value)
	if intValue == nil {
		return nil, NewError(node.Args.Exprs[0], "expected integer value for array size")
	}

	if intValue.Sign() == -1 || intValue.Int64() > math.MaxInt {
		return nil, NewErrorf(node.Args.Exprs[0], "size must be in range 0..9223372036854775807")
	}

	elemType, err := scope.TypeOf(node.X)
	if err != nil {
		return nil, err
	}

	if !types.IsTypeDesc(elemType) {
		return nil, NewErrorf(node.X, "expected type, got '%s'", elemType)
	}

	size := int(intValue.Int64())
	t := types.NewArray(size, types.SkipTypeDesc(elemType))
	return types.NewTypeDesc(t), nil
}

func typeCheckSignature(node *ast.Signature, scope *Scope) (types.Type, error) {
	params, err := scope.TypeOf(node.Params)
	if err != nil {
		return nil, err
	}

	result := types.Unit

	if node.Result != nil {
		tResult, err := scope.TypeOf(node.Result)
		if err != nil {
			return nil, err
		}

		if !types.IsTypeDesc(tResult) {
			return nil, NewErrorf(
				node.Result,
				"expected type, got (%s) instead",
				tResult,
			)
		}

		result = types.WrapInTuple(types.SkipTypeDesc(tResult))
	}

	// 'param' should be a [*types.Tuple] because 'node.Params' is a [*ast.ParenList].
	t := types.NewFunc(result, params.(*types.Tuple))
	return types.NewTypeDesc(t), nil
}

func typeCheckPrefixOp(node *ast.PrefixOp, scope *Scope) (types.Type, error) {
	tOperand, err := scope.TypeOf(node.X)
	if err != nil {
		return nil, err
	}

	switch node.Opr.Kind {
	case ast.OperatorNeg:
		if p := types.AsPrimitive(tOperand); p != nil {
			switch p.Kind() {
			case types.UntypedInt, types.UntypedFloat, types.I32:
				return tOperand, nil
			}
		}

		return nil, NewErrorf(
			node.Opr,
			"operator '%s' is not defined for the type (%s)",
			node.Opr.Kind.String(),
			tOperand.String(),
		)

	case ast.OperatorNot:
		if p, ok := tOperand.Underlying().(*types.Primitive); ok {
			switch p.Kind() {
			case types.UntypedBool, types.Bool:
				return tOperand, nil
			}
		}

		return nil, NewErrorf(
			node.X,
			"operator '%s' is not defined for the type (%s)",
			node.Opr.Kind.String(),
			tOperand.String(),
		)

	case ast.OperatorAddr:
		if types.IsTypeDesc(tOperand) {
			t := types.NewRef(types.SkipTypeDesc(tOperand))
			return types.NewTypeDesc(t), nil
		}

		return types.NewRef(types.SkipUntyped(tOperand)), nil

	case ast.OperatorMutAddr:
		panic("not implemented")

	default:
		panic("unreachable")
	}
}

func typeCheckInfixOp(node *ast.InfixOp, scope *Scope) (types.Type, error) {
	tOperandX, err := scope.TypeOf(node.X)
	if err != nil {
		return nil, err
	}

	tOperandY, err := scope.TypeOf(node.Y)
	if err != nil {
		return nil, err
	}

	if !tOperandX.Equals(tOperandY) {
		return nil, NewErrorf(node, "type mismatch (%s and %s)", tOperandX, tOperandY)
	}

	if p, ok := types.SkipAlias(tOperandX).Underlying().(*types.Primitive); ok {
		switch node.Opr.Kind {
		case ast.OperatorAdd, ast.OperatorSub, ast.OperatorMul, ast.OperatorDiv, ast.OperatorMod,
			ast.OperatorBitAnd, ast.OperatorBitOr, ast.OperatorBitXor, ast.OperatorBitShl, ast.OperatorBitShr:
			switch p.Kind() {
			case types.UntypedInt, types.UntypedFloat, types.I32:
				return tOperandX, nil
			}

		case ast.OperatorEq, ast.OperatorNe, ast.OperatorLt, ast.OperatorLe, ast.OperatorGt, ast.OperatorGe:
			switch p.Kind() {
			case types.UntypedBool, types.UntypedInt, types.UntypedFloat:
				return types.Primitives[types.UntypedBool], nil

			case types.Bool, types.I32:
				return types.Primitives[types.Bool], nil
			}

		default:
			panic("unreachable")
		}
	}

	if node.Opr.Kind == ast.OperatorAssign {
		return types.Unit, nil
	}

	return nil, NewErrorf(
		node.Opr,
		"operator '%s' is not defined for the type '%s'",
		node.Opr.Kind.String(),
		tOperandX.String(),
	)
}

func typeCheckPostfixOp(node *ast.PostfixOp, scope *Scope) (types.Type, error) {
	tOperand, err := scope.TypeOf(node.X)
	if err != nil {
		return nil, err
	}

	switch node.Opr.Kind {
	case ast.OperatorUnwrap:
		if ref := types.AsRef(tOperand); ref != nil {
			return ref.Base(), nil
		}

		return nil, NewError(node.X, "expression is not a reference type")

	case ast.OperatorTry:
		panic("not inplemented")

	default:
		panic("unreachable")
	}
}

func typeCheckBracketList(node *ast.BracketList, scope *Scope) (types.Type, error) {
	var elemType types.Type

	for _, expr := range node.Exprs {
		t, err := scope.TypeOf(expr)
		if err != nil {
			return nil, err
		}

		if elemType == nil {
			elemType = types.SkipUntyped(t)
			continue
		}

		if !elemType.Equals(t) {
			return nil, NewErrorf(
				expr,
				"expected type '%s' for element, got '%s' instead",
				elemType,
				t,
			)
		}
	}

	size := len(node.Exprs)
	return types.NewArray(size, elemType), nil
}

func typeCheckParenList(node *ast.ParenList, scope *Scope) (types.Type, error) {
	// Either typedesc or tuple contructor.

	if len(node.Exprs) == 0 {
		return types.Unit, nil
	}

	elemTypes := []types.Type{}
	isTypeDescTuple := false

	{
		t, err := scope.TypeOf(node.Exprs[0])
		if err != nil {
			return nil, err
		}

		if types.IsTypeDesc(t) {
			isTypeDescTuple = true
			elemTypes = append(elemTypes, types.SkipTypeDesc(t))
		} else {
			elemTypes = append(elemTypes, types.SkipUntyped(t))
		}
	}

	for _, expr := range node.Exprs[1:] {
		t, err := scope.TypeOf(expr)
		if err != nil {
			return nil, err
		}

		if isTypeDescTuple {
			if !types.IsTypeDesc(t) {
				return nil, NewErrorf(expr, "expected type, got '%s' instead", t)
			}

			elemTypes = append(elemTypes, types.SkipTypeDesc(t))
		} else {
			if types.IsTypeDesc(t) {
				return nil, NewErrorf(expr, "expected expression, got type '%s' instead", t)
			}

			elemTypes = append(elemTypes, types.SkipUntyped(t))
		}
	}

	t := types.NewTuple(elemTypes...)

	if isTypeDescTuple {
		return types.NewTypeDesc(t), nil
	}

	return t, nil
}

func typeCheckCurlyList(scope *Scope, node *ast.CurlyList) (types.Type, error) {
	block := NewBlock(scope)
	fmt.Printf(">>> push local\n")

	for _, node := range node.Nodes {
		if err := ast.WalkTopDown(block.visit, node); err != nil {
			return nil, err
		}
	}

	fmt.Printf(">>> pop local\n")

	return block.t, nil
}

func typeCheckIf(node *ast.If, scope *Scope) (types.Type, error) {
	// We check the body type before the condition to return the
	// body type in case the condition is not a boolean expression.
	tBody, err := scope.TypeOf(node.Body)
	if err != nil {
		return nil, err
	}

	if node.Else != nil {
		if err := typeCheckElse(node.Else, tBody, scope); err != nil {
			return tBody, err
		}
	}

	tCondition, err := scope.TypeOf(node.Cond)
	if err != nil {
		return tBody, err
	}

	if !types.Primitives[types.Bool].Equals(tCondition) {
		return tBody, NewErrorf(
			node.Cond,
			"expected type (bool) for condition, got (%s) instead",
			tCondition,
		)
	}

	return tBody, nil
}

func typeCheckElse(node *ast.Else, expectedType types.Type, scope *Scope) error {
	tBody, err := scope.TypeOf(node.Body)
	if err != nil {
		return err
	}

	if !expectedType.Equals(tBody) {
		// Find the last node in the body for better error message.
		lastNode := ast.Node(node.Body)

		switch body := node.Body.(type) {
		case *ast.CurlyList:
			lastNode = body.Nodes[len(body.Nodes)-1]

		case *ast.If:
			lastNode = body.Body.Nodes[len(body.Body.Nodes)-1]
		}

		return NewErrorf(
			lastNode,
			"all branches must have the same type with first branch (%s), got (%s) instead",
			expectedType,
			tBody,
		)
	}

	return nil
}

func typeCheckWhile(node *ast.While, scope *Scope) error {
	tBody, err := scope.TypeOf(node.Body)
	if err != nil {
		return err
	}

	if !types.Unit.Equals(tBody) {
		return NewErrorf(node.Body, "while loop body must have no type, but body has type '%s'", tBody)
	}

	tCond, err := scope.TypeOf(node.Cond)
	if err != nil {
		return err
	}

	if !types.Primitives[types.Bool].Equals(tCond) {
		return NewErrorf(node.Cond, "expected type 'bool' for condition, got '%s' instead", tCond)
	}

	return nil
}
