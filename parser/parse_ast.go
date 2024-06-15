package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

type parseFunc func() ast.Node

//------------------------------------------------
// Primitives
//------------------------------------------------

func (p *parser) parseIdent() ast.Node {
	if node := p.parseIdentNode(); node != nil {
		return node
	}

	p.errorExpectedToken(token.Ident)
	return nil
}

func (p *parser) parseLiteral() ast.Node {
	if node := p.parseLiteralNode(); node != nil {
		return node
	}

	p.errorExpectedToken(token.Int, token.Float, token.String)
	return nil
}

//------------------------------------------------
// Statements
//------------------------------------------------

func (p *parser) declOr(f parseFunc) parseFunc {
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

func (p *parser) parseStmt() ast.Node {
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

func (p *parser) parseExpr() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	switch p.tok.Kind {
	case token.KwDefer:
		return p.parseDefer()

	case token.KwReturn:
		return p.parseReturn()

	case token.KwBreak:
		return p.parseBreak()

	case token.KwContinue:
		return p.parseContinue()

	default:
		return p.parseSimpleExpr(true)
	}
}

func (p *parser) parseSimpleExpr(allowAssign bool) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	var x ast.Node

	if p.tok.Kind == token.LParen {
		params := p.parseParenList(p.declOr(p.parseExpr))
		if params == nil {
			return nil
		}

		if p.match(token.Arrow) || p.match(exprStartKinds...) {
			return p.parseFunction(params)
		}

		x = params
	}

	if allowAssign {
		return p.parseBinaryExpr(x, 1)
	}

	return p.parseBinaryExpr(x, 2)
}

func (p *parser) parseBinaryExpr(x ast.Node, precedence int) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if x == nil {
		if x = p.parsePrimaryExpr(); x == nil {
			return nil
		}
	}

	for p.tok.Precedence() >= precedence {
		tok := p.consume()

		for p.tok.Kind == token.NewLine {
			p.next()
		}

		y := p.parseBinaryExpr(nil, tok.Precedence()+1)
		if y == nil {
			return nil
		}

		binaryOpKind, ok := operators[tok.Kind]
		if !ok {
			p.errorfAt(
				ErrorInvalidBinaryOperator,
				tok.Start,
				tok.End,
				"%s cannot be used in the binary expression",
				tok.Kind.UserString(),
			)
		}

		x = &ast.Op{
			X:     x,
			Y:     y,
			Start: tok.Start,
			End:   tok.End,
			Kind:  binaryOpKind,
		}
	}

	return x
}

func (p *parser) parsePrimaryExpr() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	return p.parsePrefixExpr()
}

func (p *parser) parsePrefixExpr() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	switch p.tok.Kind {
	case token.Minus:
		minus := p.consume()

		return &ast.Op{
			X:     nil,
			Y:     p.parsePrefixExpr(),
			Start: minus.Start,
			End:   minus.End,
			Kind:  ast.OperatorNeg,
		}

	case token.Bang:
		bang := p.consume()

		return &ast.Op{
			X:     nil,
			Y:     p.parsePrefixExpr(),
			Start: bang.Start,
			End:   bang.End,
			Kind:  ast.OperatorNot,
		}

	case token.Asterisk:
		asterisk := p.consume()

		if tokMut := p.consume(token.KwMut); tokMut != nil {
			return &ast.Op{
				X:     nil,
				Y:     p.parsePrefixExpr(),
				Start: asterisk.Start,
				End:   tokMut.End,
				Kind:  ast.OperatorMutPtr,
			}
		}

		return &ast.Op{
			X:     nil,
			Y:     p.parsePrefixExpr(),
			Start: asterisk.Start,
			End:   asterisk.End,
			Kind:  ast.OperatorPtr,
		}

	case token.Amp:
		amp := p.consume()

		if tokMut := p.consume(token.KwMut); tokMut != nil {
			return &ast.Op{
				X:     nil,
				Y:     p.parsePrefixExpr(),
				Start: amp.Start,
				End:   tokMut.End,
				Kind:  ast.OperatorMutAddrOf,
			}
		}

		return &ast.Op{
			X:     nil,
			Y:     p.parsePrefixExpr(),
			Start: amp.Start,
			End:   amp.End,
			Kind:  ast.OperatorAddrOf,
		}

	case token.LBracket:
		list := p.parseBracketList(p.parseExpr)
		if list == nil {
			return nil
		}

		if p.match(simpleExprStartKinds...) {
			return &ast.ArrayType{
				X:    p.parsePrefixExpr(),
				Args: list,
			}
		}

		return list

	default:
		return p.parseSuffixExpr(nil)
	}
}

func (p *parser) parseSuffixExpr(x ast.Node) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if x == nil {
		if x = p.parseOperand(); x == nil {
			return nil
		}
	}

	for {
		if x == nil {
			return nil
		}

		switch p.tok.Kind {
		case token.Dot:
			x = p.parseDot(x)

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

		default:
			return x
		}
	}
}

