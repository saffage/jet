package parser

import (
	"errors"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser/token"
)

type parseFunc func() (ast.Node, error)

// TODO this method must be removed because it makes parser overcomplicated
func (parse *parser) or(a, b parseFunc) parseFunc {
	return func() (ast.Node, error) {
		tok, current := parse.tok, parse.current

		var nodeA, nodeB ast.Node
		var errA, errB error

		if nodeA, errA = a(); errA == nil {
			return nodeA, nil
		}

		parse.tok, parse.current = tok, current

		if nodeB, errB = b(); errB == nil {
			return nodeB, nil
		}

		parse.tok, parse.current = tok, current
		return nil, errors.Join(errA, errB)
	}
}

func (parse *parser) decl() (ast.Node, error) {
	switch parse.tok.Kind {
	case token.KwLet:
		return parse.letDecl()

	case token.KwType:
		return parse.typeDecl()

	default:
		return nil, parse.error(ErrExpectedDecl)
	}
}

func (parse *parser) variable() (ast.Node, error) {
	var (
		name ast.Ident
		ty   ast.Node
		err  error
	)

	if name, err = parse.ident(); err != nil {
		return nil, err
	}

	if parse.matchAny(token.Type, token.TypeVar, token.LParen) {
		if ty, err = parse.typeExprOrSignature(); err != nil {
			return nil, err
		}
	}

	return &ast.Decl{Name: name, Type: ty}, nil
}

func (parse *parser) variant() (ast.Node, error) {
	var (
		name   *ast.Upper
		params *ast.Parens
		err    error
	)

	if upper, err := parse.upper(); err != nil {
		return nil, err
	} else {
		name = upper.(*ast.Upper)
	}

	if parse.match(token.LParen) {
		if params, err = parse.parens(parse.labeledExpr(parse.typeExprOrSignature)); err != nil {
			return nil, err
		}
	}

	return &ast.Variant{Name: name, Params: params}, nil
}

func (parse *parser) parameter() (ast.Node, error) {
	var (
		name, _ = parse.ident()
		ty      ast.Node
		tok     token.Token
		err     error
	)

	if parse.matchAny(token.Type, token.TypeVar, token.LParen) {
		if ty, err = parse.typeExprOrSignature(); err != nil {
			return nil, err
		}
	} else if name == nil {
		if parse.match(token.KwType) {
			tok = parse.next()

			if name, err = parse.typeVar(); err != nil {
				return nil, parse.error(ErrExpectedTypeVar)
			}

			if parse.matchAny(token.Type, token.TypeVar) {
				if ty, err = parse.typeExpr(); err != nil {
					return nil, err
				}
			}
		} else {
			return nil, parse.error(ErrExpectedDecl)
		}
	}

	return &ast.Decl{TypeTok: tok.StartPos(), Name: name, Type: ty}, nil
}

func (parse *parser) typeVariable() (ast.Node, error) {
	var (
		name ast.Ident
		ty   ast.Node
		err  error
	)

	if name, err = parse.typeVar(); err != nil {
		return nil, err
	}

	if parse.matchAny(token.Type, token.TypeVar, token.LParen) {
		// type constraint
		if ty, err = parse.typeExpr(); err != nil {
			return nil, err
		}
	}

	decl := &ast.Decl{Name: name, Type: ty}

	if tok, ok := parse.take(token.Eq); ok {
		var tyDefault ast.Node

		if tyDefault, err = parse.typeExprOrSignature(); err != nil {
			return nil, err
		}

		return &ast.Op{
			X:    decl,
			Y:    tyDefault,
			Rng:  tok.Range,
			Kind: ast.OperatorAssign,
		}, nil
	}

	return decl, nil
}

func (parse *parser) labeledExpr(f parseFunc) parseFunc {
	return func() (ast.Node, error) {
		var (
			label   *ast.Lower
			expr    ast.Node
			err     error
			isShort bool
		)

		if parse.matchSequence(token.Name, token.Colon) {
			name, _ := parse.lower()
			label, _ = name.(*ast.Lower)
			parse.next()
		} else if parse.match(token.Colon) {
			isShort = true
			parse.next()
		} else {
			return f()
		}

		if expr, err = f(); err != nil {
			return nil, err
		}

		node := &ast.Label{Name: label, X: expr}

		if isShort && node.Label() == nil {
			return nil, parse.errorf(
				ErrExpectedIdent,
				"expected label name after the colon",
			)
		}

		return node, nil
	}
}

