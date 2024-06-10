package parser

import (
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

type ParseFunc func() ast.Node

//------------------------------------------------
// Primitives
//------------------------------------------------

func (p *Parser) parseIdent() ast.Node {
	if node := p.parseIdentNode(); node != nil {
		return node
	}

	p.errorExpectedToken(token.Ident)
	return nil
}

func (p *Parser) parseLiteral() ast.Node {
	if node := p.parseLiteralNode(); node != nil {
		return node
	}

	p.errorExpectedToken(token.Int, token.Float, token.String)
	return nil
}

//------------------------------------------------
// Statements
//------------------------------------------------

func (p *Parser) declOr(f ParseFunc) ParseFunc {
	first := true
	isDecl := false

	return func() ast.Node {
		// Infer what we need to parse.
		if first {
			first = false
			begin := p.save()

			// Try declaration first.
			if decl := p.parseDecl(); decl != nil {
				isDecl = true
				return decl
			}

			// Check is error was occured while parsing a declaration.
			// If not, try parse an expression.
			if p.lastErrorIs(ErrorExpectedDecl) {
				p.restore(begin)
				return f()
			}
		}

		if !isDecl {
			return f()
		} else if decl := p.parseDecl(); decl != nil {
			return decl
		}

		return nil
	}
}

func (p *Parser) parseStmt() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	for p.tok.Kind == token.NewLine {
		p.next()
	}

	if p.consume(token.Semicolon) != nil {
		return &ast.Empty{DesiredPos: p.tok.Start}
	}

	return p.declOr(p.parseExpr)()
}

//------------------------------------------------
// Expressions
//------------------------------------------------

func (p *Parser) parseExpr() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	switch p.tok.Kind {
	case token.KwIf:
		return p.parseIf()

	case token.KwWhile:
		return p.parseWhile()

	case token.KwFor:
		return p.parseFor()

	case token.KwReturn:
		return p.parseReturn()

	case token.KwBreak, token.KwContinue:
		return p.parseBreakOrContinue()

	default:
		return p.parseSimpleExpr()
	}
}

func (p *Parser) parseSimpleExpr() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	return p.parseBinaryExpr(nil, 1)
}

func (p *Parser) parseBinaryExpr(lhs ast.Node, precedence token.Precedence) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if lhs == nil {
		lhs = p.parseUnaryExpr()
	}

	if lhs == nil {
		return nil
	}

	for p.tok.Precedence() >= precedence {
		tok := p.consume()

		for p.tok.Kind == token.NewLine {
			p.next()
		}

		rhs := p.parseBinaryExpr(nil, tok.Precedence()+1)

		if rhs == nil {
			return nil
		}

		binaryOpKind := ast.UnknownOperator

		switch tok.Kind {
		case token.Plus:
			binaryOpKind = ast.OperatorAdd

		case token.Minus:
			binaryOpKind = ast.OperatorSub

		case token.Asterisk:
			binaryOpKind = ast.OperatorMul

		case token.Slash:
			binaryOpKind = ast.OperatorDiv

		case token.Percent:
			binaryOpKind = ast.OperatorMod

		case token.Eq:
			binaryOpKind = ast.OperatorAssign

		case token.PlusEq:
			binaryOpKind = ast.OperatorAddAndAssign

		case token.MinusEq:
			binaryOpKind = ast.OperatorSubAndAssign

		case token.AsteriskEq:
			binaryOpKind = ast.OperatorMultAndAssign

		case token.SlashEq:
			binaryOpKind = ast.OperatorDivAndAssign

		case token.PercentEq:
			binaryOpKind = ast.OperatorModAndAssign

		case token.EqOp:
			binaryOpKind = ast.OperatorEq

		case token.NeOp:
			binaryOpKind = ast.OperatorNe

		case token.LtOp:
			binaryOpKind = ast.OperatorLt

		case token.LeOp:
			binaryOpKind = ast.OperatorLe

		case token.GtOp:
			binaryOpKind = ast.OperatorGt

		case token.GeOp:
			binaryOpKind = ast.OperatorGe

		case token.Amp:
			binaryOpKind = ast.OperatorBitAnd

		case token.Pipe:
			binaryOpKind = ast.OperatorBitOr

		case token.Caret:
			binaryOpKind = ast.OperatorBitXor

		case token.Shl:
			binaryOpKind = ast.OperatorBitShl

		case token.Shr:
			binaryOpKind = ast.OperatorBitShr

		case token.KwAnd:
			binaryOpKind = ast.OperatorAnd

		case token.KwOr:
			binaryOpKind = ast.OperatorOr

		case token.Dot2:
			binaryOpKind = ast.OperatorRangeInclusive

		case token.Dot2Less:
			binaryOpKind = ast.OperatorRangeExclusive

		default:
			p.errorfAt(
				ErrorInvalidBinaryOperator,
				tok.Start,
				tok.End,
				"%s cannot be used in the binary expression",
				tok.Kind.UserString(),
			)
		}

		lhs = &ast.Op{
			X:     lhs,
			Y:     rhs,
			Start: tok.Start,
			End:   tok.End,
			Kind:  binaryOpKind,
		}
	}

	return lhs
}

