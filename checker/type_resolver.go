package checker

import (
	"fmt"
	"strconv"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/log"
	"github.com/saffage/jet/types"
)

// Return type is never nil, if no error.
func (scope *Scope) TypeOf(expr ast.Node) (types.Type, error) {
	switch node := expr.(type) {
	case *ast.BadNode:
		panic("ill-formed AST")

	case *ast.Empty:
		return types.Unit, nil

	case *ast.Ident:
		if sym := scope.SymbolOf(node); sym != nil {
			if sym.Type() == nil {
				return nil, NewErrorf(node, "expression `%s` has no type", node.Name)
			}

			return sym.Type(), nil
		}

		return nil, NewErrorf(node, "identifier `%s` is undefined", node.Name)

	case *ast.Operator:
		panic("todo")

	case *ast.Literal:
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

	case *ast.ArrayType:
		if len(node.Args.Exprs) != 1 {
			return nil, NewError(node.Args, "expected 1 argument")
		}

		size := uint(0)

		switch arg := node.Args.Exprs[0].(type) {
		case *ast.Literal:
			// TODO use [constant.Int].
			n, err := strconv.ParseUint(arg.Value, 0, 32)
			if err != nil {
				panic(err)
			}

			size = uint(n)

		default:
			return nil, NewError(arg, "expected integer literal for array size")
		}

		// if !types.Primitives[types.UntypedInt].Equals(nType) {
		// 	return nil, NewErrorf(
		// 		node.Args.Exprs[0],
		// 		"expected type '%s', got '%s' instead",
		// 		types.Primitives[types.UntypedInt],
		// 		nType,
		// 	)
		// }

		elemType, err := scope.TypeOf(node.X)
		if err != nil {
			return nil, err
		}

		if !types.IsTypeDesc(elemType.Underlying()) {
			return nil, NewErrorf(node.X, "expected type, got '%s'", elemType)
		}

		t := types.NewArray(size, types.SkipTypeDesc(elemType))
		return types.NewTypeDesc(t), nil

	case *ast.ParenList:
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

			if types.IsTypeDesc(t.Underlying()) {
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
				if !types.IsTypeDesc(t.Underlying()) {
					return nil, NewErrorf(expr, "expected type, got '%s' instead", t)
				}

				elemTypes = append(elemTypes, types.SkipTypeDesc(t))
			} else {
				if types.IsTypeDesc(t.Underlying()) {
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

	case *ast.BracketList:
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
		return types.NewArray(uint(size), elemType), nil

	case *ast.PrefixOp:
		switch node.Opr.Kind {
		case ast.OperatorNeg:
			type_, err := scope.TypeOf(node.X)
			if err != nil {
				return nil, err
			}

			if p, ok := type_.Underlying().(*types.Primitive); ok {
				switch p.Kind() {
				case types.UntypedInt, types.UntypedFloat, types.I32:
					return type_, nil
				}
			}

			return nil, NewErrorf(
				node.Opr,
				"operator '%s' is not defined for the type '%s'",
				node.Opr.Kind.String(),
				type_.String(),
			)

		case ast.OperatorNot:
			type_, err := scope.TypeOf(node.X)
			if err != nil {
				return nil, err
			}

			if p, ok := type_.Underlying().(*types.Primitive); ok {
				switch p.Kind() {
				case types.UntypedBool, types.Bool:
					return type_, nil
				}
			}

			return nil, NewErrorf(
				node.X,
				"operator '%s' is not defined for the type '%s'",
				node.Opr.Kind.String(),
				type_.String(),
			)

		case ast.OperatorAddr, ast.OperatorMutAddr:
			panic("not implemented")

		default:
			panic(fmt.Sprintf("unhandled prefix operator '%s'", node.Opr))
		}

	case *ast.InfixOp:
		x_type, err := scope.TypeOf(node.X)
		if err != nil {
			return nil, err
		}

		y_type, err := scope.TypeOf(node.Y)
		if err != nil {
			return nil, err
		}

		if !x_type.Equals(y_type) {
			return nil, NewErrorf(node, "type mismatch ('%s' and '%s')", x_type, y_type)
		}

		if p, ok := types.SkipAlias(x_type).Underlying().(*types.Primitive); ok {
			switch node.Opr.Kind {
			case ast.OperatorAdd, ast.OperatorSub, ast.OperatorMul, ast.OperatorDiv, ast.OperatorMod,
				ast.OperatorBitAnd, ast.OperatorBitOr, ast.OperatorBitXor, ast.OperatorBitShl, ast.OperatorBitShr:
				switch p.Kind() {
				case types.UntypedInt, types.UntypedFloat, types.I32:
					return x_type, nil
				}

			case ast.OperatorEq, ast.OperatorNe, ast.OperatorLt, ast.OperatorLe, ast.OperatorGt, ast.OperatorGe:
				switch p.Kind() {
				case types.UntypedBool, types.UntypedInt, types.UntypedFloat:
					return types.Primitives[types.UntypedBool], nil

				case types.Bool, types.I32:
					return types.Primitives[types.Bool], nil
				}

			case ast.OperatorAssign:
				// Assign expression has the same type with the left operand.
				// return x_type, nil

				return types.Unit, nil
			}
		}

		return nil, NewErrorf(
			node.Opr,
			"operator '%s' is not defined for the type '%s'",
			node.Opr.Kind.String(),
			x_type.String(),
		)

	case *ast.Call:
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

	case *ast.Index:
		t, err := scope.TypeOf(node.X)
		if err != nil {
			return nil, err
		}

		arr, ok := t.Underlying().(*types.Array)
		if !ok {
			return nil, NewError(node.X, "expression is not an array")
		}

		if len(node.Args.Exprs) != 1 {
			return nil, NewErrorf(node.Args.ExprList, "expected 1 argument")
		}

		i, err := scope.TypeOf(node.Args.Exprs[0])
		if err != nil {
			return nil, err
		}

		if !types.Primitives[types.I32].Equals(i) {
			return nil, NewErrorf(node.Args.Exprs[0], "expected type 'i32' for index, got '%s' instead", i)
		}

		return arr.ElemType(), nil

	case *ast.BuiltInCall:
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

	case *ast.CurlyList:
		block := NewBlock(scope)
		fmt.Printf(">>> push local\n")

		for _, node := range node.Nodes {
			if err := ast.WalkTopDown(block.visit, node); err != nil {
				return nil, err
			}
		}

		fmt.Printf(">>> pop local\n")

		return block.t, nil

		// if !types.IsUnknown(block.t) {
		// 	return block.t, nil
		// } else if len(node.Nodes) > 0 {
		// 	return nil, NewError(node.Nodes[len(node.Nodes)-1], "expression has no type")
		// } else {
		// 	return nil, NewError(node, "expression has no type")
		// }

	case *ast.If:
		// We checking the body type before condition for returning the body
		// type in case when the condition is not a boolean type expression.

		tBody, err := scope.TypeOf(node.Body)
		if err != nil {
			return nil, err
		}

		if node.Else != nil {
			tElse, err := scope.TypeOf(node.Else.Body)
			if err != nil {
				return tBody, err
			}

			lastNodeInBody := ast.Node(node.Else.Body)

			switch body := node.Else.Body.(type) {
			case *ast.CurlyList:
				lastNodeInBody = body.Nodes[len(body.Nodes)-1]

			case *ast.If:
				lastNodeInBody = body.Body.Nodes[len(body.Body.Nodes)-1]
			}

			if (tBody == nil && tElse != nil) ||
				(tBody != nil && !tBody.Equals(tElse)) {
				return nil, NewErrorf(
					lastNodeInBody,
					"all branches must have the same type with first branch (%s), but have type '%s'",
					tBody,
					tElse,
				)
			}
		}

		tCond, err := scope.TypeOf(node.Cond)
		if err != nil {
			return tBody, err
		}

		if !types.Primitives[types.Bool].Equals(tCond) {
			return tBody, NewErrorf(node.Cond, "expected type 'bool' for condition, got '%s' instead", tCond)
		}

		return tBody, nil

	case *ast.While:
		return nil, typeCheckWhile(node, scope)

	default:
		panic("not implemented")
	}

	log.Warn("node of type '%T' was skipped while type checking", expr)
	return types.Unit, nil
}

func (scope *Scope) SymbolOf(ident *ast.Ident) Symbol {
	if sym, _ := scope.Lookup(ident.Name); sym != nil {
		return sym
	}

	return nil
}

func typeCheckWhile(node *ast.While, scope *Scope) error {
	tBody, err := scope.TypeOf(node.Body)
	if err != nil && tBody != nil {
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
