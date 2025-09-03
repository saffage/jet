package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

//------------------------------------------------
// Primitives
//------------------------------------------------

func (parse *parser) declOrExpr() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	parse.skipNewLines()

	switch parse.Kind {
	case token.KwLet, token.KwVal, token.KwVar:
		return parse.valueDecl()

	case token.KwType:
		return parse.typeDecl()

	default:
		return parse.expr()
	}
}

func (parse *parser) variable() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	name := parse.ident()
	ty := ast.Node(nil)

	if parse.isTypeStartOrSignature() {
		ty = parse.typeOrSignature()
	}

	return &ast.Decl{
		Ident:   name,
		Type:    ty,
		TypeTok: text.NoPos,
	}
}

func (parse *parser) variant() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	name := parse.UppercaseIdent()
	params := (*ast.Parens)(nil)

	if parse.match(token.LParen) {
		params = parse.parens(
			parse.labeled(
				func(label *ast.Lower, colon text.Pos) ast.Node {
					switch {
					case colon.IsValid():
						variable := parse.try(parse.variable)

						return &ast.Label{
							Name:     label,
							X:        variable,
							ColonTok: colon,
						}

					case label != nil:
						decl := &ast.Decl{Ident: label}

						if parse.isTypeStart() {
							decl.Type = parse.typeExpr()
						}

						return decl
					}

					return parse.typeExpr()
				},
			),
		)
	}

	return &ast.Variant{
		Name:   name,
		Params: params,
	}
}

func (parse *parser) typeVariable() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	name := parse.lowerNode()
	ty := ast.Node(nil)

	// type constraint
	// if parse.isTypeStart() {
	// 	ty = parse.typeExpr()
	// }

	decl := &ast.Decl{
		Ident:   name,
		Type:    ty,
		TypeTok: text.NoPos,
	}

	// if tok, ok := parse.consume(token.Eq); ok {
	// 	var tyDefault ast.Node
	//
	// 	if tyDefault, err = parse.typeExpr(); err != nil {
	// 		return nil, err
	// 	}
	//
	// 	return &ast.Op{
	// 		X:    decl,
	// 		Y:    tyDefault,
	// 		Span: tok.Span,
	// 		Kind: ast.OperatorAssign,
	// 	}, nil
	// }

	return decl
}

func (parse *parser) valueDecl() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok := parse.expectAny(token.KwLet, token.KwVal, token.KwVar)
	decl := parse.variable().(*ast.Decl)

	parse.expect(token.Eq)
	parse.skipNewLines()

	expr := parse.externalOr(parse.expr)()

	switch tok.Kind {
	case token.KwLet:
		return &ast.LetDecl{
			LetTok: tok.Span.From,
			Decl:   decl,
			Value:  expr,
		}

	case token.KwVal:
		return &ast.ValDecl{
			ValTok: tok.Span.From,
			Decl:   decl,
			Value:  expr,
		}

	case token.KwVar:
		return &ast.VarDecl{
			VarTok: tok.Span.From,
			Decl:   decl,
			Value:  expr,
		}

	default:
		panic("unreachable")
	}
}

func (parse *parser) typeDecl() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	typeTok := parse.expect(token.KwType)
	name := parse.UppercaseIdent()
	args := (*ast.Parens)(nil)

	if parse.match(token.LParen) {
		args = parse.parens(parse.typeVariable)
	}

	switch parse.Kind {
	case token.Eq:
		eqTok := parse.next()
		expr := parse.externalOr(parse.typeExpr)()

		return &ast.TypeAlias{
			TypeTok: typeTok.Span.From,
			EqTok:   eqTok.Span.From,
			Ident:   name,
			Args:    args,
			Expr:    expr,
		}

	case token.LCurly:
		body := parse.blockFunc(parse.typeVariantOrField)

		return &ast.TypeDef{
			TypeTok: typeTok.Span.From,
			Ident:   name,
			Args:    args,
			Body:    body,
		}

	default:
		parse.error(errUnexpectedToken(
			parse.Span,
			parse.Kind,
			token.Eq.String(),
			"type definition block",
		))
		panic("unreachable")
	}
}

func (parse *parser) typeVariantOrField() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	parse.skipNewLines()

	switch parse.Kind {
	case token.LowercaseIdent, token.IdentPlaceholder, token.Colon:
		return parse.labeled(parse.variableFromLabel)()

	case token.UppercaseIdent:
		return parse.variant()

	default:
		parse.error(errUnexpectedToken(
			parse.Span,
			parse.Kind,
			"field",
			"variant",
		))
		panic("unreachable")
	}
}

func (parse *parser) IdentPlaceholder() ast.Ident {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok := parse.expect(token.IdentPlaceholder)

	return &ast.Placeholder{
		Data: tok.Data,
		Span: tok.Span,
	}
}

