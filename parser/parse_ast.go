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
		return parse.binding()

	case token.KwType:
		return parse.typeDecl()

	default:
		return parse.expr()
	}
}

func (parse *parser) fieldDecl() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	parse.labeled(
		func(label *ast.Lower, colon text.Pos) ast.Node {
			if label != nil {
				if colon.IsValid() {
					variable := parse.try(parse.typeExpr)

					return &ast.Label{
						Name:     label,
						X:        variable,
						ColonTok: colon,
					}
				} else {
					parse.error(errUnexpectedToken(
						label.Span,
						"lowercase identifier",
						"type",
					))
					panic("unreachable")
				}
			} else {
				if colon.IsValid() {
					parse.error(errUnexpectedToken(
						label.Span,
						"':' for label shorthand",
						"type",
						"labeled type",
					))
					panic("unreachable")
				} else {
					return parse.try(parse.typeExpr)
				}
			}
		},
	)

	return nil
}

func (parse *parser) variantDecl() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	name := parse.UppercaseIdent()
	params := (*ast.Parens)(nil)

	if parse.match(token.LParen) {
		params = parse.parens(
			parse.labeled(
				func(label *ast.Lower, colon text.Pos) ast.Node {
					if label != nil {
						if colon.IsValid() {
							variable := parse.try(parse.typeExpr)

							return &ast.Label{
								Name:     label,
								X:        variable,
								ColonTok: colon,
							}
						} else {
							parse.error(errUnexpectedToken(
								label.Span,
								"lowercase identifier",
								"type",
							))
							panic("unreachable")
						}
					} else {
						if colon.IsValid() {
							parse.error(errUnexpectedToken(
								label.Span,
								"':' for label shorthand",
								"type",
								"labeled type",
							))
							panic("unreachable")
						} else {
							return parse.try(parse.typeExpr)
						}
					}
				},
			),
		)
	}

	return &ast.VariantDecl{
		Name:   name,
		Params: params,
	}
}

func (parse *parser) typeVariable() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	name := parse.lowerNode().(*ast.Lower)

	return &ast.TypeVariable{
		Data: name.Data,
		Span: name.Span,
	}
}

