package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/token"
)

func (p *Parser) parseIdent() *ast.Ident {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if id := p.expect(token.Ident); id != nil {
		return &ast.Ident{
			Name:  id.Data,
			Start: id.Start,
			End:   id.End,
		}
	}

	return nil
}

func (p *Parser) parseIdentNode() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	return p.parseIdent()
}

func (p *Parser) parseAttribute() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if tok := p.expect(token.Attribute); tok != nil {
		identStart := tok.Start
		identStart.Char += 1
		ident := &ast.Ident{
			Name:  tok.Data[1:],
			Start: identStart,
			End:   tok.End,
		}

		return &ast.Attribute{
			Name: ident,
			X:    p.parseExpr(),
			Loc:  tok.Start,
		}
	}

	return nil
}

func (p *Parser) parseLiteral() ast.Node {
	if tok := p.expect(token.Int, token.Float, token.String); tok != nil {
		return &ast.Literal{
			Kind:  tok.Kind,
			Value: tok.Data,
			Start: tok.Start,
			End:   tok.End,
		}
	}

	return nil
}

//

func (p *Parser) parseMemberAccess(x ast.Node) ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if x == nil {
		panic("can't use nil node as left-hand side expression")
	}

	if dot := p.consume(token.Dot); dot != nil {
		y := ast.Node(nil)

		if star := p.consume(token.Asterisk); star != nil {
			// special case for `.*`
			y = &ast.Star{Loc: star.Start}
		} else {
			y = p.parseIdent()
		}

		if y == nil {
			return nil
		}

		x = p.parseMemberAccess(&ast.MemberAccess{
			Loc:      dot.Start,
			X:        x,
			Selector: y,
		})
	}

	return x
}

func (p *Parser) parseSuffixExpr(x ast.Node) ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if x == nil {
		x = p.parseOperand()
	}

	for {
		switch p.tok.Kind {
		case token.Dot:
			x = p.parseMemberAccess(x)

		case token.QuestionMark:
			x = &ast.Try{
				X:   x,
				Loc: p.consume().Start,
			}

		case token.Bang:
			x = &ast.Unwrap{
				X:   x,
				Loc: p.consume().Start,
			}

		case token.LBracket:
			x = &ast.Index{
				X:    x,
				Args: p.parseBracketList(p.parseExpr, token.Comma),
			}

		case token.LParen:
			x = &ast.Call{
				X:    x,
				Args: p.parseParenList(p.parseExpr),
			}

		default:
			return x
		}
	}
}

//

func (p *Parser) parseOperand() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.Ident:
		return p.parseIdentNode()

	case token.Attribute:
		return p.parseAttribute()

	case token.Int, token.Float, token.String:
		return p.parseLiteral()

	case token.LParen:
		return p.parseParenList(p.parseExpr, token.Comma)

	default:
		p.errorExpected("operand", p.tok.Start, p.tok.End)
		return nil
	}
}

func (p *Parser) parsePrimaryExpr(x ast.Node) ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if x == nil {
		x = p.parseOperand()
	}

	if x == nil {
		return nil
	}

	return p.parseSuffixExpr(x)
}

func (p *Parser) parseUnaryExpr() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.Minus:
		tok := p.consume()
		return &ast.UnaryOp{
			X:      p.parseUnaryExpr(),
			Loc:    tok.Start,
			OpKind: tok.Kind,
		}

	case token.Amp:
		ampPos := p.consume().Start
		varPos := token.Loc{}

		if varTok := p.consume(token.KwVar); varTok != nil {
			varPos = varTok.Start
		}

		return &ast.Ref{
			X:      p.parseUnaryExpr(),
			Loc:    ampPos,
			VarLoc: varPos,
		}

	default:
		return p.parsePrimaryExpr(nil)
	}
}