func (p *Parser) parseUnaryExpr() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	switch p.tok.Kind {
	case token.Minus:
		minus := p.consume()

		return &ast.Op{
			X:     nil,
			Y:     p.parseUnaryExpr(),
			Start: minus.Start,
			End:   minus.End,
			Kind:  ast.OperatorNeg,
		}

	case token.Bang:
		bang := p.consume()

		return &ast.Op{
			X:     nil,
			Y:     p.parseUnaryExpr(),
			Start: bang.Start,
			End:   bang.End,
			Kind:  ast.OperatorNot,
		}

	case token.Asterisk:
		asterisk := p.consume()

		return &ast.Op{
			X:     nil,
			Y:     p.parseUnaryExpr(),
			Start: asterisk.Start,
			End:   asterisk.End,
			Kind:  ast.OperatorStar,
		}

	case token.Amp:
		amp := p.consume()

		// if varTok := p.consume(token.KwVar); varTok != nil {
		// 	return &ast.PrefixOp{
		// 		X: p.parseUnaryExpr(),
		// 		Opr: &ast.Operator{
		// 			Start: loc,
		// 			End:   varTok.End,
		// 			Kind:  ast.OperatorMutAddr,
		// 		},
		// 	}
		// }

		return &ast.Op{
			X:     nil,
			Y:     p.parseUnaryExpr(),
			Start: amp.Start,
			End:   amp.End,
			Kind:  ast.OperatorAddrOf,
		}

	default:
		return p.parsePrimaryExpr(nil)
	}
}

func (p *Parser) parsePrimaryExpr(x ast.Node) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if x == nil {
		x = p.parseOperand()
	}

	if x == nil {
		return nil
	}

	return p.parseSuffixExpr(x)
}

func (p *Parser) parseSuffixExpr(x ast.Node) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if x == nil {
		x = p.parseOperand()
	}

	for {
		if x == nil {
			return nil
		}

		switch p.tok.Kind {
		case token.Dot:
			x = p.parseDot(x)

		// case token.Amp, token.Asterisk:
		// 	before := p.save()
		// 	tok := p.consume()

		// 	if !p.isExprStart(p.tok.Kind) {
		// 		oprKind := ast.UnknownOperator

		// 		switch tok.Kind {
		// 		case token.Asterisk:
		// 			oprKind = ast.OperatorDeref

		// 		case token.Amp:
		// 			oprKind = ast.OperatorAddrOf
		// 		}

		// 		x = &ast.PostfixOp{
		// 			X: x,
		// 			Opr: &ast.Operator{
		// 				Start: tok.Start,
		// 				End:   tok.End,
		// 				Kind:  oprKind,
		// 			},
		// 		}
		// 	} else {
		// 		p.restore(before)
		// 		return x
		// 	}

		case token.LBracket:
			x = &ast.Index{
				X:    x,
				Args: p.parseBracketList(p.parseExpr),
			}

		case token.LParen:
			x = &ast.Call{
				X:    x,
				Args: p.parseParenList(p.declOr(p.parseExpr)),
			}

		case token.KwAs:
			asTok := p.consume()

			x = &ast.Op{
				X:     x,
				Y:     p.parseType(),
				Start: asTok.Start,
				End:   asTok.End,
				Kind:  ast.OperatorAs,
			}

		default:
			return x
		}
	}
}

