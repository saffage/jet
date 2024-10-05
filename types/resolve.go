package types

import (
	"errors"
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

// Type checks 'expr' and returns its type. If error was occurred,
// result is undefined.
func (check *checker) typeOfInternal(expr ast.Node) (t Type, err error) {
	switch node := expr.(type) {
	case nil:
		panic(errorf(nil, "got nil node for expr"))

	case *ast.LetDecl, *ast.TypeDecl:
		panic(errorf(expr, "unhandled declaration"))

	case *ast.BadNode, *ast.Stmts, *ast.AttributeList:
		panic(errorf(expr, "ill-formed AST"))

	case *ast.Empty:
		return NoneType, nil

	case *ast.Lower:
		if sym := check.symbolOf(node); sym != nil {
			if sym.Type() != nil {
				check.newUse(node, sym)
				t = sym.Type()
			} else {
				err = errorf(node, "symbols `%s` is not yet resolved", sym.Name())
			}
		}

	case *ast.Upper:
		if sym := check.typeSymbolOf(node); sym != nil {
			if sym.typedesc != nil {
				check.newUse(node, sym)
				t = sym.Type()
			} else {
				err = errorf(node, "symbols `%s` is not yet resolved", sym.Name())
			}
		}

	case *ast.Literal:
		return FromAst(node), nil

	case *ast.Call:
		return check.typeOfCall(node)

	case *ast.Dot:
		return check.typeOfDot(node)

	case *ast.Op:
		return check.typeOfOp(node)

	case *ast.When:
		return check.typeOfWhen(node)

	case *ast.Block:
		t = check.typeOfBlock(node)

	case *ast.List:
		return check.typeOfList(node)

	case *ast.Parens:
		panic(&errorIllFormedAst{node})

	default:
		panic(fmt.Sprintf("type checking of %T is not implemented", expr))
	}

	return t, err
}

func (check *checker) symbolOf(ident ast.Ident) Symbol {
	if sym, _ := check.env.Lookup(ident.String()); sym != nil {
		return sym
	}
	return check.module.SymbolOf(ident)
}

func (check *checker) typeSymbolOf(ident ast.Ident) *TypeDef {
	if t, _ := check.env.LookupType(ident.String()); t != nil {
		return t
	}
	return check.module.TypeSymbolOf(ident)
}

func (check *checker) typeOfCall(node *ast.Call) (Type, error) {
	tOperand, err := check.typeOf(node.X)
	if err != nil || tOperand == nil {
		return nil, err
	}

	if fn := As[*Function](tOperand); fn != nil {
		tParens, err := check.typeOfParens(node.Args)

		if err != nil || tParens == nil {
			return nil, err
		}

		for i := range tParens {
			if fn.params[i] == nil {
				// unresolved parameter
				continue
			}
			tParens[i] = IntoTyped(tParens[i], fn.params[i])
		}

		if idx, err := fn.CheckArgs(tParens); err != nil {
			n := ast.Node(node.Args)

			if idx < len(node.Args.Nodes) {
				n = node.Args.Nodes[idx]
			}

			return nil, errorf(n, "%s", err.Error())
		}

		return fn.Result(), nil
	}

	return nil, errorf(
		node.X,
		"expression is not a function or a constructor: %s", // "expression is neither a function nor a constructor: `%s`",
		tOperand,
	)
}

func (check *checker) typeOfDot(node *ast.Dot) (Type, error) {
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

	tOperand, err := check.typeOf(node.X)

	if err != nil || tOperand == nil {
		return nil, err
	}

	// TODO get symbol of the type.
	if typedesc := As[*TypeDesc](tOperand); typedesc != nil {
		switch typedesc.Base().Underlying().(type) {
		// case *Struct:
		// 	return check.structInit(node.Args, typedesc)

		default:
			panic("unreachable")
		}
	}

	// if tyStruct := AsStruct(tyOperand); tyStruct != nil {
	// 	return check.structMember(node, tyStruct)
	// }

	return nil, errorf(
		node.X,
		"expected module or struct variable, got `%s` instead",
		tOperand,
	)
}

func (check *checker) typeOfOp(node *ast.Op) (Type, error) {
	if node.X != nil {
		tOperandX, err := check.typeOf(node.X)

		if err != nil || tOperandX == nil {
			return nil, err
		}

		if node.Y != nil {
			tOperandY, err := check.typeOf(node.Y)

			if err != nil || tOperandY == nil {
				return nil, err
			}

			return check.infix(node, tOperandX, tOperandY)
		}

		return check.postfix(node, tOperandX), nil
	}

	if node.Y != nil {
		tOperandY, err := check.typeOf(node.Y)

		if err != nil || tOperandY == nil {
			return nil, err
		}

		return check.prefix(node, tOperandY), nil
	}

	panic("unreachable: operator doesn't have operands")
}

func (check *checker) typeOfList(node *ast.List) (Type, error) {
	tListElem := Type(nil)

	for _, elem := range node.Nodes {
		tElem, err := check.typeOf(elem)

		if err != nil || tElem == nil {
			continue
		}

		if tListElem == nil {
			tListElem = tElem
			continue
		}

		if !tElem.Equal(tListElem) {
			return nil, problem(
				errorf(
					elem,
					"expected type `%s` for this element, got `%s` instead",
					tListElem,
					tElem,
				),
				notef(node.Nodes[0], "because of this"),
			)
		}
	}

	size := len(node.Nodes)
	return NewFixedArray(size, tListElem), nil
}

func (check *checker) typeOfBlock(node *ast.Block) Type {
	block := &block{check, NoneType}

	defer check.setEnv(check.env)
	check.env = NewNamedEnv(check.env, "block")
	report.Debug("push %s", check.env.name)

	for _, node := range node.Stmts.Nodes {
		ast.WalkTopDown(node, block)
	}

	report.Debug("pop %s", check.env.name)
	return block.t
}

func (check *checker) typeOfWhen(node *ast.When) (Type, error) {
	tExpr, err := check.typeOf(node.Expr)

	if err != nil {
		return nil, err
	}

	assert(tExpr != nil)

	for _, case_ := range node.Body.Stmts.Nodes {
		op, ok := case_.(*ast.Op)

		if !ok || op.Kind != ast.OperatorFatArrow || op.X == nil || op.Y == nil {
			panic(&errorIllFormedAst{case_})
		}
	}

	return nil, nil
}

// func (check *checker) typeOfIf(node *ast.If) Type {
// 	tCondition := check.typeOf(node.Cond)
// 	// Don't return if 'tCondition == nil', check the body.

// 	if tCondition != nil && !tCondition.Equals(Bool) {
// 		check.errorf(
// 			node.Cond,
// 			"expected type 'bool' for condition, got '%s' instead",
// 			tCondition,
// 		)
// 		// Don't return, check the body.
// 	}

// 	tBody := check.typeOf(node.Body)
// 	if tBody == nil {
// 		return nil
// 	}

// 	if node.Else != nil {
// 		if !check.typeOfElse(node.Else, tBody) {
// 			return nil
// 		}
// 	}

// 	return tBody
// }

// func (check *checker) typeOfElse(node *ast.Else, tExpected Type) bool {
// 	tBody := check.typeOf(node.Body)
// 	if tBody == nil {
// 		return false
// 	}

// 	tTypedBody := SkipUntyped(tBody)
// 	if !tBody.Equals(tExpected) && !tTypedBody.Equals(tExpected) {
// 		// Find the last node in the body for better error message.
// 		lastNode := ast.Node(node.Body)

// 		switch body := node.Body.(type) {
// 		case *ast.CurlyList:
// 			lastNode = body.Nodes[len(body.Nodes)-1]

// 		case *ast.If:
// 			lastNode = body.Body.Nodes[len(body.Body.Nodes)-1]
// 		}

// 		check.errorf(
// 			lastNode,
// 			"all branches must have the same type with first branch (%s), got (%s) instead",
// 			tExpected,
// 			tBody,
// 		)
// 		return false
// 	}

// 	return true
// }

func (check *checker) typeOfParens(node *ast.Parens) (Params, error) {
	if len(node.Nodes) == 0 {
		return nil, nil
	}

	t, err := check.typeOf(node.Nodes[0])

	if err != nil || t == nil {
		return nil, err
	}

	tElems := append(make([]Type, 0, len(node.Nodes)), t)
	isTypeDesc := Is[*TypeDesc](t)
	errs := []error(nil)

	for _, expr := range node.Nodes[1:] {
		t, err := check.typeOf(expr)

		if err != nil {
			errs = append(errs, err)
			continue
		}
		if t == nil {
			continue
		}

		if isTypeDesc {
			if !Is[*TypeDesc](t) {
				errs = append(errs, errorf(expr, "expected type, got value of type '%s' instead", t))
				continue
			}
		} else if Is[*TypeDesc](t) {
			errs = append(errs, errorf(expr, "expected expression, got type '%s' instead", t))
			continue
		}

		tElems = append(tElems, t)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return tElems, nil
}