func (p *Parser) parseBinaryExpr(x ast.Node, precedence token.Precedence) ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if x == nil {
		x = p.parseUnaryExpr()
	}

	if x == nil {
		return nil
	}

	for p.tok.Precedence() >= precedence {
		tok := p.consume()
		y := p.parseBinaryExpr(nil, tok.Precedence()+1)

		if y == nil {
			return nil
		}

		x = &ast.BinaryOp{
			X:      x,
			Y:      y,
			Loc:    tok.Start,
			OpKind: tok.Kind,
		}
	}

	return x
}

//

func (p *Parser) parseSimpleExpr() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	return p.parseBinaryExpr(nil, token.LowestPrec+1)
}

func (p *Parser) parseExpr() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.KwIf:
		return p.parseIf()

	case token.KwWhile:
		return p.parseWhile()

	case token.LCurly:
		return p.parseBlock()

	default:
		return p.parseSimpleExpr()
	}
}

func (p *Parser) parseComplexExpr() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	for {
		for p.tok.Kind == token.NewLine {
			p.next()
		}

		if annotation := p.parseAnnotation(); annotation != nil {
			p.annots = append(p.annots, annotation)
		} else {
			break
		}
	}

	node := ast.Node(nil)

	switch p.tok.Kind {
	case token.KwVar, token.KwVal, token.KwConst:
		node = p.parseGenericDecl()

	case token.KwFunc:
		node = p.parseFuncDecl()

	case token.KwEnum:
		node = p.parseEnumDecl()

	case token.KwModule:
		node = p.parseModule()

	case token.KwAlias:
		node = p.parseAlias()

	case token.KwReturn:
		node = p.parseReturn()

	case token.KwBreak:
		node = p.parseBreak()

	case token.KwContinue:
		node = p.parseContinue()

	default:
		node = p.parseExpr()
	}

	if len(p.annots) > 0 {
		if decl, isDecl := node.(ast.Decl); isDecl {
			setAnnotations(decl, p.annots)
		} else {
			p.addError(NewError(
				"unexpected annotation",
				p.annots[0].Pos(),
				p.annots[0].PosEnd(),
				"",
				"only a declaration can have annotation",
			))
		}
		p.annots = nil
	}

	return node
}

func (p *Parser) parseStmt() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if p.tok.Kind == token.Semicolon || p.tok.Kind == token.EOF {
		p.next()
		return &ast.Empty{Loc: p.tok.Start}
	}

	return p.parseComplexExpr()
}

//

func (p *Parser) parseAnnotation() *ast.Annotation {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if tok := p.consume(token.At); tok != nil {
		if ident := p.parseIdent(); ident != nil {
			args := (*ast.ParenList)(nil)

			if p.match(token.LParen) {
				args = p.parseParenList(p.parseStmt, token.Comma)
			}

			return &ast.Annotation{
				Loc:  tok.Start,
				Name: ident,
				Args: args,
			}
		}
	}

	return nil
}

func (p *Parser) parseBlock() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if list := p.parseCurlyList(p.parseStmt); list != nil {
		return list
	}

	return nil
}

func (p *Parser) parseField() ast.Node {
	// Field <- IdentifierList (Type? '=' Expr | Type ('=' Expr)?)
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	nameNodes := p.parseExprList(p.parseIdentNode, token.Comma)
	typ := ast.Node(nil)

	if p.tok.Kind != token.Eq {
		typ = p.parseType()
		if typ == nil {
			start, end := p.skipTo()
			p.errorExpected("type name or '='", start, end)
			return nil
		}
	}

	names := make([]*ast.Ident, 0, len(nameNodes.Nodes))
	value := ast.Node(nil)
	tokIdx := -1
	parseValue := false

	if typ == nil {
		p.expect(token.Eq)
		parseValue = true
	} else if p.consume(token.Eq) != nil {
		tokIdx = p.save()
		parseValue = true
	}

	if parseValue {
		p.consume(token.NewLine)
		value = p.parseExpr()

		if value == nil {
			if tokIdx >= 0 {
				p.restore(tokIdx)
			}

			start, end := p.skipTo(endOfStmtKinds...)
			p.errorExpected("expression", start, end)
			return nil
		}
	}

	for _, name := range nameNodes.Nodes {
		if n, ok := name.(*ast.Ident); ok {
			names = append(names, n)
		} else {
			panic("unreachable?")
		}
	}

	assert.Ok(typ != nil || value != nil)

	return &ast.Field{
		Names: names,
		Type:  typ,
		Value: value,
	}
}