func (p *Parser) parseOperand() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	switch p.tok.Kind {
	case token.Ident:
		return p.parseIdent()

	case token.Dollar:
		return p.parseBuiltIn()

	case token.Int, token.Float, token.String:
		return p.parseLiteral()

	case token.LParen:
		operand := p.parseParenList(p.declOr(p.parseExpr))
		if p.match(token.Arrow) || !p.match(endOfExprKinds...) {
			return p.parseFunction(operand)
		}
		return operand

	case token.LCurly:
		return p.parseBlock()

	case token.LBracket:
		if list := p.parseBracketList(p.parseExpr); list != nil {
			return list
		}
		return nil

	case token.KwStruct:
		// NOTE needed for expressions like `struct{}()`
		return p.parseStructType()

	case token.KwEnum:
		// NOTE needed for expressions like `enum{}.Foo`
		return p.parseEnumType()

	default:
		p.error(ErrorExpectedOperand)
		return nil
	}
}

//------------------------------------------------
// Complex expressions
//------------------------------------------------

func (p *Parser) parseFunction(params *ast.ParenList) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	// NOTE can be a type

	var result, body ast.Node

	if p.consume(token.Arrow) != nil {
		// (...) -> T {}
		result = p.parseType()
		if result == nil {
			return nil
		}

		if !slices.Contains(endOfExprKinds, p.tok.Kind) {
			if block := p.parseBlock(); block != nil {
				body = block
			}
		}
	} else {
		// (...) expr
		if !slices.Contains(endOfExprKinds, p.tok.Kind) {
			if expr := p.parseExpr(); expr != nil {
				body = expr
			}
		}
	}

	return &ast.Function{
		Signature: &ast.Signature{
			Params: params,
			Result: result,
		},
		Body: body,
	}
}

func (p *Parser) parseDot(x ast.Node) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if x == nil {
		panic("can't use nil node as left-hand side expression")
	}

	dot := p.consume(token.Dot)
	if dot == nil {
		return x
	}

	if star := p.consume(token.Asterisk); star != nil {
		return &ast.Deref{
			X:       x,
			DotPos:  dot.Start,
			StarPos: star.Start,
		}
	}

	y := p.parseIdentNode()
	if y == nil {
		p.errorExpectedToken(token.Ident)
		return nil
	}

	return &ast.Dot{
		X:      x,
		Y:      y,
		DotPos: dot.Start,
	}
}

func (p *Parser) parseBuiltIn() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.Dollar)
	if tok == nil {
		return nil
	}

	ident := p.parseIdentNode()
	if ident == nil {
		p.errorExpectedToken(token.Ident)
		return nil
	}

	return &ast.BuiltIn{
		Ident:  ident,
		TokPos: tok.Start,
	}
}

//------------------------------------------------
// Declarations
//------------------------------------------------

func (p *Parser) parseDecl() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	attributes := p.parseAttributeListNode()

	if attributes != nil {
		for p.tok.Kind == token.NewLine {
			p.next()
		}
	}

	mutLoc := token.Pos{}
	if tokMut := p.consume(token.KwMut); tokMut != nil {
		mutLoc = tokMut.Start
	}

	if p.matchSequence(token.Ident, token.Colon) {
		if decl := p.parseDeclNode(mutLoc, p.parseIdentNode()); decl != nil {
			if attributes != nil {
				decl.Attrs = attributes
			}
			// decl.Docs = p.commentGroup
			// p.commentGroup = nil
			return decl
		}
	} else if mutLoc.IsValid() {
		p.error(ErrorExpectedIdentAfterMut)
	} else if attributes != nil {
		p.error(ErrorExpectedDeclAfterAttrs)
	} else {
		p.error(ErrorExpectedDecl)
	}

	return nil
}

