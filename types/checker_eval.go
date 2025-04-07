package types

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/saffage/jet/ast"
	. "github.com/saffage/jet/internal/debug"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

// Evaluates a value\type of node.
//
// Evaluation of node can be forces by setting [checker.forceEval] before.
//
// An optional dest parameter can be specified to check if the computed value
// matches the dest constraint, otherwise [errorTypeMismatch] error is returned
// with evaluated value.
//
// If error was occurred, result is undefined.
func (check *checker) eval(node ast.Node, expected ...Type) (*Value, error) {
	if len(expected) > 0 {
		return check.evalExpected(node, nil, expected[0], "")
	}

	return check.evalExpected(node, nil, nil, "")
}

func (check *checker) evalExpected(
	node, dest ast.Node,
	expected Type,
	reason string,
) (*Value, error) {
	value, err := check.evalNode(node)

	if err != nil {
		return nil, err
	}

	const assertMessage = "types.(*checker).eval: cannot resolve type of the expression"
	Assert(value != nil, assertMessage)
	Assert(value.T != nil, assertMessage)

	check.setValue(node, value)

	if expected != nil && !value.T.Equal(expected) {
		return value, errTypeMismatch(
			node,
			dest,
			reason,
			Render(value.T),
			expected,
		)
	}

	// TODO constant folding

	if check.flags&ForceEval != 0 && value.V == nil {
		return value, errExpectedConstantValue(node, value.T)
	}

	return value, err
}

func (check *checker) evalNode(node ast.Node) (*Value, error) {
	switch node := node.(type) {
	case nil:
		panic(errInternal(text.Span{}, "got nil node for expr"))

	case *ast.LetDecl, *ast.TypeAlias, *ast.TypeDef:
		panic(errInternal(node.Range(), "unhandled declaration"))

	case *ast.BadNode:
		if check.flags&IgnoreBadNodes != 0 {
			return &Value{T: NoneType}, nil
		}

		panic(errInternal(node.Range(), "ill-formed AST"))

	case *ast.Stmts:
		panic(errInternal(node.Range(), "ill-formed AST"))

	case *ast.Parens:
		panic(errIllFormedAst(node))

	case *ast.Placeholder:
		return &Value{T: NoneType}, nil

	case ast.Ident:
		return check.evalIdent(node)

	case *ast.Literal:
		return check.evalLiteral(node), nil

	case *ast.Call:
		return check.evalCall(node)

	case *ast.Dot:
		return check.resolveSelector(node)

	case *ast.Function:
		return check.evalFunction(node)

	case *ast.Op:
		return check.evalOp(node)

	case *ast.When:
		return check.evalWhen(node)

	case *ast.Block:
		return check.evalBlock(node)

	// case *ast.List:
	// 	return check.evalList(node)

	default:
		panic(fmt.Sprintf("types.(*checker).eval: type checking of %T is not implemented", node))
	}
}

func (check *checker) evalIdent(ident ast.Ident) (*Value, error) {
	symbol := check.symbolOf(ident)

	if symbol == nil {
		return nil, errUndefinedIdent(ident, nil)
	}

	if !IsResolved(symbol.Type()) {
		return nil, errInternal(ident.Range(), "`%s` is not resolved", symbol.Name())
	}

	check.newUse(ident, symbol)
	return symbol.Value(), nil
}

func (check *checker) evalLiteral(node *ast.Literal) *Value {
	switch node.Kind {
	case ast.IntLiteral:
		value := node.Value

		if suffixIdx := strings.LastIndex(node.Value, "'"); suffixIdx != -1 {
			value = value[:suffixIdx]
		}

		if value, ok := big.NewInt(0).SetString(value, 0); ok {
			return &Value{V: NewBigIntConstant(value)}
		}

		// Unreachable?
		panic(fmt.Sprintf("invalid integer value for constant: '%s'", value))

	case ast.FloatLiteral:
		value := node.Value

		if suffixIdx := strings.LastIndex(node.Value, "'"); suffixIdx != -1 {
			value = value[:suffixIdx]
		}

		if value, ok := big.NewFloat(0.0).SetString(value); ok {
			return &Value{V: NewBigFloatConstant(value)}
		}

		// Unreachable?
		panic(fmt.Sprintf("invalid float value for constant: '%s'", node.Value))

	case ast.StringLiteral:
		start := strings.IndexAny(node.Value, "\"'")
		end := strings.LastIndexAny(node.Value, "\"'")
		value := node.Value[start+1 : end]

		return &Value{V: NewStringConstant(value)}

	default:
		panic("unreachable")
	}
}