func (p *Parser) parseTypeName() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.Ident:
		switch expr := p.parseMemberAccess(p.parseIdent()); expr.(type) {
		case *ast.Ident:
			return expr

		default:
			start, end := p.skipTo()
			p.errorExpected("type name", start, end)
			return nil
		}

	default:
		start, end := p.skipTo()
		p.error("expected type name", start, end)
		return nil
	}
}

func (p *Parser) parseSignature(funcTok *token.Token) *ast.Signature {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	paramList := p.parseParenList(p.parseField, token.Comma)

	if paramList == nil {
		return nil
	}

	returnType := ast.Node(nil)

	if p.tok.Kind != token.Arrow && p.tok.Kind != token.LCurly {
		returnType = p.parseType()

		if returnType == nil {
			return nil
		}
	}

	tokPos := token.Loc{}

	if funcTok != nil {
		tokPos = funcTok.Start
	}

	return &ast.Signature{
		Loc:    tokPos,
		Params: paramList,
		Result: returnType,
	}
}

func (p *Parser) parseType() ast.Node {
	// Type <- ('&' 'var'?)? TypeName ('.' TypeName)
	// TypeName <- Macro | Ident
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.Amp:
		ampPos := p.expect().Start
		varPos := token.Loc{}
		if varTok := p.consume(token.KwVar); varTok != nil {
			varPos = varTok.Start
		}
		return &ast.Ref{
			X:      p.parseType(),
			Loc:    ampPos,
			VarLoc: varPos,
		}

	case token.LBracket:
		openPos := p.consume().Start
		n := p.parseSimpleExpr()
		closePos := token.Loc{}
		if closeTok := p.consume(token.RBracket); closeTok != nil {
			closePos = closeTok.Start
		}
		return &ast.ArrayType{
			X:     p.parseType(),
			N:     n,
			Open:  openPos,
			Close: closePos,
		}

	case token.Ident:
		return p.parseTypeName()

	case token.KwFunc:
		return p.parseSignature(p.expect(token.KwFunc))

	default:
		return nil
	}
}

//

func (p *Parser) parseGenericDecl() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwVar, token.KwVal, token.KwConst)
	field, ok := p.parseField().(*ast.Field)

	if !ok {
		return nil
	}

	return &ast.GenericDecl{
		Loc:   tok.Start,
		Kind:  tok.Kind,
		Field: field,
	}
}

func (p *Parser) parseFuncDecl() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwFunc)
	name := p.parseIdent()

	if name == nil {
		return nil
	}

	signature := p.parseSignature(nil)

	if signature == nil {
		return nil
	}

	body := ast.Node(nil)

	if p.consume(token.Arrow) != nil {
		body = p.parseExpr()

		if body == nil {
			start, end := p.skipTo()
			p.errorExpected("expression", start, end)
			return nil
		}
	} else if p.tok.Kind == token.LCurly {
		body = p.parseBlock()

		if body == nil {
			start, end := p.skipTo()
			p.errorExpected("body", start, end)
			return nil
		}
	} else if p.tok.Kind != token.NewLine && p.tok.Kind != token.EOF {
		start, end := p.skipTo()
		p.errorExpectedToken(start, end, token.Arrow, token.LCurly, token.NewLine)
	}

	if body == nil {
		return nil
	}

	return &ast.FuncDecl{
		Loc:       tok.Start,
		Name:      name,
		Signature: signature,
		Body:      body,
	}
}

func (p *Parser) parseEnumDecl() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	enumTok := p.expect(token.KwEnum)

	if enumTok == nil {
		return nil
	}

	name := p.parseIdent()

	if name == nil {
		return nil
	}

	body := p.parseCurlyList(p.parseSimpleExpr)

	if body == nil {
		return nil
	}

	p.validateEnumBody(body)

	return &ast.EnumDecl{
		Name: name,
		Body: body,
		Loc:  enumTok.Start,
	}
}