func (p *Parser) parseDeclNode(mut token.Pos, name *ast.Ident) *ast.Decl {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if p.expect(token.Colon) == nil {
		return nil
	}

	var ty, value ast.Node
	isVar := true

	switch {
	case p.tok.Kind != token.Colon && p.tok.Kind != token.Eq:
		ty = p.parseType()
		if ty == nil {
			return nil
		}
		fallthrough

	default:
		if tok := p.consume(token.Colon, token.Eq); tok != nil {
			// TODO value can be a type.
			isVar = tok.Kind != token.Colon
			value = p.parseExpr()
			if value == nil {
				return nil
			}
		} else if ty == nil {
			p.error(ErrorExpectedTypeOrValue)
			return nil
		}
	}

	return &ast.Decl{
		Ident: name,
		Mut:   mut,
		Type:  ty,
		Value: value,
		IsVar: isVar,
	}
}

//------------------------------------------------
// IDK
//------------------------------------------------

func (p *Parser) parseAttributeListNode() *ast.AttributeList {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if p.matchSequence(token.At, token.LBracket) {
		tok := p.consume(token.At)

		if list := p.parseBracketList(p.parseExpr); list != nil {
			return &ast.AttributeList{
				TokLoc: tok.Start,
				List:   list,
			}
		}
	}

	return nil
}

func (p *Parser) parseType() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	switch p.tok.Kind {
	case token.Ellipsis:
		// TODO allow only 1 variadic

		elipsis := p.consume()
		op := &ast.Op{
			X:     nil,
			Y:     nil,
			Start: elipsis.Start,
			End:   elipsis.End,
			Kind:  ast.OperatorEllipsis,
		}

		typeStart := p.save()

		if x := p.parseType(); x != nil {
			op.X = x
			return op
		}

		if p.lastErrorIs(ErrorExpectedType) {
			p.restore(typeStart)
		}

		return op

	case token.Asterisk:
		star := p.consume()

		return &ast.Op{
			X:     nil,
			Y:     p.parseType(),
			Start: star.Start,
			End:   star.End,
			Kind:  ast.OperatorStar,
		}

	case token.LBracket:
		brackets := p.parseBracketList(p.parseExpr)

		return &ast.ArrayType{
			X:    p.parseType(),
			Args: brackets,
		}

	case token.LParen:
		params := p.parseParenList(p.declOr(p.parseType))
		if p.consume(token.Arrow) != nil {
			return &ast.Signature{
				Params: params,
				Result: p.parseType(),
			}
		}
		return params

	case token.Ident:
		return p.parseTypeName()

	case token.KwStruct:
		return p.parseStructType()

	case token.KwEnum:
		return p.parseEnumType()

	case token.Dollar:
		return p.parseBuiltIn()
	}

	p.error(ErrorExpectedType)
	return nil
}

func (p *Parser) parseStructType() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwStruct)
	if tok == nil {
		return nil
	}

	body := p.parseBlockFunc(p.parseDecl)
	if body == nil {
		return nil
	}

	fields := make([]*ast.Decl, len(body.Nodes))

	// Filter bad nodes.
	for i := range fields {
		fields[i], _ = body.Nodes[i].(*ast.Decl)
	}

	return &ast.StructType{
		Fields: fields,
		TokPos: tok.Start,
		Open:   body.Open,
		Close:  body.Close,
	}
}

func (p *Parser) parseEnumType() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwEnum)
	if tok == nil {
		return nil
	}

	body := p.parseBlockFunc(p.parseIdent)
	if body == nil {
		return nil
	}

	fields := make([]*ast.Ident, len(body.Nodes))

	// Filter bad nodes.
	for i := range fields {
		fields[i], _ = body.Nodes[i].(*ast.Ident)
	}

	return &ast.EnumType{
		Fields: fields,
		TokPos: tok.Start,
		Open:   body.Open,
		Close:  body.Close,
	}
}

func (p *Parser) parseElse() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if elseTok := p.consume(token.KwElse); elseTok != nil {
		body := ast.Node(nil)

		if p.tok.Kind == token.KwIf {
			body = p.parseIf()
		} else if block := p.parseBlock(); block != nil {
			body = block
		} else {
			start, end := p.skip()
			p.errorAt(ErrorExpectedBlockOrIf, start, end)
			return nil
		}

		return &ast.Else{
			TokPos: elseTok.Start,
			Body:   body,
		}
	}

	return (*ast.Else)(nil)
}