func (parse *parser) lowerNode() ast.Ident {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok := parse.expect(token.LowercaseIdent)

	return &ast.Lower{
		Data: tok.Data,
		Span: tok.Span,
	}
}

func (parse *parser) LowercaseIdent() *ast.Lower {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok, ok := parse.consume(token.LowercaseIdent)

	if !ok {
		return nil
	}

	return &ast.Lower{Data: tok.Data, Span: tok.Span}
}

func (parse *parser) upperNode() ast.Ident {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok := parse.expect(token.UppercaseIdent)

	return &ast.Upper{
		Data: tok.Data,
		Span: tok.Span,
	}
}

func (parse *parser) UppercaseIdent() *ast.Upper {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok, ok := parse.consume(token.UppercaseIdent)

	if !ok {
		return nil
	}

	return &ast.Upper{
		Data: tok.Data,
		Span: tok.Span,
	}
}

func (parse *parser) literal() ast.Node {
	tok, ok := parse.consumeAny(token.Int, token.Float, token.String)

	if !ok {
		parse.error(errExpectedOperand(parse.Span))
		panic("unreachable")
	}

	return &ast.Literal{
		Value: tok.Data,
		Span:  tok.Span,
		Kind:  literals[tok.Kind],
	}
}

func (parse *parser) ident() ast.Ident {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	switch parse.Kind {
	case token.LowercaseIdent:
		return parse.lowerNode()

	case token.IdentPlaceholder:
		return parse.IdentPlaceholder()

	default:
		parse.error(errUnexpectedToken(
			parse.Span,
			parse.Kind,
			token.LowercaseIdent,
			token.IdentPlaceholder,
		))
		panic("unreachable")
	}
}

func (parse *parser) identOrNil() ast.Ident {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	switch parse.Kind {
	case token.LowercaseIdent:
		return parse.lowerNode()

	case token.IdentPlaceholder:
		return parse.IdentPlaceholder()

	default:
		return nil
	}
}

func (parse *parser) block() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	return parse.blockFunc(parse.declOrExpr)
}

func (parse *parser) args() *ast.Parens {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	return parse.parens(
		parse.labeled(
			func(label *ast.Lower, colon text.Pos) ast.Node {
				switch {
				case colon.IsValid():
					expr := ast.Node(nil)

					if label == nil {
						expr = parse.try(func() ast.Node {
							// TODO: this must be a valid short label expression.
							return parse.ident()
						})
					} else {
						expr = parse.try(parse.expr)
					}

					return &ast.Label{
						Name:     label,
						X:        expr,
						ColonTok: colon,
					}

				case label != nil:
					return parse.binaryExprFrom(parse.primaryFrom(label), 1)
				}

				return parse.expr()
			},
		),
	)
}

func (parse *parser) typeArgs() ast.Node {
	return parse.parens(parse.typeExpr)
}

func (parse *parser) isTypeStart() bool {
	return parse.matchAny(token.UppercaseIdent, token.LowercaseIdent, token.LParen)
}

func (parse *parser) isTypeStartOrSignature() bool {
	return parse.matchAny(
		token.UppercaseIdent,
		token.LowercaseIdent,
		token.KwFn,
		token.LParen,
	)
}

// `A | B | C | ...`
func (parse *parser) typeExpr() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	expr := parse.simpleTypeExpr()

	for {
		pipeTok, ok := parse.consume(token.Bar)

		if !ok {
			break
		}

		node := parse.try(parse.simpleTypeExpr)

		expr = &ast.Op{
			X:    expr,
			Y:    node,
			Kind: ast.OperatorBitOr,
			Span: pipeTok.Span,
		}
	}

	return expr
}

func (parse *parser) simpleTypeExpr() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	switch parse.Kind {
	case token.UppercaseIdent:
		node := parse.upperNode()

		if parse.match(token.LParen) {
			typeArgs := parse.parens(parse.typeExpr)

			return &ast.Call{
				X:    node,
				Args: typeArgs,
			}
		}

		return node

	case token.LowercaseIdent:
		// TODO: allow type arguments?
		tok := parse.next()

		return &ast.Lower{
			Data: tok.Data,
			Span: tok.Span,
		}

	case token.KwFn:
		return parse.functionType()

	default:
		parse.error(errUnexpectedToken(
			parse.Span,
			parse.Kind,
			token.UppercaseIdent,
			token.LowercaseIdent,
			token.KwFn,
		))
		panic("unreachable")
	}
}

func (parse *parser) typeOrSignature() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	switch {
	case parse.match(token.LParen):
		return parse.signature(
			parse.labeled(parse.variableFromLabel),
		)()

	case parse.isTypeStart():
		return parse.typeExpr()

	default:
		parse.error(errUnexpectedToken(
			parse.Span,
			parse.Kind.String(),
			token.UppercaseIdent.String(),
			token.LowercaseIdent.String(),
			token.KwFn.String(),
			"'(' for function signature",
		))
		panic("unreachable")
	}
}