func (p *Parser) validateEnumBody(body *ast.CurlyList) {
	for _, expr := range body.Nodes {
		switch e := expr.(type) {
		case *ast.Ident:
			continue

		case *ast.BinaryOp:
			if e.OpKind == token.Eq {
				continue
			}
		}

		p.error("expected identifier or assignment", expr.Pos(), expr.PosEnd())
	}
}

func (p *Parser) parseElse() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if elseTok := p.consume(token.KwElse); elseTok != nil {
		body := ast.Node(nil)

		if p.tok.Kind == token.KwIf {
			body = p.parseIf()
		} else {
			body = p.parseBlock()

			if body == nil {
				start, end := p.skipTo()
				p.error("expected `if` clause or block after `else`", start, end)
				return nil
			}
		}

		return &ast.Else{
			Loc:  elseTok.Start,
			Body: body,
		}
	}

	return (*ast.Else)(nil)
}

func (p *Parser) parseIf() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwIf)

	if tok == nil {
		return nil
	}

	cond := p.parseSimpleExpr()

	if cond == nil {
		start, end := p.skipTo()
		p.error("expected conditional expression for `if` clause", start, end)
		return nil
	}

	body := p.parseBlock()

	if body == nil {
		start, end := p.skipTo()
		p.error("expected body for 'if' clause", start, end)
		return nil
	}

	elseNode := p.parseElse()
	elseClause, ok := elseNode.(*ast.Else)

	if !ok {
		return nil
	}

	return &ast.If{
		Loc:  tok.Start,
		Cond: cond,
		Body: body,
		Else: elseClause,
	}
}

func (p *Parser) parseWhile() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwWhile)
	cond := p.parseExpr()
	body := p.parseBlock()

	return &ast.While{
		Loc:  tok.Start,
		Cond: cond,
		Body: body,
	}
}

func (p *Parser) parseReturn() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwReturn)
	x := p.parseExpr()

	return &ast.Return{
		X:   x,
		Loc: tok.Start,
	}
}

func (p *Parser) parseBreak() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwBreak)
	label := (*ast.Ident)(nil)

	if p.tok.Kind == token.Ident {
		label = p.parseIdent()
	}

	return &ast.Break{
		Label: label,
		Loc:   tok.Start,
	}
}

func (p *Parser) parseContinue() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwContinue)
	label := (*ast.Ident)(nil)

	if p.tok.Kind == token.Ident {
		label = p.parseIdent()
	}

	return &ast.Continue{
		Label: label,
		Loc:   tok.Start,
	}
}

func (p *Parser) parseModule() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if tok := p.consume(token.KwModule); tok != nil {
		if name := p.parseIdent(); name != nil {
			body := ast.Node(nil)

			if p.tok.Kind == token.LCurly {
				body = p.parseBlock()
			}

			return &ast.ModuleDecl{
				Name: name,
				Body: body,
				Loc:  tok.Start,
			}
		}
	}

	p.error("first statement in the file should be `module`", p.tok.Start, p.tok.End)
	return nil
}

func (p *Parser) parseAlias() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if tokAlias := p.consume(token.KwAlias); tokAlias != nil {
		id := p.parseIdent()

		if id == nil {
			return nil
		}

		p.expect(token.Eq)
		p.consume(token.NewLine)

		expr := p.parseType()

		if expr == nil {
			p.errorExpected("type", p.tok.Start, p.tok.End)
			return nil
		}

		return &ast.AliasDecl{
			Loc:  tokAlias.Start,
			Name: id,
			Expr: expr,
		}
	}

	return nil
}

//

func (p *Parser) parseExprList(
	parseFunc func() ast.Node,
	separators ...token.Kind,
) *ast.List {
	if len(separators) < 1 {
		panic("parser.parseExprList: expect at least 1 separator")
	}

	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	nodes := []ast.Node{}

	for {
		for p.tok.Kind == token.NewLine {
			p.next()
		}

		if node := parseFunc(); node != nil {
			nodes = append(nodes, node)
		} else if p.consume(separators...) == nil {
			nodes = append(nodes, &ast.BadNode{Loc: p.tok.Start})
			p.skipTo(append(endOfStmtKinds, separators...)...)
		}

		if p.consume(separators...) == nil {
			break
		}
	}

	return &ast.List{Nodes: nodes}
}