func (p *Parser) parseIf() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwIf)

	if tok == nil {
		return nil
	}

	cond := p.parseSimpleExpr()

	if cond == nil {
		start, end := p.skip()
		p.errorAt(ErrorExpectedExpr, start, end)
		return nil
	}

	body := p.parseBlock()
	if body == nil {
		return nil
	}

	elseNode := p.parseElse()
	elseClause, ok := elseNode.(*ast.Else)

	if !ok {
		return nil
	}

	return &ast.If{
		TokPos: tok.Start,
		Cond:   cond,
		Body:   body,
		Else:   elseClause,
	}
}

func (p *Parser) parseWhile() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwWhile)
	if tok == nil {
		return nil
	}

	cond := p.parseExpr()
	if cond == nil {
		return nil
	}

	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return &ast.While{
		TokPos: tok.Start,
		Cond:   cond,
		Body:   body,
	}
}

func (p *Parser) parseFor() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwFor)
	if tok == nil {
		return nil
	}

	declList := p.parseForLoopDeclList()
	if declList == nil {
		return nil
	}

	if p.consume(token.KwIn) == nil {
		return nil
	}

	iterExpr := p.parseExpr()
	if iterExpr == nil {
		return nil
	}

	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return &ast.For{
		DeclList: declList,
		IterExpr: iterExpr,
		Body:     body,
		TokPos:   tok.Start,
	}
}

func (p *Parser) parseForLoopDeclList() *ast.List {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	var decls []ast.Node

	for {
		decl := p.parseForLoopDecl()
		if decl == nil {
			return nil
		}
		decls = append(decls, decl)
		if p.consume(token.Comma) == nil {
			break
		}
	}

	return &ast.List{Nodes: decls}
}

func (p *Parser) parseForLoopDecl() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	attributes := p.parseAttributeListNode()

	if attributes != nil {
		for p.tok.Kind == token.NewLine {
			p.next()
		}
	}

	mutLoc := token.Pos{}
	if tokMut := p.consume(token.KwMut); tokMut != nil {
		mutLoc = tokMut.Start
	}

	name := p.parseIdentNode()
	if name == nil {
		if mutLoc.IsValid() {
			p.error(ErrorExpectedIdentAfterMut)
		} else if attributes != nil {
			p.error(ErrorExpectedDeclAfterAttrs)
		} else {
			p.error(ErrorExpectedIdent)
		}

		return nil
	}

	var ty ast.Node

	if p.consume(token.Colon) != nil {
		ty = p.parseType()
		if ty == nil {
			return nil
		}
	}

	return &ast.Decl{
		Attrs: attributes,
		Ident: name,
		Mut:   mutLoc,
		Type:  ty,
	}
}

func (p *Parser) parseReturn() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwReturn)
	x := p.parseExpr()

	return &ast.Return{
		X:      x,
		TokPos: tok.Start,
	}
}

func (p *Parser) parseBreakOrContinue() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	// TODO implement new label syntax `break@label`

	tok := p.expect(token.KwBreak, token.KwContinue)
	label := (*ast.Ident)(nil)

	if p.tok.Kind == token.Ident {
		label = p.parseIdentNode()
	}

	switch tok.Kind {
	case token.KwBreak:
		return &ast.Break{
			Label:  label,
			TokPos: tok.Start,
		}

	case token.KwContinue:
		return &ast.Continue{
			Label:  label,
			TokPos: tok.Start,
		}

	default:
		panic("unreachable")
	}
}

//------------------------------------------------
// Lists
//------------------------------------------------

func (p *Parser) parseBracketList(f ParseFunc) *ast.BracketList {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if exprs, openLoc, closeLoc, _ := p.parseBracketedList(
		f,
		token.LBracket,
		token.RBracket,
		token.Comma,
	); exprs != nil {
		return &ast.BracketList{
			List:  &ast.List{Nodes: exprs},
			Open:  openLoc,
			Close: closeLoc,
		}
	}

	return nil
}