func (parse *parser) externOr(f parseFunc) parseFunc {
	return func() (ast.Node, error) {
		if tok, ok := parse.take(token.KwExtern); ok {
			var args *ast.Parens

			if parse.match(token.LParen) {
				parens, err := parse.args()
				if err != nil {
					return nil, err
				}
				args = parens.(*ast.Parens)
			}

			return &ast.Extern{TokPos: tok.StartPos(), Args: args}, nil
		}

		return f()
	}
}

func (parse *parser) typeVariantOrField() (ast.Node, error) {
	switch parse.tok.Kind {
	case token.Name, token.Colon:
		return parse.labeledExpr(parse.variable)()

	case token.Type:
		return parse.variant()

	default:
		return nil, parse.errorf(ErrUnexpectedToken, "expected name or type")
	}
}

func (parse *parser) letDecl() (ast.Node, error) {
	var (
		letTok token.Token
		decl   ast.Node
		expr   ast.Node
		err    error
	)

	if letTok, err = parse.expect(token.KwLet); err != nil {
		return nil, err
	}

	if decl, err = parse.variable(); err != nil {
		return nil, err
	}

	if _, err = parse.expect(token.Eq); err != nil {
		return nil, err
	}

	if expr, err = parse.externOr(parse.expr)(); err != nil {
		return nil, err
	}

	return &ast.LetDecl{
		LetTok: letTok.StartPos(),
		Decl:   decl.(*ast.Decl),
		Value:  expr,
	}, nil
}

func (parse *parser) typeDecl() (ast.Node, error) {
	var (
		typeTok token.Token
		eqTok   token.Token
		name    ast.Ident
		args    *ast.Parens
		expr    ast.Node
		err     error
	)

	if typeTok, err = parse.expect(token.KwType); err != nil {
		return nil, err
	}

	if name, err = parse.upper(); err != nil {
		return nil, err
	}

	if parse.tok.Kind == token.LParen {
		var parenList ast.Node

		if parenList, err = parse.parens(parse.typeVariable); err != nil {
			return nil, err
		}

		args = parenList.(*ast.Parens)
	}

	switch parse.tok.Kind {
	case token.Eq:
		eqTok = parse.next()

		if expr, err = parse.externOr(parse.typeExprOrSignature)(); err != nil {
			return nil, err
		}

	case token.LCurly:
		if expr, err = parse.blockFunc(parse.typeVariantOrField); err != nil {
			return nil, err
		}

	default:
		return nil, parse.error(ErrExpectedTypeOrBlock)
	}

	return &ast.TypeDecl{
		TypeTok: typeTok.StartPos(),
		EqTok:   eqTok.StartPos(),
		Name:    name.(*ast.Upper),
		Args:    args,
		Expr:    expr,
	}, nil
}

func (parse *parser) underscore() (ast.Ident, error) {
	tok, err := parse.expect(token.Underscore)

	if err != nil {
		return nil, err
	}

	return &ast.Underscore{Data: tok.Data, Rng: tok.Range}, nil
}

func (parse *parser) lower() (ast.Ident, error) {
	tok, err := parse.expect(token.Name)

	if err != nil {
		return nil, parse.error(ErrExpectedIdent)
	}

	return &ast.Lower{Data: tok.Data, Rng: tok.Range}, nil
}

func (parse *parser) upper() (ast.Ident, error) {
	tok, err := parse.expect(token.Type)

	if err != nil {
		return nil, parse.error(ErrExpectedIdent)
	}

	return &ast.Upper{Data: tok.Data, Rng: tok.Range}, nil
}

func (parse *parser) typeVar() (ast.Ident, error) {
	tok, err := parse.expect(token.TypeVar)

	if err != nil {
		return nil, parse.error(ErrExpectedIdent)
	}

	return &ast.Lower{Data: tok.Data, Rng: tok.Range}, nil
}

func (parse *parser) literal() (ast.Node, error) {
	tok, ok := parse.takeAny(token.Int, token.Float, token.String)

	if !ok {
		return nil, parse.error(ErrExpectedOperand)
	}

	return &ast.Literal{
		Kind: literals[tok.Kind],
		Data: tok.Data,
		Rng:  tok.Range,
	}, nil
}