func (p *parser) parseOperand() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	switch p.tok.Kind {
	case token.Ident:
		return p.parseIdent()

	case token.Int, token.Float, token.String:
		return p.parseLiteral()

	case token.Dollar:
		return p.parseBuiltIn()

	case token.KwIf:
		return p.parseIf()

	case token.KwWhile:
		return p.parseWhile()

	case token.KwFor:
		return p.parseFor()

	case token.KwStruct:
		return p.parseStructType()

	case token.KwEnum:
		return p.parseEnumType()

	case token.LCurly:
		if block := p.parseBlock(); block != nil {
			return block
		}

	case token.LBracket:
		if list := p.parseBracketList(p.parseExpr); list != nil {
			return list
		}

	case token.LParen:
		if list := p.parseParenList(p.declOr(p.parseExpr)); list != nil {
			return list
		}
	}

	p.error(ErrorExpectedOperand)
	return nil
}

//------------------------------------------------
// Complex expressions
//------------------------------------------------

func (p *parser) parseEllipsisExpr() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if tok := p.consume(token.Ellipsis); tok != nil {
		var y ast.Node

		if p.match(simpleExprStartKinds...) {
			if y = p.parseSimpleExpr(false); y == nil {
				return nil
			}
		}

		return &ast.Op{
			X:     nil,
			Y:     y,
			Start: tok.Start,
			End:   tok.End,
			Kind:  ast.OperatorEllipsis,
		}
	}

	return p.parseSimpleExpr(false)
}

func (p *parser) parseFunction(params *ast.ParenList) ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	var result, body ast.Node

	if p.consume(token.Arrow) != nil {
		// (...) -> T
		// (...) -> T {...}
		result = p.parseSimpleExpr(false)
		if result == nil {
			return nil
		}

		if p.match(token.LCurly) {
			if block := p.parseBlock(); block != nil {
				body = block
			}
		}
	} else if p.match(exprStartKinds...) {
		// (...) expr
		if expr := p.parseExpr(); expr != nil {
			body = expr
		}
	}

	if body == nil {
		if result == nil {
			return nil
		}

		return &ast.Signature{
			Params: params,
			Result: result,
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

func (p *parser) parseDot(x ast.Node) ast.Node {
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

func (p *parser) parseBuiltIn() ast.Node {
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

func (p *parser) parseDecl() ast.Node {
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

func (p *parser) parseDeclNode(mut token.Pos, name *ast.Ident) *ast.Decl {
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
		ty = p.parseEllipsisExpr()
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
// Language constructions
//------------------------------------------------

func (p *parser) parseAttributeListNode() *ast.AttributeList {
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

func (p *parser) parseStructType() ast.Node {
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

func (p *parser) parseEnumType() ast.Node {
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

func (p *parser) parseIf() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwIf)

	if tok == nil {
		return nil
	}

	cond := p.parseSimpleExpr(false)

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

func (p *parser) parseElse() ast.Node {
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

func (p *parser) parseWhile() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwWhile)
	if tok == nil {
		return nil
	}

	cond := p.parseSimpleExpr(false)
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

func (p *parser) parseFor() ast.Node {
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

	iterExpr := p.parseSimpleExpr(false)
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

func (p *parser) parseForLoopDeclList() *ast.List {
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

func (p *parser) parseForLoopDecl() ast.Node {
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
		ty = p.parseExpr()
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

func (p *parser) parseDefer() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwDefer)
	if tok == nil {
		return nil
	}

	x := p.parseExpr()
	if x == nil {
		return nil
	}

	return &ast.Defer{X: x, TokPos: tok.Start}
}

func (p *parser) parseReturn() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwReturn)
	if tok == nil {
		return nil
	}

	var x ast.Node

	if !p.match(append(endOfStmtKinds, token.EOF)...) {
		x = p.parseExpr()
		if x == nil {
			return nil
		}
	}

	return &ast.Return{X: x, TokPos: tok.Start}
}

func (p *parser) parseBreak() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwBreak)
	if tok == nil {
		return nil
	}

	return &ast.Break{
		Label:  p.parseIdentNode(),
		TokPos: tok.Start,
	}
}

func (p *parser) parseContinue() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	tok := p.expect(token.KwContinue)
	if tok == nil {
		return nil
	}

	return &ast.Continue{
		Label:  p.parseIdentNode(),
		TokPos: tok.Start,
	}
}

//------------------------------------------------
// Lists
//------------------------------------------------

func (p *parser) parseBracketList(f parseFunc) *ast.BracketList {
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

func (p *parser) parseParenList(f parseFunc) *ast.ParenList {
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

func (p *parser) parseCurlyList(f parseFunc) *ast.CurlyList {
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

func (p *parser) parseBlock() *ast.CurlyList {
	return p.parseBlockFunc(p.parseStmt)
}

func (p *parser) parseBlockFunc(f parseFunc) *ast.CurlyList {
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

func (p *parser) parseDeclList() *ast.StmtList {
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

func (p *parser) listWithDelimiter(
	f parseFunc,
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

func (p *parser) parseBracketedList(
	f parseFunc,
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
