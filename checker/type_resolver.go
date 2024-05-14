package checker

import (
	"fmt"
	"math"
	"math/big"
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

// Type checks 'expr' and returns its type.
// If error was occured, result is undefined
func (check *Checker) typeOfInternal(expr ast.Node) types.Type {
	switch node := expr.(type) {
	case nil:
		panic("got nil node for expr")

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
		return types.Unit

	case *ast.Ident:
		return check.typeOfIdent(node)

	case *ast.Literal:
		return check.typeOfLiteral(node)

	// case *ast.Operator:
	// 	panic("not implemented")

	case *ast.BuiltInCall:
		return check.typeOfBuiltInCall(node)

	case *ast.Call:
		return check.typeOfCall(node)

	case *ast.Index:
		return check.typeOfIndex(node)

	case *ast.ArrayType:
		return check.typeOfArrayType(node)

	case *ast.Signature:
		return check.typeOfSignature(node)

	case *ast.MemberAccess:
		return check.typeOfMemberAccess(node)

	case *ast.PrefixOp:
		return check.typeOfPrefixOp(node)

	case *ast.InfixOp:
		return check.typeOfInfixOp(node)

	case *ast.PostfixOp:
		return check.typeOfPostfixOp(node)

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

	// case *ast.Return, *ast.Break, *ast.Continue:
	// 	panic("not implemented")

	default:
		panic(fmt.Sprintf("type checking of %T is not implemented", expr))
	}
}

func (check *Checker) symbolOf(ident *ast.Ident) Symbol {
	if ident != nil {
		if sym := check.module.Defs[ident]; sym != nil {
			return sym
		}
		if sym := check.module.Uses[ident]; sym != nil {
			return sym
		}
		if sym, _ := check.scope.Lookup(ident.Name); sym != nil {
			return sym
		}
	}

	return nil
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

	check.errorf(node, "identifier `%s` is undefined", node.Name)
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

func (check *Checker) typeOfBuiltInCall(node *ast.BuiltInCall) types.Type {
	var builtIn *BuiltIn
	idx := slices.IndexFunc(builtIns, func(b *BuiltIn) bool {
		return b.name == node.Name.Name
	})

	if idx != -1 {
		builtIn = builtIns[idx]
	}

	if builtIn == nil {
		check.errorf(node.Name, "unknown built-in function '@%s'", node.Name.Name)
		return nil
	}

	args, _ := node.Args.(*ast.ParenList)
	if args == nil {
		check.errorf(node.Args, "block as built-in function argument is not yet supported")
		return nil
	}

	tArgList := check.typeOfParenList(args)
	if tArgList == nil {
		return nil
	}

	tArgs, _ := tArgList.(*types.Tuple)
	if tArgs == nil {
		return nil
	}

	if idx, err := builtIn.t.CheckArgs(tArgs); err != nil {
		n := ast.Node(args)

		if idx < len(args.Exprs) {
			n = args.Exprs[idx]
		}

		check.errorf(n, err.Error())
		return nil
	}

	vArgs := make([]*TypedValue, tArgs.Len())

	for i := range len(vArgs) {
		vArgs[i] = check.module.Types[args.Exprs[i]]
	}

	value, err := builtIn.f(args, vArgs)
	if err != nil {
		check.addError(err)
		return nil
	}
	if value == nil {
		return nil
	}

	return value.Type
}

func (check *Checker) typeOfCall(node *ast.Call) types.Type {
	tOperand := check.typeOf(node.X)
	if tOperand == nil {
		return nil
	}

	fn := types.AsFunc(tOperand)
	if fn == nil {
		check.errorf(node.X, "expression is not a function")
		return nil
	}

	tArgs := types.SkipUntyped(check.typeOfParenList(node.Args))
	if tArgs == nil {
		return nil
	}

	if idx, err := fn.CheckArgs(tArgs.(*types.Tuple)); err != nil {
		n := ast.Node(node.Args)

		if idx < len(node.Args.Exprs) {
			n = node.Args.Exprs[idx]
		}

		check.errorf(n, err.Error())
		return nil
	}

	return fn.Result()
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

	if len(node.Args.Exprs) != 1 {
		check.errorf(node.Args.ExprList, "expected 1 argument")
		return nil
	}

	tIndex := check.typeOf(node.Args.Exprs[0])
	if tIndex == nil {
		return nil
	}

	if array := types.AsArray(t); array != nil {
		if !types.I32.Equals(tIndex) {
			check.errorf(node.Args.Exprs[0], "expected type (i32) for index, got (%s) instead", tIndex)
			return nil
		}

		return array.ElemType()
	} else if tuple := types.AsTuple(t); tuple != nil {
		value := check.valueOf(node.Args.Exprs[0])
		if value == nil || value.Value == nil || value.Value.Kind() != constant.Int {
			check.errorf(node.Args.Exprs[0], "expected compile-time integer")
			return nil
		}

		index := constant.AsInt(value.Value)
		tupleLen := big.NewInt(int64(tuple.Len() - 1))

		if index.Sign() == -1 || index.Cmp(tupleLen) == 1 {
			check.errorf(node.Args.Exprs[0], "index must be in range 0..%d", tuple.Len()-1)
			return nil
		}

		return tuple.Types()[index.Int64()]
	}

	check.errorf(node.X, "expression is not an array or tuple")
	return nil
}

func (check *Checker) typeOfArrayType(node *ast.ArrayType) types.Type {
	if len(node.Args.Exprs) == 0 {
		check.errorf(node.Args, "slices are not implemented")
		return nil
	}

	if len(node.Args.Exprs) > 1 {
		check.errorf(node.Args, "expected 1 argument")
		return nil
	}

	value := check.valueOf(node.Args.Exprs[0])
	if value == nil {
		check.errorf(node.Args.Exprs[0], "array size cannot be infered")
		return nil
	}

	intValue := constant.AsInt(value.Value)
	if intValue == nil {
		check.errorf(node.Args.Exprs[0], "expected integer value for array size")
		return nil
	}

	if intValue.Sign() == -1 || intValue.Int64() > math.MaxInt {
		check.errorf(node.Args.Exprs[0], "size must be in range 0..9223372036854775807")
		return nil
	}

	elemType := check.typeOf(node.X)
	if elemType == nil {
		return nil
	}

	if !types.IsTypeDesc(elemType) {
		check.errorf(node.X, "expected type, got (%s)", elemType)
		return nil
	}

	size := int(intValue.Int64())
	t := types.NewArray(size, types.SkipTypeDesc(elemType))
	return types.NewTypeDesc(t)
}

func (check *Checker) typeOfSignature(node *ast.Signature) types.Type {
	tParams := check.typeOfParenList(node.Params)
	if tParams == nil {
		return nil
	}

	tResult := types.Unit

	if node.Result != nil {
		tActualResult := check.typeOf(node.Result)
		if tActualResult == nil {
			return nil
		}

		if !types.IsTypeDesc(tActualResult) {
			check.errorf(node.Result, "expected type, got (%s) instead", tActualResult)
			return nil
		}

		tResult = types.WrapInTuple(types.SkipTypeDesc(tActualResult))
	}

	t := types.NewFunc(tResult, tParams.(*types.Tuple))
	return types.NewTypeDesc(t)
}

func (check *Checker) typeOfMemberAccess(node *ast.MemberAccess) types.Type {
	if ident, _ := node.X.(*ast.Ident); ident != nil {
		if m, _ := check.symbolOf(ident).(*Module); m != nil {
			if member, _ := node.Selector.(*ast.Ident); member != nil {
				if sym, _ := m.Scope.Lookup(member.Name); sym != nil {
					if sym.Type() == nil {
						check.errorf(node.Selector, "expression has no type")
					}
					return sym.Type()
				}
				check.errorf(
					node.Selector,
					"identifier `%s` is not defined in the module `%s`",
					member,
					m.Name(),
				)
				return nil
			}
			check.errorf(node.Selector, "expected identifier in module member access expression")
			return nil
		}
	}

	tOperand := check.typeOf(node.X)
	if tOperand == nil {
		return nil
	}

	if typedesc := types.AsTypeDesc(tOperand); typedesc != nil {
		return check.structInit(node, typedesc)
	}

	if tStruct := types.AsStruct(tOperand); tStruct != nil {
		return check.structMember(node, tStruct)
	}

	return nil
}

func (check *Checker) typeOfPrefixOp(node *ast.PrefixOp) types.Type {
	tOperand := check.typeOf(node.X)
	if tOperand == nil {
		return nil
	}

	return check.prefix(node, tOperand)
}

func (check *Checker) typeOfInfixOp(node *ast.InfixOp) types.Type {
	tOperandX := check.typeOf(node.X)
	if tOperandX == nil {
		return nil
	}

	tOperandY := check.typeOf(node.Y)
	if tOperandY == nil {
		return nil
	}

	return check.infix(node, tOperandX, tOperandY)
}

func (check *Checker) typeOfPostfixOp(node *ast.PostfixOp) types.Type {
	check.errorf(node, "postfix operators are not supported")
	return nil
}

func (check *Checker) typeOfBracketList(node *ast.BracketList) types.Type {
	var elemType types.Type

	for _, expr := range node.Exprs {
		t := check.typeOf(expr)
		if t == nil {
			return nil
		}

		if elemType == nil {
			elemType = types.SkipUntyped(t)
			continue
		}

		if !elemType.Equals(t) {
			check.errorf(expr, "expected type (%s) for element, got (%s) instead", elemType, t)
			return nil
		}
	}

	size := len(node.Exprs)
	return types.NewArray(size, elemType)
}

// The result is always [*types.Tuple] or its typedesc.
func (check *Checker) typeOfParenList(node *ast.ParenList) types.Type {
	if len(node.Exprs) == 0 {
		return types.Unit
	}

	elemTypes := []types.Type{}
	isTypeDescTuple := false

	t := check.typeOf(node.Exprs[0])
	if t == nil {
		return nil
	}

	if types.IsTypeDesc(t) {
		isTypeDescTuple = true
		elemTypes = append(elemTypes, types.SkipTypeDesc(t))
	} else {
		elemTypes = append(elemTypes, t)
	}

	for _, expr := range node.Exprs[1:] {
		t := check.typeOf(expr)
		if t == nil {
			return nil
		}

		if isTypeDescTuple {
			if !types.IsTypeDesc(t) {
				check.errorf(expr, "expected type, got '%s' instead", t)
				return nil
			}

			elemTypes = append(elemTypes, types.SkipTypeDesc(t))
		} else {
			if types.IsTypeDesc(t) {
				check.errorf(expr, "expected expression, got type '%s' instead", t)
				return nil
			}

			elemTypes = append(elemTypes, t)
		}
	}

	if isTypeDescTuple {
		return types.NewTypeDesc(types.NewTuple(elemTypes...))
	}

	return types.NewTuple(elemTypes...)
}

func (check *Checker) typeOfCurlyList(node *ast.CurlyList) types.Type {
	local := NewScope(check.scope)
	block := NewBlock(local)

	defer check.setScope(check.scope)
	check.scope = local
	report.TaggedDebugf("checker", "push local scope")

	for _, node := range node.Nodes {
		ast.WalkTopDown(check.blockVisitor(block), node)
	}

	report.TaggedDebugf("checker", "pop local scope")
	return block.t
}

func (check *Checker) typeOfIf(node *ast.If) types.Type {
	tCondition := check.typeOf(node.Cond)
	// Don't return if 'tCondition == nil', check the body.

	if !types.Bool.Equals(tCondition) {
		check.errorf(
			node.Cond,
			"expected type (bool) for condition, got (%s) instead",
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

func (check *Checker) typeOfElse(node *ast.Else, expectedType types.Type) bool {
	tBody := check.typeOf(node.Body)
	if tBody == nil {
		return false
	}

	tTypedBody := types.SkipUntyped(tBody)
	if !expectedType.Equals(tBody) && !expectedType.Equals(tTypedBody) {
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
			expectedType,
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

	if !types.Bool.Equals(tCond) {
		check.errorf(node.Cond, "expected type 'bool' for condition, got (%s) instead", tCond)
		// Don't return, check the body.
	}

	tBody := check.typeOf(node.Body)
	if tBody == nil {
		return nil
	}

	if !types.Unit.Equals(tBody) {
		check.errorf(node.Body, "while loop body must have no type, but got (%s)", tBody)
		return nil
	}

	return types.Unit
}