func (parse *parser) variableFromLabel(
	label *ast.Lower,
	colon text.Pos,
) ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	switch {
	case colon.IsValid():
		variable := parse.try(parse.variable)

		return &ast.Label{
			Name:     label,
			X:        variable,
			ColonTok: colon,
		}

	case label != nil:
		decl := &ast.Decl{Ident: label}

		if parse.isTypeStart() {
			decl.Type = parse.typeExpr()
		}

		return decl
	}

	return parse.variable()
}

func (parse *parser) functionType() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	parse.expect(token.KwFn)

	signature := parse.signature(parse.typeExpr)()

	return &ast.Function{
		Signature: signature.(*ast.Signature),
		Body:      nil,
		FnTok:     text.NoPos,
		EqTok:     text.NoPos,
	}
}

func (parse *parser) function() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	fn := parse.functionType().(*ast.Function)
	tok := parse.expect(token.Eq)
	fn.EqTok = tok.Span.From
	fn.Body = parse.expr()

	return fn
}

func (parse *parser) expr() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	switch {
	case parse.match(token.KwWhen):
		return parse.whenExpr()

	// case parse.match(token.KwFn):
	// 	return parse.function()

	default:
		return parse.binaryExpr(1)
	}
}

func (parse *parser) whenExpr() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	whenTok := parse.expect(token.KwWhen)
	expr := parse.expr()
	body := parse.blockFunc(parse.case_)

	return &ast.When{
		Expr:    expr,
		Body:    body,
		WhenTok: whenTok.Span.From,
	}
}

func (parse *parser) case_() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	pattern := parse.casePattern()
	arrowTok := parse.expect(token.Arrow)
	expr := parse.expr()

	return &ast.Case{
		Pattern:  pattern,
		Expr:     expr,
		ArrowTok: arrowTok.Span.From,
	}
}

func (parse *parser) casePattern() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	node := ast.Node(nil)

	switch {
	case parse.matchAny(token.Int, token.Float, token.String):
		node = parse.literal()

	case parse.matchAny(token.LowercaseIdent, token.IdentPlaceholder):
		name := parse.ident()

		if parse.isTypeStart() {
			ty := parse.typeExpr()

			node = &ast.Decl{Ident: name, Type: ty}
		} else {
			node = name
		}

	case parse.match(token.UppercaseIdent):
		node = parse.upperNode()

		if parse.match(token.LParen) {
			spreadOrLabeledExpr := parse.spread(
				parse.labeled(
					func(label *ast.Lower, colon text.Pos) ast.Node {
						switch {
						case label != nil && colon.IsValid():
							x := parse.try(parse.casePattern)

							return &ast.Label{
								Name:     label,
								X:        x,
								ColonTok: colon,
							}

						case label != nil && !colon.IsValid():
							if parse.isTypeStart() {
								ty := parse.typeExpr()

								node = &ast.Decl{Ident: label, Type: ty}
							} else {
								node = label
							}

						case label == nil && colon.IsValid():
							if ident := parse.LowercaseIdent(); ident != nil {
								return &ast.Label{
									X:        ident,
									ColonTok: colon,
								}
							}
						}
						return parse.casePattern()
					},
				),
			)

			list := parse.parens(spreadOrLabeledExpr)
			node = &ast.Call{X: node, Args: list}
		}

	case parse.match(token.LBracket):
		node = parse.brackets(parse.spread(parse.casePattern))

	case parse.match(token.LParen):
		node = parse.parens(parse.casePattern)

	default:
		parse.error(errExpectedPattern(parse.Span, 0))
		panic("unreachable")
	}

	if tok, ok := parse.consume(token.KwAs); ok {
		name := parse.lowerNode()

		return &ast.As{
			Expr:    node,
			NewName: name,
			AsTok:   tok.Span.From,
		}
	}

	return node
}

func (parse *parser) binaryExpr(precedence int) ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	return parse.binaryExprFrom(parse.prefix(), precedence)
}

func (parse *parser) binaryExprFrom(x ast.Node, precedence int) ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	if x == nil {
		panic("unreachable")
	}

	for oprKind, ok := operators[parse.Kind]; ok &&
		precedences[parse.Kind] >= precedence; {

		operatorTok := parse.next()
		y := parse.binaryExpr(precedences[operatorTok.Kind] + 1)

		x = &ast.Op{
			X:    x,
			Y:    y,
			Span: operatorTok.Span,
			Kind: oprKind,
		}
	}

	return x
}

func (parse *parser) prefix() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	if tok, ok := parse.consume(token.Minus); ok {
		y := parse.prefix()

		return &ast.Op{
			Y:    y,
			Span: tok.Span,
			Kind: ast.OperatorNeg,
		}
	}

	return parse.primary()
}