func (parse *parser) binding() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok := parse.expectAny(token.KwLet, token.KwVal, token.KwVar)
	pattern := parse.pattern(patternKindValue).(ast.Pattern)

	parse.expect(token.Eq)
	parse.skipNewLines()

	value := parse.externalOr(parse.expr)()
	bindingKind := ast.ValueLet

	switch tok.Kind {
	case token.KwLet:
		// Default value.

	case token.KwVal:
		bindingKind = ast.ValueVal

	case token.KwVar:
		bindingKind = ast.ValueVar

	default:
		panic("unreachable")
	}

	return &ast.ValueDecl{
		Pattern:    pattern,
		Value:      value,
		KeywordPos: tok.Span.From,
		Kind:       bindingKind,
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

		return &ast.TypeAliasDecl{
			TypeTok: typeTok.Span.From,
			EqTok:   eqTok.Span.From,
			Ident:   name,
			Args:    args,
			Expr:    expr,
		}

	case token.LCurly:
		body := parse.blockFunc(parse.typeVariantOrField)

		return &ast.TypeDecl{
			TypeTok: typeTok.Span.From,
			Ident:   name,
			Args:    args,
			Body:    body,
		}

	default:
		parse.error(errUnexpectedToken(
			parse.Span,
			parse.Kind,
			"'=' for type alias",
			"'{' for type definition",
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
		return parse.fieldDecl()

	case token.UppercaseIdent:
		return parse.variantDecl()

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

	return &ast.Capitalized{
		Data: tok.Data,
		Span: tok.Span,
	}
}

func (parse *parser) UppercaseIdent() *ast.Capitalized {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok, ok := parse.consume(token.UppercaseIdent)

	if !ok {
		return nil
	}

	return &ast.Capitalized{
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

// func (parse *parser) typeArgs() *ast.Parens {
// 	return parse.parens(parse.typeExpr)
// }

func (parse *parser) isTypeStart() bool {
	return parse.matchAny(
		token.UppercaseIdent,
		token.LowercaseIdent,
		token.LParen,
	)
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

			// TODO replace with distinct node.
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
		return parse.signature(func() ast.Node { return fromLabel(parse, fieldFromLabel{}) })()

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

// func (parse *parser) variableFromLabel(
// 	label *ast.Lower,
// 	colon text.Pos,
// ) ast.Node {
// 	if parse.pushTrace() {
// 		defer parse.popTrace()
// 	}

// 	switch {
// 	case colon.IsValid():
// 		variable := parse.try(parse.variable)

// 		return &ast.Label{
// 			Name:     label,
// 			X:        variable,
// 			ColonTok: colon,
// 		}

// 	case label != nil:
// 		decl := &ast.Decl{Ident: label}

// 		if parse.isTypeStart() {
// 			decl.Type = parse.typeExpr()
// 		}

// 		return decl
// 	}

// 	return parse.variable()
// }

func (parse *parser) functionType() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	tok := parse.expect(token.KwFn)
	signature := parse.signature(parse.typeExpr)()

	return &ast.FnType{
		Signature: signature.(*ast.Signature),
		FnTok:     tok.Span.From,
	}
}

func (parse *parser) function() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	fnType := parse.functionType().(*ast.FnType)
	fnBody := parse.block().(*ast.Block)

	return &ast.Fn{
		Type: fnType,
		Body: fnBody,
	}
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
	body := parse.blockFunc(parse.caseClause)

	return &ast.When{
		Cond:    expr,
		Clauses: body,
		WhenTok: whenTok.Span.From,
	}
}

func (parse *parser) caseClause() ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	pattern := parse.pattern(patternKindCase)
	arrowTok := parse.expect(token.Arrow)
	expr := parse.expr()

	return &ast.CaseClause{
		Pattern:  pattern,
		Expr:     expr,
		ArrowTok: arrowTok.Span.From,
	}
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

	if parse.Kind != token.LowercaseIdent {
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

	return parse.lowerNode()
}

//
//
//

func (parse *parser) labeled(f func(label *ast.Lower, colon text.Pos) ast.Node) func() ast.Node {
	return func() ast.Node {
		if parse.pushTrace() {
			defer parse.popTrace()
		}

		// f   ->   nil, nil 	- no label, just 'f'
		// :f  ->   nil, colon 	- short label, 'f' must be able to parse an ident
		// l:f -> ident, colon 	- regular label, followed by 'f'
		// l   -> ident, nil 	- 'f' must be able to parse an ident
		ident := parse.LowercaseIdent()
		colon, _ := parse.consume(token.Colon)

		return f(ident, colon.Span.From)
	}
}

func (parse *parser) externalOr(f func() ast.Node) func() ast.Node {
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

func (parse *parser) signature(parseParamFunc func() ast.Node) func() ast.Node {
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

// func (parse *parser) spread(fallback func() ast.Node) func() ast.Node {
// 	return func() ast.Node {
// 		if parse.pushTrace() {
// 			defer parse.popTrace()
// 		}

// 		if tok, ok := parse.consume(token.Dot2); ok {
// 			ident := parse.identOrNil()

// 			return &ast.Spread{
// 				Expr:      ident,
// 				SpreadTok: tok.Span.From,
// 			}
// 		}

// 		return fallback()
// 	}
// }

func (parse *parser) blockFunc(f func() ast.Node) *ast.Block {
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

func (parse *parser) parens(f func() ast.Node) *ast.Parens {
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

func (parse *parser) brackets(f func() ast.Node) *ast.List {
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

type labelParser interface {
	FromLabel(parse *parser, name *ast.Lower, colon text.Pos) ast.Node
	FromLower(parse *parser, name *ast.Lower) ast.Node
	FromColon(parse *parser, colon text.Pos) ast.Node
	FromNothing(parse *parser) ast.Node
}

func fromLabel[T labelParser](parse *parser, t T) ast.Node {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	ident := parse.LowercaseIdent()
	colon, _ := parse.consume(token.Colon)

	// l:f -> ident, colon 	- regular label, followed by 'f'
	// l   -> ident, nil 	- 'f' must be able to parse an ident
	// :f  ->   nil, colon 	- short label, 'f' must be able to parse an ident
	// f   ->   nil, nil 	- no label, just 'f'
	if ident != nil {
		if colon.Span.IsValid() {
			return t.FromLabel(parse, ident, colon.Span.From)
		} else {
			return t.FromLower(parse, ident)
		}
	} else {
		if colon.Span.IsValid() {
			return t.FromColon(parse, colon.Span.From)
		} else {
			return t.FromNothing(parse)
		}
	}
}

type fieldFromLabel struct{}

func (fieldFromLabel) FromLabel(parse *parser, name *ast.Lower, colon text.Pos) ast.Node {
	ty := parse.try(parse.typeExpr)

	return &ast.FieldDecl{
		Label: nil,
		Name:  name,
		Type:  ty,
	}
}

func (fieldFromLabel) FromLower(parse *parser, name *ast.Lower) ast.Node {
	ty := parse.try(parse.typeExpr)

	return &ast.FieldDecl{
		Label: nil,
		Name:  name,
		Type:  ty,
	}
}

func (fieldFromLabel) FromColon(parse *parser, colon text.Pos) ast.Node {
	name := parse.try(func() ast.Node { return parse.lowerNode() }).(*ast.Lower)
	ty := parse.try(parse.typeExpr)

	return &ast.FieldDecl{
		Label: nil,
		Name:  name,
		Type:  ty,
	}
}

func (fieldFromLabel) FromNothing(parse *parser) ast.Node {
	if parse.match(token.UppercaseIdent) {
		return parse.variantDecl()
	}

	parse.error(errUnexpectedToken(
		parse.Span,
		parse.Kind,
		"'lowercase identifier' or ':' for field",
		"'capitalized identifier' for variant",
	))
	panic("unreachable")
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