func (check *checker) evalCall(node *ast.Call) (*Value, error) {
	operand, err := check.eval(node.X)

	if err != nil {
		return nil, err
	}

	switch t := operand.T.(type) {
	case Descriptor:
		switch base := t.base.(type) {
		case *Custom:
			if base.FieldsLen() > 0 {
				panic("unimplemented: type construction with arguments")
			}

			return &Value{T: base}, nil

		case *Alias:

		}

	case *Fn:
		fn, err := check.resolveCall(node, t)
		return &Value{T: fn}, err
	}

	if Is[Descriptor](operand.T) {
		return nil, errTypeIsNotParametrized(node.X, operand.T)
	}

	return nil, errExprIsNotCallable(node.X, operand.T)
}

func (check *checker) evalFunction(node *ast.Function) (*Value, error) {
	fn, paramsEnv, err := check.resolveSignature(node.Signature)

	if err != nil {
		return nil, err
	}

	if node.Body == nil {
		return &Value{T: NewDescriptor(fn)}, nil
	}

	_ = paramsEnv
	panic("unimplemented: function literal")
}

func (check *checker) evalOp(node *ast.Op) (*Value, error) {
	if node.X != nil {
		x, err := check.eval(node.X)

		if err != nil {
			return nil, err
		}

		if node.Y != nil {
			y, err := check.eval(node.Y)

			if err != nil {
				return nil, err
			}

			return check.infix(node, x, y)
		}

		return check.postfix(node, x)
	}

	if node.Y != nil {
		y, err := check.eval(node.Y)

		if err != nil {
			return nil, err
		}

		return check.prefix(node, y)
	}

	// What about operator as identifier?
	panic("unreachable: operator doesn't have operands")
}

func (check *checker) evalWhen(node *ast.When) (*Value, error) {
	expr, err := check.eval(node.Expr)

	if err != nil {
		return nil, err
	}

	var errs []error
	var t Type
	var first ast.Node

	for _, case_ := range node.Body.Stmts.Items {
		op, _ := case_.(*ast.Case)

		if op == nil {
			panic(errIllFormedAst(case_))
		}

		case_, err := check.evalExpected(
			op.Expr,
			first,
			t,
			"expected because of this branch expression",
		)

		if err != nil {
			errs = append(errs, err)
			continue
		}

		if t == nil {
			t = IntoTyped(case_)
			first = op.Expr
		}

		// TODO implement pattern exhaustiveness checking
	}

	return expr, report.Join(errs...)
}

func (check *checker) evalBlock(node *ast.Block) (*Value, error) {
	value := &Value{T: NoneType} // TODO implement constant None value
	err := error(nil)

	defer func() {
		check.setEnv(check.env)
		report.Debug("pop block")
	}()

	check.env = NewEnv(ScopeEnv, check.env)
	report.Debug("push block")

loop:
	for _, node := range node.Stmts.Items {
		switch node := node.(type) {
		case *ast.LetDecl:
			check.resolveLetDecl(node)
			value.T = NoneType

		case *ast.TypeAlias:
			check.resolveTypeAlias(node)
			value.T = NoneType

		case *ast.TypeDef:
			check.resolveTypeDef(node)
			value.T = NoneType

		default:
			value, err = check.eval(node)

			if err != nil {
				break loop
			}
		}
	}

	return value, err
}

// func (check *checker) evalList(node *ast.List) (*Value, error) {
// 	var tListElem Type
// 	var errs []error

// 	for _, elem := range node.Nodes {
// 		value, err := check.eval(elem)

// 		if err != nil {
// 			continue
// 		}

// 		if tListElem == nil {
// 			tListElem = value.T
// 			continue
// 		}

// 		if !value.T.Equal(tListElem) {
// 			return nil, errElemTypeMismatch(
// 				elem,
// 				node.Nodes[0],
// 				value.T,
// 				tListElem,
// 			)
// 		}
// 	}

// 	size := len(node.Nodes)
// 	return &Value{T: NewFixedArray(size, tListElem)}, report.Join(errs...)
// }