func (parse *parser) primary() ast.Node {
	return parse.primaryFrom(parse.operand())
}

func (parse *parser) primaryFrom(operand ast.Node) ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	if operand == nil {
		panic("unreachable")
	}

	for {
		switch parse.Kind {
		case token.Dot:
			dotTok := parse.next()

			parse.skipNewLines()

			operand = &ast.Dot{
				X:      operand,
				Y:      parse.try(parse.selector),
				DotPos: dotTok.Span.From,
			}

		case token.LParen:
			operand = &ast.Call{
				X:    operand,
				Args: parse.args(),
			}

		default:
			return operand
		}
	}
}

func (parse *parser) operand() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	switch parse.Kind {
	case token.LowercaseIdent, token.IdentPlaceholder:
		return parse.ident()

	case token.UppercaseIdent:
		return parse.upperNode()

	case token.Int, token.Float, token.String:
		return parse.literal()

	case token.LCurly:
		return parse.block()

	case token.LBracket:
		return parse.brackets(parse.expr)

	default:
		parse.error(errExpectedOperand(parse.Span))
		panic("unreachable")
	}
}

//
//
//

func (parse *parser) selector() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	switch parse.Kind {
	case token.LowercaseIdent:
		return parse.lowerNode()

	case token.UppercaseIdent:
		return parse.upperNode()

	case token.LCurly:
		return parse.block()

	case token.LParen:
		return parse.parens(parse.expr)

	case token.LBracket:
		return parse.brackets(parse.expr)

	default:
		parse.error(errUnexpectedToken(
			parse.Span,
			parse.Kind,
			token.LowercaseIdent,
			token.UppercaseIdent,
			token.LCurly,
			token.LParen,
			token.LBracket,
		))
		panic("unreachable")
	}
}

//
//
//

func (parse *parser) labeled(f parseLabeledFunc) parseFunc {
	return func() ast.Node {
		if parse.pushTrace() {
			defer parse.popTrace()
		}

		// f   ->   nil, nil 	- no ident
		// :f  ->   nil, colon 	- short ident, 'f' must be able to parse an ident
		// l:f -> ident, colon 	- regular ident, followed by 'f'
		// l   -> ident, nil 	- 'f' must be able to parse an ident
		ident := parse.LowercaseIdent()
		colon, _ := parse.consume(token.Colon)

		return f(ident, colon.Span.From)
	}
}

func (parse *parser) externalOr(f parseFunc) parseFunc {
	return func() ast.Node {
		if parse.pushTrace() {
			defer parse.popTrace()
		}

		if tok, ok := parse.consume(token.KwExternal); ok {
			args := (*ast.Parens)(nil)

			if parse.match(token.LParen) {
				args = parse.args()
			}

			return &ast.External{
				ExternalTok: tok.Span.From,
				Args:        args,
			}
		}

		return f()
	}
}

func (parse *parser) signature(parseParamFunc parseFunc) parseFunc {
	return func() ast.Node {
		if parse.pushTrace() {
			defer parse.popTrace()
		}

		params := parse.parens(parseParamFunc)
		result := ast.Node(nil)

		if parse.isTypeStart() {
			result = parse.typeExpr()
		}

		return &ast.Signature{
			Params: params,
			Result: result,
		}
	}
}

func (parse *parser) spread(fallback parseFunc) parseFunc {
	return func() ast.Node {
		if parse.pushTrace() {
			defer parse.popTrace()
		}

		if tok, ok := parse.consume(token.Dot2); ok {
			ident := parse.identOrNil()

			return &ast.Spread{
				Expr:      ident,
				SpreadTok: tok.Span.From,
			}
		}

		return fallback()
	}
}

func (parse *parser) blockFunc(f parseFunc) *ast.Block {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	if !parse.match(token.LCurly) {
		parse.error(errExpectedBlock(parse.Span))
		panic("unreachable")
	}

	nodes, span := parse.listOpenClose(
		f,
		token.LCurly,
		token.RCurly,
		token.Semicolon,
		token.Newline,
	)

	return &ast.Block{
		Stmts: &ast.Stmts{Items: nodes},
		Span:  span,
	}
}

func (parse *parser) parens(f parseFunc) *ast.Parens {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	nodes, span := parse.listOpenClose(
		f,
		token.LParen,
		token.RParen,
		token.Comma,
	)

	return &ast.Parens{
		Nodes: nodes,
		Span:  span,
	}
}

func (parse *parser) brackets(f parseFunc) *ast.List {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	nodes, span := parse.listOpenClose(
		f,
		token.LBracket,
		token.RBracket,
		token.Comma,
	)

	return &ast.List{
		Nodes: nodes,
		Span:  span,
	}
}

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