// Not stops is no separator was found. Iterates until a `closing` token.
func (p *Parser) parseClosingExprList(
	parseFunc func() ast.Node,
	closing token.Kind,
	separators ...token.Kind,
) *ast.List {
	if len(separators) < 1 {
		panic("expect at least 1 separator")
	}

	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	nodes := []ast.Node{}
	delimiters := append([]token.Kind{closing}, separators...)

	// List :: `{` [Expr {Separator Expr} [Separator]] `}`
	for {
		// Element can start at new line
		for p.tok.Kind == token.NewLine {
			p.next()
		}

		// Possible cases:
		//  - empty list `{}`
		//  - list closes after separators `{x,}`
		//  - unterminated list `{...`
		if p.tok.Kind == closing {
			break
		} else if p.tok.Kind == token.EOF {
			return nil
		}

		// Save expression start position for the bad node.
		exprStartPos := p.tok.Start

		if node := parseFunc(); node != nil {
			if p.consume(separators...) != nil || p.match(closing) {
				// All is OK.
				nodes = append(nodes, node)
				continue
			}

			// [parseFunc] jet the correct node, but no separator was found.
			// Report it and assign [ast.BadNode].
			p.error("unterminated expression", node.Pos(), node.PosEnd())
		}

		// Something went wrong, advance to some delimiter and
		// continue parsing elements until we find the [closing] token.
		p.skipTo(delimiters...)
		p.consume(separators...)
		nodes = append(nodes, &ast.BadNode{Loc: exprStartPos})
	}

	return &ast.List{Nodes: nodes}
}

type AnyList interface {
	ast.ParenList | ast.CurlyList | ast.BracketList
}

// If `separators` are not specified, [token.Semicolon] and [token.NewLine]
// will be used as separators.
func parseList[T AnyList](
	p *Parser,
	parseFunc func() ast.Node,
	opening, closing token.Kind,
	separators ...token.Kind,
) *T {
	if len(separators) == 0 {
		separators = []token.Kind{token.Semicolon, token.NewLine}
	}

	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tokOpening := p.expect(opening)

	if tokOpening == nil {
		return nil
	}

	// parseFuncOrClosingToken := func() ast.Node {
	// 	if p.tok.Kind == closing {
	// 		return nil
	// 	}

	// 	return parseFunc()
	// }

	exprList := p.parseClosingExprList(parseFunc, closing, separators...)

	if exprList == nil {
		return nil
	}

	tokClose := p.consume(closing)

	if tokClose == nil {
		if p.tok.Kind == token.EOF {
			p.error("bracket is never closed (end of file reached)", tokOpening.Start, tokOpening.End)
		} else {
			start, end := p.skipTo()
			p.errorExpectedToken(start, end, append(separators, closing)...)
		}

		return nil
	}

	return &T{
		Open:  tokOpening.Start,
		Close: tokClose.Start,
		List:  exprList,
	}
}

func (p *Parser) parseParenList(parseFunc func() ast.Node, separators ...token.Kind) *ast.ParenList {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	return parseList[ast.ParenList](p, parseFunc, token.LParen, token.RParen, separators...)
}

func (p *Parser) parseCurlyList(parseFunc func() ast.Node, separators ...token.Kind) *ast.CurlyList {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	return parseList[ast.CurlyList](p, parseFunc, token.LCurly, token.RCurly, separators...)
}

func (p *Parser) parseBracketList(parseFunc func() ast.Node, separators ...token.Kind) *ast.BracketList {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	return parseList[ast.BracketList](p, parseFunc, token.LBracket, token.RBracket, separators...)
}

func (p *Parser) parseStmtList() *ast.List {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	return p.parseExprList(p.parseStmt, endOfStmtKinds...)
}