func (parse *parser) ident() (ast.Ident, error) {
	switch parse.tok.Kind {
	case token.Name:
		return parse.lower()

	case token.Underscore:
		return parse.underscore()

	default:
		return nil, parse.errorf(
			ErrUnexpectedToken,
			"expected %s or %s",
			token.Name.UserString(),
			token.Underscore.UserString(),
		)
	}
}

func (parse *parser) block() (ast.Node, error) {
	return parse.blockFunc(parse.exprOrDecl)
}

func (parse *parser) exprOrDecl() (ast.Node, error) {
	switch parse.tok.Kind {
	case token.KwLet:
		return parse.letDecl()

	case token.KwType:
		return parse.typeDecl()

	case token.KwWhen:
		return parse.whenExpr()

	default:
		binary := func() (ast.Node, error) { return parse.binaryExpr(2) }
		return parse.or(parse.function, binary)()
	}
}

func (parse *parser) blockFunc(f parseFunc) (*ast.Block, error) {
	if !parse.match(token.LCurly) {
		return nil, parse.error(ErrExpectedBlock)
	}

	list, err := parse.listOpenClose(f, token.LCurly, token.RCurly, 0)

	if err != nil {
		return nil, err
	}

	return &ast.Block{Nodes: list.nodes, Rng: list.rng}, nil
}

func (parse *parser) parens(f parseFunc) (*ast.Parens, error) {
	list, err := parse.listOpenClose(
		f,
		token.LParen,
		token.RParen,
		token.Comma,
	)

	if err != nil {
		return nil, err
	}

	return &ast.Parens{
		Nodes: list.nodes,
		Rng:   list.rng,
	}, nil
}

func (parse *parser) brackets(f parseFunc) (*ast.List, error) {
	list, err := parse.listOpenClose(
		f,
		token.LBracket,
		token.RBracket,
		token.Comma,
	)

	if err != nil {
		return nil, err
	}

	return &ast.List{
		Nodes: list.nodes,
		Rng:   list.rng,
	}, nil
}

func (parse *parser) args() (ast.Node, error) {
	return parse.parens(parse.expr)
}

func (parse *parser) typeArgs() (ast.Node, error) {
	return parse.parens(parse.typeExprOrSignature)
}

func (parse *parser) typeExpr() (ast.Node, error) {
	switch parse.tok.Kind {
	case token.Type:
		node, _ := parse.upper()

		if parse.match(token.LParen) {
			typeArgs, err := parse.typeArgs()

			if err != nil {
				return nil, err
			}

			return &ast.Call{
				X:    node,
				Args: typeArgs.(*ast.Parens),
			}, nil
		}

		return node, nil

	case token.TypeVar:
		tok := parse.next()

		return &ast.TypeVar{
			Data: tok.Data,
			Rng:  tok.Range,
		}, nil

	default:
		return nil, parse.errorf(
			ErrUnexpectedToken,
			"expected %s or %s",
			token.Type.UserString(),
			token.TypeVar.UserString(),
		)
	}
}

func (parse *parser) signature() (ast.Node, error) {
	params, err := parse.parens(parse.parameter)

	if err != nil {
		return nil, err
	}

	var result, effects ast.Node
	var withToken token.Pos

	if parse.matchAny(token.Type, token.TypeVar, token.LParen) {
		if result, err = parse.typeExprOrSignature(); err != nil {
			return nil, err
		}
	}

	if tok, ok := parse.take(token.KwWith); ok {
		withToken = tok.StartPos()

		if effects, err = parse.binaryExpr(2); err != nil {
			return nil, err
		}
	}

	return &ast.Signature{
		Params:  params,
		Result:  result,
		Effects: effects,
		WithTok: withToken,
	}, nil
}

func (parse *parser) typeExprOrSignature() (ast.Node, error) {
	switch parse.tok.Kind {
	case token.LParen:
		return parse.signature()

	case token.Type, token.TypeVar:
		return parse.typeExpr()

	default:
		return nil, parse.errorf(
			ErrUnexpectedToken,
			"expected %s, %s or function signature",
			token.Type.UserString(),
			token.TypeVar.UserString(),
		)
	}
}