func (p *Parser) parseParenList(f ParseFunc) *ast.ParenList {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if exprs, openLoc, closeLoc, _ := p.parseBracketedList(
		f,
		token.LParen,
		token.RParen,
		token.Comma,
	); exprs != nil {
		return &ast.ParenList{
			List:  &ast.List{Nodes: exprs},
			Open:  openLoc,
			Close: closeLoc,
		}
	}

	return nil
}

func (p *Parser) parseCurlyList(f ParseFunc) *ast.CurlyList {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if nodes, openLoc, closeLoc, _ := p.parseBracketedList(
		f,
		token.LCurly,
		token.RCurly,
		token.Semicolon,
		token.NewLine,
	); nodes != nil {
		return &ast.CurlyList{
			StmtList: &ast.StmtList{Nodes: nodes},
			Open:     openLoc,
			Close:    closeLoc,
		}
	}

	return nil
}

func (p *Parser) parseBlock() *ast.CurlyList {
	return p.parseBlockFunc(p.parseStmt)
}

func (p *Parser) parseBlockFunc(f ParseFunc) *ast.CurlyList {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if !p.match(token.LCurly) {
		p.error(ErrorExpectedBlock)
		return nil
	}

	if list := p.parseCurlyList(f); list != nil {
		return list
	}

	return nil
}

func (p *Parser) parseDeclList() *ast.StmtList {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if nodes, _ := p.listWithDelimiter(
		p.parseDecl,
		token.EOF,
		token.Semicolon,
		token.NewLine,
	); len(nodes) > 0 {
		return &ast.StmtList{Nodes: nodes}
	}

	return nil
}

//------------------------------------------------
// Helper functions
//------------------------------------------------

func (p *Parser) listWithDelimiter(
	f ParseFunc,
	delimiter token.Kind,
	separators ...token.Kind,
) (nodes []ast.Node, wasSeparator bool) {
	if len(separators) < 1 {
		panic("expect at least 1 separator")
	}

	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	nodes = []ast.Node{}

	// List = Expr {Separator Expr} [Separator]
	for {
		// Element can start at new line.
		for p.tok.Kind == token.NewLine {
			p.next()
		}

		// Possible cases:
		//  - empty list `{}`
		//  - list closes after separators `{x,}`
		//  - unterminated list `{...`
		if p.tok.Kind == delimiter {
			break
		} else if p.tok.Kind == token.EOF {
			return nil, false
		}

		nodeStart := p.tok.Start

		if node := f(); node != nil {
			switch {
			case p.consume(separators...) != nil:
				wasSeparator = true
				fallthrough

			case p.match(delimiter):
				nodes = append(nodes, node)
				continue

			default:
				// [parseFunc] set the correct node, but no separator was found.
				// Report it and assign [ast.BadNode] instead.
				p.errorAt(ErrorUnterminatedExpr, node.Pos(), node.PosEnd())
			}
		}

		// Something went wrong, advance to some delimiter and
		// continue parsing elements until we find the [closing] token.
		p.skip(append(separators, delimiter)...)
		p.consume(separators...)
		nodes = append(nodes, &ast.BadNode{DesiredPos: nodeStart})
	}

	return nodes, wasSeparator
}

func (p *Parser) parseBracketedList(
	f ParseFunc,
	opening, closing token.Kind,
	separators ...token.Kind,
) (nodes []ast.Node, openLoc, closeLoc token.Pos, wasSeparator bool) {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if tok := p.expect(opening); tok != nil {
		openLoc = tok.Start
	} else {
		return nil, token.Pos{}, token.Pos{}, false
	}

	nodes, wasSeparator = p.listWithDelimiter(f, closing, separators...)

	if nodes == nil {
		return nil, token.Pos{}, token.Pos{}, false
	}

	if tok := p.consume(closing); tok != nil {
		closeLoc = tok.Start
	} else {
		if p.tok.Kind == token.EOF {
			p.errorAt(ErrorBracketIsNeverClosed, openLoc, openLoc)
		} else {
			start, end := p.skip()
			p.errorExpectedTokenAt(start, end, append(separators, closing)...)
		}
		return nil, token.Pos{}, token.Pos{}, false
	}

	return
}
