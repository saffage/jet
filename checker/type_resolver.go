package checker

import (
	"fmt"
	"math"
	"math/big"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

// Type checks 'expr' and returns its type.
// If error was occured, result is undefined
func (check *Checker) typeOfInternal(expr ast.Node) types.Type {
	switch node := expr.(type) {
	case nil:
		panic("got nil node for expr")

	case *ast.Decl:
		panic("unhandled declaration at " + expr.Pos().String())

	case *ast.BadNode,
		*ast.Comment,
		*ast.CommentGroup,
		*ast.Else,
		*ast.StmtList,
		*ast.List,
		*ast.AttributeList:
		panic("ill-formed AST")

	case *ast.Empty:
		return types.Unit

	case *ast.Ident:
		return check.typeOfIdent(node)

	case *ast.Literal:
		return check.typeOfLiteral(node)

	case *ast.Call:
		return check.typeOfCall(node)

	case *ast.Index:
		return check.typeOfIndex(node)

	case *ast.ArrayType:
		return check.typeOfArrayType(node)

	case *ast.Signature:
		return check.typeOfSignature(node)

	case *ast.Function:
		return check.typeOfFunction(node)

	case *ast.Dot:
		return check.typeOfDot(node)

	case *ast.Deref:
		return check.typeOfDeref(node)

	case *ast.Op:
		return check.typeOfOp(node)

	case *ast.BracketList:
		return check.typeOfBracketList(node)

	case *ast.ParenList:
		return check.typeOfParenList(node)

	case *ast.CurlyList:
		return check.typeOfCurlyList(node)

	case *ast.If:
		return check.typeOfIf(node)

	case *ast.While:
		return check.typeOfWhile(node)

	case *ast.For:
		return check.typeOfFor(node)

	case *ast.Defer:
		return check.typeOfDefer(node)

	case *ast.Return:
		check.errorf(node, "'return' statement are not implemented")
		return types.Unit

	// NOTE implementation of break & continue are not finished.
	case *ast.Break:
		if node.Label != nil {
			check.errorf(node.Label, "labels are not implemented")
		}
		return types.Unit

	case *ast.Continue:
		if node.Label != nil {
			check.errorf(node.Label, "labels are not implemented")
		}
		return types.Unit

	default:
		panic(fmt.Sprintf("type checking of %T is not implemented", expr))
	}
}

func (check *Checker) symbolOf(ident *ast.Ident) Symbol {
	if sym, _ := check.scope.Lookup(ident.Name); sym != nil {
		return sym
	}
	return check.module.SymbolOf(ident)
}

func (check *Checker) typeOfIdent(node *ast.Ident) types.Type {
	if sym := check.symbolOf(node); sym != nil {
		if sym.Type() != nil {
			check.newUse(node, sym)
			return sym.Type()
		}

		check.errorf(node, "expression has no type")
		return nil
	}

	return nil
}

func (check *Checker) typeOfLiteral(node *ast.Literal) types.Type {
	switch node.Kind {
	case ast.IntLiteral:
		return types.UntypedInt

	case ast.FloatLiteral:
		return types.UntypedFloat

	case ast.StringLiteral:
		return types.UntypedString

	default:
		panic(fmt.Sprintf("unhandled literal kind: '%s'", node.Kind.String()))
	}
}

func (check *Checker) typeOfCall(node *ast.Call) types.Type {
	if builtIn, _ := node.X.(*ast.BuiltIn); builtIn != nil {
		if tv := check.resolveBuiltInCall(builtIn, node); tv != nil {
			return tv.Type
		}

		return nil
	}

	tyOperand := check.typeOf(node.X)
	if tyOperand == nil {
		return nil
	}

	if fn := types.AsFunc(tyOperand); fn != nil {
		tArgs := types.SkipUntyped(check.typeOfParenList(node.Args))
		if tArgs == nil {
			return nil
		}

		if idx, err := fn.CheckArgs(tArgs.(*types.Tuple)); err != nil {
			n := ast.Node(node.Args)

			if idx < len(node.Args.Nodes) {
				n = node.Args.Nodes[idx]
			}

			check.errorf(n, err.Error())
			return nil
		}

		if fn.Result().Len() == 1 {
			return fn.Result().Types()[0]
		}

		return fn.Result()
	} else if tyStruct := types.AsStruct(types.SkipTypeDesc(tyOperand)); tyStruct != nil {
		check.structInit(node.Args, tyStruct)
		return tyStruct
	}

	check.errorf(node.X, "expression is not a function or a struct type (%s)", tyOperand)
	return nil
}

func (check *Checker) typeOfIndex(node *ast.Index) types.Type {
	t := check.typeOf(node.X)
	if t == nil {
		return nil
	}

	if t.Equals(types.Unit) {
		check.errorf(node.X, "expession is of type (unit) and cannot be indexed")
		return nil
	}

	if len(node.Args.Nodes) != 1 {
		check.errorf(node.Args.List, "expected 1 argument")
		return nil
	}

	tIndex := check.typeOf(node.Args.Nodes[0])
	if tIndex == nil {
		return nil
	}

	if array := types.AsArray(t); array != nil {
		if !tIndex.Equals(types.I32) {
			check.errorf(node.Args.Nodes[0], "expected type (i32) for index, got (%s) instead", tIndex)
			return nil
		}
		if !check.assignable(node.X) {
			check.errorf(node.X, "expression cannot be indexed")
			return nil
		}
		return array.ElemType()
	} else if tuple := types.AsTuple(t); tuple != nil {
		value := check.valueOf(node.Args.Nodes[0])
		if value == nil || value.Value == nil || value.Value.Kind() != constant.Int {
			check.errorf(node.Args.Nodes[0], "expected compile-time integer")
			return nil
		}

		index := constant.AsInt(value.Value)
		tupleLen := big.NewInt(int64(tuple.Len() - 1))

		if index.Sign() == -1 || index.Cmp(tupleLen) == 1 {
			check.errorf(node.Args.Nodes[0], "index must be in range 0..%d", tuple.Len()-1)
			return nil
		}

		return tuple.Types()[index.Int64()]
	}

	check.errorf(node.X, "expression is not an array or tuple")
	return nil
}

func (check *Checker) typeOfArrayType(node *ast.ArrayType) types.Type {
	if len(node.Args.Nodes) == 0 {
		check.errorf(node.Args, "slices are not implemented")
		return nil
	}

	if len(node.Args.Nodes) > 1 {
		check.errorf(node.Args, "expected 1 argument")
		return nil
	}

	value := check.valueOf(node.Args.Nodes[0])
	if value == nil {
		check.errorf(node.Args.Nodes[0], "array size cannot be infered")
		return nil
	}

	intValue := constant.AsInt(value.Value)
	if intValue == nil {
		check.errorf(node.Args.Nodes[0], "expected integer value for array size")
		return nil
	}

	if intValue.Sign() == -1 || intValue.Int64() > math.MaxInt {
		check.errorf(node.Args.Nodes[0], "size must be in range 0..9223372036854775807")
		return nil
	}

	elemType := check.typeOf(node.X)
	if elemType == nil {
		return nil
	}

	if !types.IsTypeDesc(elemType) {
		check.errorf(node.X, "expected type, got '%s'", elemType)
		return nil
	}

	size := int(intValue.Int64())
	t := types.NewArray(size, types.SkipTypeDesc(elemType))
	return types.NewTypeDesc(t)
}

func (check *Checker) typeOfSignature(node *ast.Signature) types.Type {
	tyParams := check.typeOfParenList(node.Params)
	if tyParams == nil {
		return nil
	}

	tyResult := types.Unit

	if node.Result != nil {
		tyResultActual := check.typeOf(node.Result)
		if tyResultActual == nil {
			return nil
		}

		if !types.IsTypeDesc(tyResultActual) &&
			!tyResultActual.Equals(types.Unit) {
			check.errorf(
				node.Result,
				"expected type, got value of type '%s' instead",
				tyResultActual,
			)
			return nil
		}

		tyResult = types.WrapInTuple(types.SkipTypeDesc(tyResultActual))
	}

	ty := types.NewFunc(
		types.AsTuple(types.SkipTypeDesc(tyParams)),
		tyResult,
		nil,
	)
	return types.NewTypeDesc(ty)
}

func (check *Checker) typeOfFunction(node *ast.Function) types.Type {
	check.errorf(node, "closures are not implemented")
	return nil
}

func (check *Checker) typeOfDot(node *ast.Dot) types.Type {
	// if ident, _ := node.X.(*ast.Ident); ident != nil {
	// 	if m, _ := check.symbolOf(ident).(*Module); m != nil {
	// 		if sym, _ := m.Scope.Lookup(node.Y.Name); sym != nil {
	// 			if sym.Type() == nil {
	// 				check.errorf(node.Y, "expression has no type")
	// 			}
	// 			return sym.Type()
	// 		}
	// 		check.errorf(
	// 			node.Y,
	// 			"identifier `%s` is not defined in the module `%s`",
	// 			node.Y,
	// 			m.Name(),
	// 		)
	// 		return nil
	// 	}
	// }

	tyOperand := check.typeOf(node.X)
	if tyOperand == nil {
		return nil
	}

	// TODO get symbol of the type.
	if typedesc := types.AsTypeDesc(tyOperand); typedesc != nil {
		switch t := typedesc.Base().Underlying().(type) {
		// case *types.Struct:
		// 	return check.structInit(node.Args, typedesc)

		case *types.Enum:
			return check.enumMember(node, t)

		default:
			panic("unreachable")
		}
	}

	if tyStruct := types.AsStruct(tyOperand); tyStruct != nil {
		return check.structMember(node, tyStruct)
	}

	check.errorf(node.X, "expected module or struct variable, got '%s' instead", tyOperand)
	return nil
}

func (check *Checker) typeOfDeref(node *ast.Deref) types.Type {
	tyOperand := check.typeOf(node.X)
	if tyOperand == nil {
		return nil
	}

	if ref := types.AsRef(tyOperand); ref != nil {
		return ref.Base()
	}

	check.errorf(node.X, "expression is not a pointer")
	return nil
}

func (check *Checker) typeOfOp(node *ast.Op) types.Type {
	if node.X != nil {
		tyOperandX := check.typeOf(node.X)
		if tyOperandX == nil {
			return nil
		}

		if node.Y != nil {
			tyOperandY := check.typeOf(node.Y)
			if tyOperandY == nil {
				return nil
			}

			return check.infix(node, tyOperandX, tyOperandY)
		}

		return check.postfix(node, tyOperandX)
	}
	if node.Y != nil {
		tyOperandY := check.typeOf(node.Y)
		if tyOperandY == nil {
			return nil
		}

		return check.prefix(node, tyOperandY)
	}

	panic("unreachable")
}

func (check *Checker) typeOfBracketList(node *ast.BracketList) types.Type {
	var tyElem types.Type

	for _, expr := range node.Nodes {
		ty := check.typeOf(expr)
		if ty == nil {
			return nil
		}

		if tyElem == nil {
			tyElem = ty
			continue
		}

		if !ty.Equals(tyElem) {
			check.errorf(expr, "expected type '%s' for element, got '%s' instead", tyElem, ty)
			return nil
		}
	}

	size := len(node.Nodes)
	return types.NewArray(size, tyElem)
}

// The result is always [*types.Tuple].
func (check *Checker) typeOfParenList(node *ast.ParenList) types.Type {
	if len(node.Nodes) == 0 {
		return types.Unit
	}

	ty := check.typeOf(node.Nodes[0])
	if ty == nil {
		return nil
	}

	var (
		tyElems    = make([]types.Type, 0, len(node.Nodes))
		isTypeDesc = types.IsTypeDesc(ty)
		wasError   = false
	)

	tyElems = append(tyElems, ty)

	for _, expr := range node.Nodes[1:] {
		if ty = check.typeOf(expr); ty == nil {
			wasError = true
			continue
		}

		if isTypeDesc {
			if !types.IsTypeDesc(ty) {
				wasError = true
				check.errorf(expr, "expected type, got value of type '%s' instead", ty)
				continue
			}
		} else {
			if types.IsTypeDesc(ty) {
				wasError = true
				check.errorf(expr, "expected expression, got type '%s' instead", ty)
				continue
			}
		}

		tyElems = append(tyElems, ty)
	}

	if wasError {
		return nil
	}

	return types.NewTuple(tyElems...)
}

func (check *Checker) typeOfCurlyList(node *ast.CurlyList) types.Type {
	local := NewScope(check.scope, "block")
	block := NewBlock(local)

	defer check.setScope(check.scope)
	check.scope = local
	report.TaggedDebugf("checker", "push %s", local.name)

	visitor := ast.Visitor(check.visitBlock(block))

	for _, node := range node.Nodes {
		visitor.WalkTopDown(node)
	}

	report.TaggedDebugf("checker", "pop %s", local.name)
	return block.t
}

func (check *Checker) typeOfIf(node *ast.If) types.Type {
	tCondition := check.typeOf(node.Cond)
	// Don't return if 'tCondition == nil', check the body.

	if tCondition != nil && !tCondition.Equals(types.Bool) {
		check.errorf(
			node.Cond,
			"expected type 'bool' for condition, got '%s' instead",
			tCondition,
		)
		// Don't return, check the body.
	}

	tBody := check.typeOf(node.Body)
	if tBody == nil {
		return nil
	}

	if node.Else != nil {
		if !check.typeOfElse(node.Else, tBody) {
			return nil
		}
	}

	return tBody
}

func (check *Checker) typeOfElse(node *ast.Else, tExpected types.Type) bool {
	tBody := check.typeOf(node.Body)
	if tBody == nil {
		return false
	}

	tTypedBody := types.SkipUntyped(tBody)
	if !tBody.Equals(tExpected) && !tTypedBody.Equals(tExpected) {
		// Find the last node in the body for better error message.
		lastNode := ast.Node(node.Body)

		switch body := node.Body.(type) {
		case *ast.CurlyList:
			lastNode = body.Nodes[len(body.Nodes)-1]

		case *ast.If:
			lastNode = body.Body.Nodes[len(body.Body.Nodes)-1]
		}

		check.errorf(
			lastNode,
			"all branches must have the same type with first branch (%s), got (%s) instead",
			tExpected,
			tBody,
		)
		return false
	}

	return true
}

func (check *Checker) typeOfWhile(node *ast.While) types.Type {
	tCond := check.typeOf(node.Cond)
	if tCond == nil {
		return nil
	}

	if !tCond.Equals(types.Bool) {
		check.errorf(node.Cond, "expected type 'bool' for condition, got (%s) instead", tCond)
		// Don't return, check the body.
	}

	tBody := check.typeOf(node.Body)
	if tBody == nil {
		return nil
	}

	if !tBody.Equals(types.Unit) {
		check.errorf(node.Body, "while loop body must have no type, but got (%s)", tBody)
		return nil
	}

	return types.Unit
}

func (check *Checker) typeOfFor(node *ast.For) (ty types.Type) {
	ty = types.Unit

	// TODO implement range types
	infix, _ := node.IterExpr.(*ast.Op)
	if infix == nil {
		check.errorf(node.IterExpr, "expected range expression")
		return
	}
	if infix.Kind != ast.OperatorRangeInclusive && infix.Kind != ast.OperatorRangeExclusive {
		check.errorf(infix, "expected range operator")
		return
	}
	if infix.X == nil || infix.Y == nil {
		check.errorf(node.IterExpr, "expected range expression")
		return
	}

	tyX := check.typeOf(infix.X)
	if tyX == nil {
		return
	}

	tyY := check.typeOf(infix.Y)
	if tyY == nil {
		return
	}

	if !tyY.Equals(tyX) {
		check.errorf(infix, "type mismatch (%s and %s)", infix.X, infix.Y)
		return
	}

	// TODO allow only integral types
	tyLoopVar := tyX

	if len(node.Decls.Nodes) > 1 {
		check.errorf(node.Decls.Nodes[1], "invalid loop variables count (expected 1)")
		return
	}

	loopVarDecl, _ := node.Decls.Nodes[0].(*ast.Decl)
	if loopVarDecl == nil {
		panic("unreachable")
	}
	if loopVarDecl.Type != nil {
		tyLoopVarExplicit := check.typeOf(loopVarDecl.Type)
		if tyLoopVarExplicit == nil {
			return
		}
		if !types.IsTypeDesc(tyLoopVarExplicit) {
			check.errorf(loopVarDecl.Type, "'%s' is not a type", loopVarDecl.Type)
			return
		}
		tyLoopVarExplicit = types.SkipTypeDesc(tyLoopVarExplicit)
		if !tyLoopVar.Equals(tyLoopVarExplicit) {
			check.errorf(
				infix,
				"type mismatch, expected '%s' for loop variable, got '%s' instead",
				tyLoopVarExplicit,
				tyLoopVar,
			)
			return
		}
		tyLoopVar = tyLoopVarExplicit
	}

	tyLoopVar = types.SkipUntyped(tyLoopVar)

	bodyScope := NewScope(check.scope, "loop body")
	loopVar := NewVar(bodyScope, tyLoopVar, loopVarDecl)
	bodyScope.Define(loopVar)
	check.newDef(loopVarDecl.Ident, loopVar)

	var tyBody types.Type
	{
		block := NewBlock(bodyScope)

		defer check.setScope(check.scope)
		check.scope = bodyScope
		report.TaggedDebugf("checker", "push %s", bodyScope.name)

		visitor := ast.Visitor(check.visitBlock(block))

		for _, node := range node.Body.Nodes {
			visitor.WalkTopDown(node)
		}

		report.TaggedDebugf("checker", "pop %s", bodyScope.name)
		tyBody = block.t
	}
	if tyBody == nil {
		return
	}
	if !tyBody.Equals(types.Unit) {
		check.errorf(node.Body, "body must have no type, but got '%s'", tyBody)
	}

	return
}

func (check *Checker) typeOfDefer(node *ast.Defer) (ty types.Type) {
	ty = types.Unit
	check.scope.defers = append(check.scope.defers, node)

	tyX := check.typeOf(node.X)
	if tyX == nil {
		return
	}

	return
}