func (parse *parser) function() (ast.Node, error) {
	var (
		params *ast.Parens
		body   ast.Node
		eq     token.Range
		err    error
	)

	if _, err := parse.expect(token.KwFn); err != nil {
		return nil, err
	}

	if x, err := parse.parens(parse.parameter); err != nil {
		return nil, err
	} else {
		params = x
	}

	if eqTok, err := parse.expect(token.Eq); err != nil {
		return nil, err
	} else {
		eq = eqTok.Range
	}

	if body, err = parse.expr(); err != nil {
		return nil, err
	}

	return &ast.Function{
		Params: params,
		Body:   body,
		EqTok:  eq,
	}, nil
}

func (parse *parser) expr() (ast.Node, error) {
	switch {
	case parse.match(token.KwWhen):
		return parse.whenExpr()

	case parse.match(token.KwFn):
		return parse.function()

	default:
		return parse.binaryExpr(2)
	}
}

func (parse *parser) whenExpr() (ast.Node, error) {
	var (
		whenTok token.Token
		expr    ast.Node
		body    *ast.Block
		err     error
	)

	if whenTok, err = parse.expect(token.KwWhen); err != nil {
		return nil, err
	}

	if expr, err = parse.expr(); err != nil {
		return nil, err
	}

	if body, err = parse.blockFunc(parse.whenClause); err != nil {
		return nil, err
	}

	return &ast.When{
		TokPos: whenTok.StartPos(),
		Expr:   expr,
		Body:   body,
	}, nil
}

func (parse *parser) whenClause() (ast.Node, error) {
	var (
		arrowTok token.Token
		pattern  ast.Node
		expr     ast.Node
		err      error
	)

	if pattern, err = parse.pattern(); err != nil {
		return nil, err
	}

	if arrowTok, err = parse.expect(token.FatArrow); err != nil {
		return nil, err
	}

	if expr, err = parse.expr(); err != nil {
		return nil, err
	}

	return &ast.Op{
		X:    pattern,
		Y:    expr,
		Rng:  arrowTok.Range,
		Kind: ast.OperatorFatArrow,
	}, nil
}

func (parse *parser) pattern() (ast.Node, error) {
	switch {
	case parse.match(token.Name):
		return parse.lower()

	case parse.match(token.Underscore):
		return parse.underscore()

	case parse.match(token.Type):
		ty, _ := parse.upper()

		if parse.match(token.LParen) {
			dot2OrLabeledExpr := parse.dot2(parse.labeledExpr(parse.pattern))

			if list, err := parse.parens(dot2OrLabeledExpr); err != nil {
				return nil, err
			} else {
				return &ast.Call{X: ty, Args: list}, nil
			}
		}

		return ty, nil

	case parse.match(token.LBracket):
		return parse.brackets(parse.dot2(parse.pattern))

	case parse.match(token.LParen):
		return parse.parens(parse.pattern)

	default:
		return nil, ErrExpectedPattern
	}
}

func (parse *parser) dot2(fallback parseFunc) parseFunc {
	return func() (ast.Node, error) {
		if tok, ok := parse.take(token.Dot2); ok {
			var ident ast.Ident

			if parse.matchAny(token.Name, token.Underscore) {
				var err error

				if ident, err = parse.ident(); err != nil {
					return nil, err
				}
			}

			return &ast.Op{
				Y:    ident,
				Rng:  tok.Range,
				Kind: ast.OperatorRangeInclusive,
			}, nil
		}

		return fallback()
	}
}

func (parse *parser) binaryExpr(precedence int) (ast.Node, error) {
	var err error
	var x ast.Node

	if x, err = parse.prefix(); err != nil {
		return nil, err
	}

	for oprKind, ok := operators[parse.tok.Kind]; ok &&
		precedences[parse.tok.Kind] >= precedence; {

		oprTok := parse.next()
		y, err := parse.binaryExpr(precedences[oprTok.Kind] + 1)

		if err != nil {
			return nil, err
		}

		x = &ast.Op{
			X:    x,
			Y:    y,
			Rng:  oprTok.Range,
			Kind: oprKind,
		}
	}

	return x, nil
}

