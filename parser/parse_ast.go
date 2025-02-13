package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

//------------------------------------------------
// Primitives
//------------------------------------------------

func (parse *parser) decl() (ast.Node, error) {
	switch parse.Kind {
	case token.KwLet:
		return parse.letDecl()

	case token.KwType:
		return parse.typeDecl()

	default:
		return nil, errExpectedDecl(parse.Span)
	}
}

func (parse *parser) declOrExpr() (ast.Node, error) {
	switch parse.Kind {
	case token.KwLet:
		return parse.letDecl()

	case token.KwType:
		return parse.typeDecl()

	default:
		return parse.expr()
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

	if parse.isTypeStartOrSignature() {
		if ty, err = parse.typeOrSignature(); err != nil {
			return nil, err
		}
	}

	return &ast.Decl{Ident: name, Type: ty}, nil
}

func (parse *parser) variant() (ast.Node, error) {
	name := parse.UppercaseIdent()

	if name == nil {
		return nil, errUnexpectedToken(parse.Span, token.UppercaseIdent)
	}

	var (
		params *ast.Parens
		err    error
	)

	if parse.match(token.LParen) {
		if params, err = parse.parens(parse.labeled(parse.typeExpr)); err != nil {
			return nil, err
		}
	}

	return &ast.Variant{Name: name, Params: params}, nil
}

func (parse *parser) typeVariable() (ast.Node, error) {
	var (
		name ast.Ident
		ty   ast.Node
		err  error
	)

	if name, err = parse.lowerNode(); err != nil {
		return nil, err
	}

	if parse.isTypeStart() {
		// type constraint
		if ty, err = parse.typeExpr(); err != nil {
			return nil, err
		}
	}

	decl := &ast.Decl{Ident: name, Type: ty}

	if tok, ok := parse.consume(token.Eq); ok {
		var tyDefault ast.Node

		if tyDefault, err = parse.typeExpr(); err != nil {
			return nil, err
		}

		return &ast.Op{
			X:    decl,
			Y:    tyDefault,
			Span: tok.Span,
			Kind: ast.OperatorAssign,
		}, nil
	}

	return decl, nil
}

func (parse *parser) labeled(f parseFunc) parseFunc {
	return func() (ast.Node, error) {
		var (
			label   *ast.Lower
			expr    ast.Node
			err     error
			isShort bool
		)

		switch {
		// case parse.matchSeq(token.LowercaseIdent, token.Colon):
		// 	label = parse.LowercaseIdent()
		// 	parse.next()

		case parse.match(token.Colon):
			isShort = true
			parse.next()

		default:
			return f()
		}

		if expr, err = f(); err != nil {
			return nil, err
		}

		node := &ast.Label{Name: label, X: expr}

		if isShort && node.Label() == nil {
			return nil, errUnexpectedToken(
				parse.Span,
				"label name after the colon",
			)
		}

		return node, nil
	}
}

func (parse *parser) externOr(f parseFunc) parseFunc {
	return func() (ast.Node, error) {
		if tok, ok := parse.consume(token.KwExtern); ok {
			var args *ast.Parens

			if parse.match(token.LParen) {
				parens, err := parse.args()
				if err != nil {
					return nil, err
				}
				args = parens.(*ast.Parens)
			}

			return &ast.Extern{
				ExternTok: tok.Span.From,
				Args:      args,
			}, nil
		}

		return f()
	}
}

func (parse *parser) letDecl() (ast.Node, error) {
	var (
		letTok token.Token
		decl   *ast.Decl
		expr   ast.Node
		err    error
	)

	if letTok, err = parse.expect(token.KwLet); err != nil {
		return nil, err
	}

	if x, err := parse.variable(); err != nil {
		return nil, err
	} else {
		decl = x.(*ast.Decl)
	}

	if _, err = parse.expect(token.Eq); err != nil {
		return nil, err
	}

	if expr, err = parse.externOr(parse.expr)(); err != nil {
		return nil, err
	}

	return &ast.LetDecl{
		LetTok: letTok.Span.From,
		Decl:   decl,
		Value:  expr,
	}, nil
}

func (parse *parser) typeDecl() (ast.Node, error) {
	var (
		typeTok token.Token
		name    *ast.Upper
		args    *ast.Parens
		err     error
	)

	if typeTok, err = parse.expect(token.KwType); err != nil {
		return nil, err
	}

	if name = parse.UppercaseIdent(); name == nil {
		return nil, errUnexpectedToken(parse.Span, token.UppercaseIdent)
	}

	if parse.Kind == token.LParen {
		var parenList ast.Node

		if parenList, err = parse.parens(parse.typeVariable); err != nil {
			return nil, err
		}

		args = parenList.(*ast.Parens)
	}

	switch parse.Kind {
	case token.Eq:
		var expr ast.Node

		eqTok := parse.next()

		if expr, err = parse.externOr(parse.typeExpr)(); err != nil {
			return nil, err
		}

		return &ast.TypeAlias{
			TypeTok: typeTok.Span.From,
			EqTok:   eqTok.Span.From,
			Ident:   name,
			Args:    args,
			Expr:    expr,
		}, nil

	case token.LCurly:
		var body *ast.Block

		if body, err = parse.blockFunc(parse.typeVariantOrField); err != nil {
			return nil, err
		}

		return &ast.TypeDef{
			TypeTok: typeTok.Span.From,
			Ident:   name,
			Args:    args,
			Body:    body,
		}, nil

	default:
		panic("todo")
		// return nil, errExpectedTypeOrBlock(parse.Span)
	}
}

func (parse *parser) typeVariantOrField() (ast.Node, error) {
	switch parse.Kind {
	case token.LowercaseIdent, token.IdentPlaceholder, token.Colon:
		return parse.labeled(parse.variable)()

	case token.UppercaseIdent:
		return parse.variant()

	default:
		return nil, errUnexpectedToken(
			parse.Span,
			token.LowercaseIdent,
			token.UppercaseIdent,
			token.Colon,
		)
	}
}

func (parse *parser) IdentPlaceholder() (ast.Ident, error) {
	tok, err := parse.expect(token.IdentPlaceholder)

	if err != nil {
		return nil, err
	}

	return &ast.Placeholder{Data: tok.Data, Span: tok.Span}, nil
}

func (parse *parser) lowerNode() (ast.Ident, error) {
	tok, err := parse.expect(token.LowercaseIdent)

	if err != nil {
		return nil, errUnexpectedToken(parse.Span, token.LowercaseIdent)
	}

	return &ast.Lower{Data: tok.Data, Span: tok.Span}, nil
}

func (parse *parser) LowercaseIdent() *ast.Lower {
	tok, ok := parse.consume(token.LowercaseIdent)

	if !ok {
		return nil
	}

	return &ast.Lower{Data: tok.Data, Span: tok.Span}
}

func (parse *parser) upperNode() (ast.Ident, error) {
	tok, err := parse.expect(token.UppercaseIdent)

	if err != nil {
		return nil, errUnexpectedToken(parse.Span, token.UppercaseIdent)
	}

	return &ast.Upper{Data: tok.Data, Span: tok.Span}, nil
}

func (parse *parser) UppercaseIdent() *ast.Upper {
	tok, ok := parse.consume(token.UppercaseIdent)

	if !ok {
		return nil
	}

	return &ast.Upper{Data: tok.Data, Span: tok.Span}
}

func (parse *parser) literal() (ast.Node, error) {
	tok, ok := parse.consumeAny(token.Int, token.Float, token.String)

	if !ok {
		return nil, errExpectedOperand(parse.Span)
	}

	return &ast.Literal{
		Value: tok.Data,
		Span:  tok.Span,
		Kind:  literals[tok.Kind],
	}, nil
}

func (parse *parser) ident() (ast.Ident, error) {
	switch parse.Kind {
	case token.LowercaseIdent:
		return parse.lowerNode()

	case token.IdentPlaceholder:
		return parse.IdentPlaceholder()

	default:
		return nil, errUnexpectedToken(
			parse.Span,
			token.LowercaseIdent,
			token.IdentPlaceholder,
		)
	}
}

func (parse *parser) block() (ast.Node, error) {
	return parse.blockFunc(parse.declOrExpr)
}

func (parse *parser) blockFunc(f parseFunc) (*ast.Block, error) {
	if !parse.match(token.LCurly) {
		return nil, errExpectedBlock(parse.Span)
	}

	nodes, span, err := parse.listOpenClose(f, token.LCurly, token.RCurly, 0)

	if err != nil {
		return nil, err
	}

	return &ast.Block{
		Stmts: &ast.Stmts{Items: nodes},
		Span:  span,
	}, nil
}

func (parse *parser) parens(f parseFunc) (*ast.Parens, error) {
	nodes, span, err := parse.listOpenClose(
		f,
		token.LParen,
		token.RParen,
		token.Comma,
	)

	if err != nil {
		return nil, err
	}

	return &ast.Parens{
		Nodes: nodes,
		Span:  span,
	}, nil
}

func (parse *parser) brackets(f parseFunc) (*ast.List, error) {
	nodes, span, err := parse.listOpenClose(
		f,
		token.LBracket,
		token.RBracket,
		token.Comma,
	)

	if err != nil {
		return nil, err
	}

	return &ast.List{
		Nodes: nodes,
		Span:  span,
	}, nil
}

func (parse *parser) args() (ast.Node, error) {
	return parse.parens(parse.expr)
}

func (parse *parser) typeArgs() (ast.Node, error) {
	return parse.parens(parse.typeExpr)
}

func (parse *parser) isTypeStart() bool {
	return parse.matchAny(token.UppercaseIdent, token.LowercaseIdent, token.KwFn)
}

func (parse *parser) isTypeStartOrSignature() bool {
	return parse.matchAny(token.UppercaseIdent, token.LowercaseIdent, token.KwFn, token.LParen)
}

// `T | U | ...`
func (parse *parser) typeExpr() (ast.Node, error) {
	var expr ast.Node

	expr, err := parse.simpleTypeExpr()

	if err != nil {
		return nil, err
	}

	for {
		pipeTok, ok := parse.consume(token.Bar)

		if !ok {
			break
		}

		nodeStart := parse.Span.From
		node, err := parse.simpleTypeExpr()

		if err != nil {
			node = &ast.BadNode{DesiredPos: nodeStart}
		}

		expr = &ast.Op{
			X:    expr,
			Y:    node,
			Kind: ast.OperatorBitOr,
			Span: pipeTok.Span,
		}

		if err != nil {
			break
		}
	}

	return expr, nil
}

func (parse *parser) simpleTypeExpr() (ast.Node, error) {
	switch parse.Kind {
	case token.UppercaseIdent:
		node, _ := parse.upperNode()

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

	case token.LowercaseIdent:
		// TODO: allow type arguments?
		tok := parse.next()

		return &ast.Lower{
			Data: tok.Data,
			Span: tok.Span,
		}, nil

	case token.KwFn:
		return parse.functionType()

	default:
		return nil, errUnexpectedToken(
			parse.Span,
			token.UppercaseIdent,
			token.LowercaseIdent,
			token.KwFn,
		)
	}
}

func (parse *parser) signature(parseParamFunc parseFunc) parseFunc {
	return func() (ast.Node, error) {
		params, err := parse.parens(parseParamFunc)

		if err != nil {
			return nil, err
		}

		var result ast.Node

		if parse.isTypeStart() {
			if result, err = parse.typeExpr(); err != nil {
				return nil, err
			}
		}

		return &ast.Signature{
			Params: params,
			Result: result,
		}, nil
	}
}

func (parse *parser) typeOrSignature() (ast.Node, error) {
	switch {
	case parse.match(token.LParen):
		return parse.signature(parse.labeled(parse.variable))()

	case parse.isTypeStart():
		return parse.typeExpr()

	default:
		return nil, errUnexpectedToken(
			parse.Span,
			token.UppercaseIdent,
			token.LowercaseIdent,
			token.KwFn,
			"'(' for function signature",
		)
	}
}

func (parse *parser) functionType() (ast.Node, error) {
	if _, err := parse.expect(token.KwFn); err != nil {
		return nil, err
	}

	signature, err := parse.signature(parse.typeExpr)()

	if err != nil {
		return nil, err
	}

	return &ast.Function{Signature: signature.(*ast.Signature)}, nil
}

func (parse *parser) function() (ast.Node, error) {
	fnType, err := parse.functionType()

	if err != nil {
		return nil, err
	}

	tok, err := parse.expect(token.Eq)

	if err != nil {
		return nil, err
	}

	fn := fnType.(*ast.Function)
	fn.EqTok = tok.Span.From
	fn.Body, err = parse.expr()

	if err != nil {
		return nil, err
	}

	return fn, nil
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

	if body, err = parse.blockFunc(parse.case_); err != nil {
		return nil, err
	}

	return &ast.When{
		Expr:    expr,
		Body:    body,
		WhenTok: whenTok.Span.From,
	}, nil
}

func (parse *parser) case_() (ast.Node, error) {
	var (
		arrowTok token.Token
		pattern  ast.Node
		expr     ast.Node
		err      error
	)

	if pattern, err = parse.casePattern(); err != nil {
		return nil, err
	}

	if arrowTok, err = parse.expect(token.Arrow); err != nil {
		return nil, err
	}

	if expr, err = parse.expr(); err != nil {
		return nil, err
	}

	return &ast.Case{
		Pattern:  pattern,
		Expr:     expr,
		ArrowTok: arrowTok.Span.From,
	}, nil
}

func (parse *parser) casePattern() (ast.Node, error) {
	var node ast.Node
	var err error

	switch {
	case parse.matchAny(token.Int, token.Float, token.String):
		node, _ = parse.literal()

	case parse.matchAny(token.LowercaseIdent, token.IdentPlaceholder):
		name, _ := parse.ident()

		if parse.isTypeStart() {
			ty, err := parse.typeExpr()

			if err != nil {
				return nil, err
			}

			node = &ast.Decl{Ident: name, Type: ty}
		} else {
			node = name
		}

	case parse.match(token.UppercaseIdent):
		node, _ = parse.upperNode()

		if parse.match(token.LParen) {
			dot2OrLabeledExpr := parse.dot2(parse.labeled(parse.casePattern))

			if list, err := parse.parens(dot2OrLabeledExpr); err != nil {
				return nil, err
			} else {
				node = &ast.Call{X: node, Args: list}
			}
		}

	case parse.match(token.LBracket):
		node, err = parse.brackets(parse.dot2(parse.casePattern))

	case parse.match(token.LParen):
		node, err = parse.parens(parse.casePattern)

	default:
		return nil, errExpectedPattern(parse.Span, 0)
	}

	if err != nil {
		return nil, err
	}

	if tok, ok := parse.consume(token.KwAs); ok {
		name, err := parse.lowerNode()

		if err != nil {
			return nil, err
		}

		return &ast.As{
			Lhs:   node,
			Rhs:   name,
			AsTok: tok.Span.From,
		}, nil
	}

	return node, nil
}

func (parse *parser) dot2(fallback parseFunc) parseFunc {
	return func() (ast.Node, error) {
		if tok, ok := parse.consume(token.Dot2); ok {
			var ident ast.Ident

			if parse.matchAny(token.LowercaseIdent, token.IdentPlaceholder) {
				var err error

				if ident, err = parse.ident(); err != nil {
					return nil, err
				}
			}

			return &ast.Spread{
				Expr:      ident,
				SpreadTok: tok.Span.From,
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

	for oprKind, ok := operators[parse.Kind]; ok &&
		precedences[parse.Kind] >= precedence; {

		oprTok := parse.next()
		y, err := parse.binaryExpr(precedences[oprTok.Kind] + 1)

		if err != nil {
			return nil, err
		}

		x = &ast.Op{
			X:    x,
			Y:    y,
			Span: oprTok.Span,
			Kind: oprKind,
		}
	}

	return x, nil
}

func (parse *parser) prefix() (ast.Node, error) {
	if tok, ok := parse.consume(token.Minus); ok {
		y, err := parse.prefix()

		if err != nil {
			return nil, err
		}

		return &ast.Op{
			Y:    y,
			Span: tok.Span,
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

		switch parse.Kind {
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
	switch parse.Kind {
	case token.LowercaseIdent:
		return parse.lowerNode()

	case token.UppercaseIdent:
		return parse.upperNode()

	case token.Int, token.Float, token.String:
		return parse.literal()

	case token.LCurly:
		return parse.block()

	case token.LBracket:
		return parse.brackets(parse.expr)

	default:
		return nil, errExpectedOperand(parse.Span)
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

	switch parse.Kind {
	case token.LowercaseIdent:
		selector, _ = parse.lowerNode()

	case token.UppercaseIdent:
		selector, _ = parse.upperNode()

	case token.LCurly:
		if selector, err = parse.block(); err != nil {
			return nil, err
		}

	case token.LParen:
		if selector, err = parse.parens(parse.expr); err != nil {
			return nil, err
		}

	case token.LBracket:
		if selector, err = parse.brackets(parse.expr); err != nil {
			return nil, err
		}

	default:
		return nil, errUnexpectedToken(
			parse.Span,
			token.LowercaseIdent,
			token.UppercaseIdent,
			token.LCurly,
			token.LParen,
			token.LBracket,
		)
	}

	return &ast.Dot{
		X:      x,
		Y:      selector,
		DotPos: dotTok.Span.From,
	}, nil
}

//
//
//

/*
type comparableNode interface {
	comparable
	ast.Node
}

type comparableIdent interface {
	comparable
	ast.Ident
}

func node[T comparableNode](f func() (T, error)) func() (ast.Node, error) {
	return func() (ast.Node, error) {
		var zero T
		node, err := f()

		if err != nil {
			return nil, err
		}

		if node == zero {
			return nil, nil
		}

		return node, nil
	}
}

func ident[T comparableIdent](f func() (T, error)) func() (ast.Ident, error) {
	return func() (ast.Ident, error) {
		var zero T
		node, err := f()

		if err != nil {
			return nil, err
		}

		if node == zero {
			return nil, nil
		}

		return node, nil
	}
}
*/

var precedences = map[token.Kind]int{
	token.Asterisk: 10,
	token.Slash:    10,
	token.Percent:  10,
	token.Plus:     9,
	token.Minus:    9,
	token.Shl:      8,
	token.Shr:      8,
	token.Amp:      7,
	token.Bar:      7,
	token.Caret:    7,
	token.EqOp:     6,
	token.NeOp:     6,
	token.LtOp:     6,
	token.GtOp:     6,
	token.LeOp:     6,
	token.GeOp:     6,
	// token.And:        5,
	// token.Or:         4,
	token.KwAs:       3,
	token.Dot2:       2,
	token.Eq:         1,
	token.PlusEq:     1,
	token.MinusEq:    1,
	token.AsteriskEq: 1,
	token.SlashEq:    1,
	token.PercentEq:  1,
	token.AmpEq:      1,
	token.BarEq:      1,
	token.CaretEq:    1,
	token.ShlEq:      1,
	token.ShrEq:      1,
}

var operators = map[token.Kind]ast.OperatorKind{
	token.Eq:         ast.OperatorAssign,
	token.EqOp:       ast.OperatorEq,
	token.NeOp:       ast.OperatorNe,
	token.Bang:       ast.OperatorNot,
	token.Plus:       ast.OperatorAdd,
	token.PlusEq:     ast.OperatorAddAssign,
	token.Minus:      ast.OperatorSub,
	token.MinusEq:    ast.OperatorSubAssign,
	token.Asterisk:   ast.OperatorMul,
	token.AsteriskEq: ast.OperatorMulAssign,
	token.Slash:      ast.OperatorDiv,
	token.SlashEq:    ast.OperatorDivAssign,
	token.Percent:    ast.OperatorMod,
	token.PercentEq:  ast.OperatorModAssign,
	token.LtOp:       ast.OperatorLt,
	token.LeOp:       ast.OperatorLe,
	token.GtOp:       ast.OperatorGt,
	token.GeOp:       ast.OperatorGe,
	token.Amp:        ast.OperatorBitAnd,
	token.Bar:        ast.OperatorBitOr,
}

var literals = map[token.Kind]ast.LiteralKind{
	token.Int:    ast.IntLiteral,
	token.Float:  ast.FloatLiteral,
	token.String: ast.StringLiteral,
}