func (parse *parser) prefix() (ast.Node, error) {
	if tok, ok := parse.take(token.Minus); ok {
		operand, err := parse.primary()

		if err != nil {
			return nil, err
		}

		return &ast.Op{
			Y:    operand,
			Rng:  tok.Range,
			Kind: ast.OperatorNeg,
		}, nil
	}

	return parse.primary()
}

func (parse *parser) primary() (ast.Node, error) {
	x, err := parse.operand()

	if err != nil {
		return nil, err
	}

	for {
		if err != nil {
			return nil, err
		}

		switch parse.tok.Kind {
		case token.Dot:
			x, err = parse.dotExpr(x)

		case token.LParen:
			var args ast.Node

			if args, err = parse.args(); err != nil {
				return nil, err
			}

			x = &ast.Call{X: x, Args: args.(*ast.Parens)}

		default:
			return x, nil
		}
	}
}

func (parse *parser) operand() (ast.Node, error) {
	switch parse.tok.Kind {
	case token.Name:
		return parse.lower()

	case token.Type:
		return parse.upper()

	case token.Int, token.Float, token.String:
		return parse.literal()

	case token.LCurly:
		return parse.block()

	case token.LBracket:
		return parse.brackets(parse.expr)

	default:
		return nil, parse.error(ErrExpectedOperand)
	}
}

///
///
///

func (parse *parser) dotExpr(x ast.Node) (ast.Node, error) {
	dotTok, err := parse.expect(token.Dot)

	if err != nil {
		return nil, err
	}

	var selector ast.Node

	switch {
	case parse.match(token.Name):
		selector, _ = parse.lower()

	case parse.match(token.LCurly):
		if selector, err = parse.block(); err != nil {
			return nil, err
		}

	case parse.match(token.LParen):
		if selector, err = parse.parens(parse.expr); err != nil {
			return nil, err
		}

	case parse.match(token.LBracket):
		if selector, err = parse.brackets(parse.expr); err != nil {
			return nil, err
		}

	default:
		return nil, parse.error(ErrExpectedIdent, "expected identifier, parenthesis, brackets or braces")
	}

	return &ast.Dot{
		X:      x,
		Y:      selector,
		DotPos: dotTok.StartPos(),
	}, nil
}

///
///
///

type list struct {
	nodes []ast.Node
	rng   token.Range
}

func (parse *parser) listOpenClose(
	f parseFunc,
	opening, closing, separator token.Kind,
) (*list, error) {
	openTok, ok := parse.take(opening)

	if !ok {
		return nil, parse.errorf(
			ErrUnexpectedToken,
			"expected %s",
			opening.UserString(),
		)
	}

	nodes, err := parse.listDelimiter(f, closing, separator)

	if err != nil {
		if err == ErrUnterminatedList {
			return nil, newError(err, openTok.Range)
		}

		return nil, err
	}

	closeTok, ok := parse.take(closing)

	if !ok {
		panic("unreachable")
	}

	return &list{nodes, openTok.StartPos().WithEnd(closeTok.EndPos())}, nil
}

func (parse *parser) listDelimiter(
	f parseFunc,
	delim, sep token.Kind,
) ([]ast.Node, error) {
	var nodes []ast.Node
	var errs []error

	// list(f) = f {separator f} [separator]
	for parse.tok.Kind != delim {
		// Possible cases:
		//  - empty list `{}`
		//  - trailing separator `{x,}`
		//  - unterminated list `{`
		if parse.tok.Kind == token.EOF {
			return nodes, ErrUnterminatedList
		}

		tok := parse.tok
		node, err := f()

		if err == nil {
			if sep == 0 || parse.consume(sep) || parse.match(delim) {
				nodes = append(nodes, node)
				continue
			}

			// The node is correct, but no separator\delimiter
			// was found. Report it and assign [ast.BadNode] instead.
			rng := token.RangeFrom(node.Pos(), node.PosEnd())
			err = newError(ErrUnterminatedExpr, rng)
		}

		// Something went wrong, advance to some delimiter and
		// continue parsing elements until we find the 'closing' token.
		parse.skipUntil(sep, delim)
		parse.take(sep)

		errs = append(errs, err)
		nodes = append(nodes, &ast.BadNode{DesiredRange: tok.Range})
	}

	return nodes, errors.Join(errs...)
}
