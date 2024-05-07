package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

//------------------------------------------------
// Primitives
//------------------------------------------------

func (p *Parser) parseIdentNode() *ast.Ident {
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

func (p *Parser) parseIdent() ast.Node {
	if ident := p.parseIdentNode(); ident != nil {
		return ident
	}

	return nil
}

func (p *Parser) parseLiteral() ast.Node {
	if tok := p.expect(token.Int, token.Float, token.String); tok != nil {
		litKind := ast.UnknownLiteral

		switch tok.Kind {
		case token.Int:
			litKind = ast.IntLiteral

		case token.Float:
			litKind = ast.FloatLiteral

		case token.String:
			litKind = ast.StringLiteral

		default:
			panic("unreachable")
		}

		return &ast.Literal{
			Kind:  litKind,
			Value: tok.Data,
			Start: tok.Start,
			End:   tok.End,
		}
	}

	return nil
}

//------------------------------------------------
// Statements
//------------------------------------------------

func (p *Parser) parseStmt() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	for p.tok.Kind == token.NewLine {
		p.next()
	}

	attributes := p.parseAttributes()

	for p.tok.Kind == token.NewLine {
		p.next()
	}

	node := ast.Node(nil)

	switch p.tok.Kind {
	case token.Semicolon:
		node = p.parseEmptyStmt()

	case token.KwVar:
		node = p.parseVarDecl()

	// case token.KwVar, token.KwVal, token.KwConst:
	// 	node = p.parseGenericDecl()

	case token.KwWhile:
		return p.parseWhile()

	case token.KwFunc:
		node = p.parseFuncDecl()

	case token.KwModule:
		node = p.parseModule()

	case token.KwAlias:
		node = p.parseAlias()

	case token.KwReturn:
		node = p.parseReturn()

	case token.KwBreak, token.KwContinue:
		node = p.parseBreakOrContinue()

	default:
		node = p.parseExpr()
	}

	if attributes != nil {
		if decl, isDecl := node.(ast.Decl); isDecl {
			setAttributes(decl, attributes)
		} else {
			p.error(
				attributes.Pos(),
				attributes.LocEnd(),
				"unexpected attribute list (only a declaration can have attributes)",
			)
		}
	}

	// if decl, isDecl := node.(ast.Decl); isDecl {
	// 	setDoc(decl, p.commentGroup)
	// }
	// p.commentGroup = nil

	return node
}

func (p *Parser) parseEmptyStmt() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if p.tok.Kind == token.Semicolon {
		p.next()
		return &ast.Empty{Loc: p.tok.Start}
	}

	return nil
}

//------------------------------------------------
// Expressions
//------------------------------------------------

func (p *Parser) parseExpr() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.KwIf:
		return p.parseIf()

	case token.LCurly:
		return p.parseBlock()

	case token.LBracket:
		return p.parseBracketExpr()

	default:
		return p.parseSimpleExpr()
	}
}

func (p *Parser) parseSimpleExpr() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	return p.parseBinaryExpr(nil, token.LowestPrec+1)
}

func (p *Parser) parseBinaryExpr(lhs ast.Node, precedence token.Precedence) ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if lhs == nil {
		lhs = p.parseUnaryExpr()
	}

	if lhs == nil {
		return nil
	}

	for p.tok.Precedence() >= precedence {
		tok := p.consume()
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

		default:
			p.errorf(
				tok.Start,
				tok.End,
				"%s cannot be used in the binary expression",
				tok.Kind.UserString(),
			)
		}

		lhs = &ast.InfixOp{
			X: lhs,
			Y: rhs,
			Opr: &ast.Operator{
				Start: tok.Start,
				End:   tok.End,
				Kind:  binaryOpKind,
			},
		}
	}

	return lhs
}

func (p *Parser) parseUnaryExpr() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.Minus:
		loc := p.consume().Start

		return &ast.PrefixOp{
			X: p.parseUnaryExpr(),
			Opr: &ast.Operator{
				Start: loc,
				End:   loc,
				Kind:  ast.OperatorNeg,
			},
		}

	case token.Bang:
		loc := p.consume().Start

		return &ast.PrefixOp{
			X: p.parseUnaryExpr(),
			Opr: &ast.Operator{
				Start: loc,
				End:   loc,
				Kind:  ast.OperatorNot,
			},
		}

	case token.Amp:
		loc := p.consume().Start

		if varTok := p.consume(token.KwVar); varTok != nil {
			return &ast.PrefixOp{
				X: p.parseUnaryExpr(),
				Opr: &ast.Operator{
					Start: loc,
					End:   varTok.End,
					Kind:  ast.OperatorMutAddr,
				},
			}
		}

		return &ast.PrefixOp{
			X: p.parseUnaryExpr(),
			Opr: &ast.Operator{
				Start: loc,
				End:   loc,
				Kind:  ast.OperatorAddr,
			},
		}

	default:
		return p.parsePrimaryExpr(nil)
	}
}

func (p *Parser) parseOperand() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.Ident:
		return p.parseIdent()

	case token.At:
		return p.parseBuiltIn()

	case token.Int, token.Float, token.String:
		return p.parseLiteral()

	case token.LParen:
		return p.parseParenList(p.parseExpr)

	default:
		p.errorExpected(p.tok.Start, p.tok.End, "operand")
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

		case token.QuestionMark, token.Bang:
			opr := p.consume()
			postfixOpKind := ast.UnknownOperator

			switch opr.Kind {
			case token.QuestionMark:
				postfixOpKind = ast.OperatorTry

			case token.Bang:
				postfixOpKind = ast.OperatorUnwrap

			default:
				p.errorf(
					opr.Start,
					opr.End,
					"%s can't be used as postfix operator",
					opr.Kind.UserString(),
				)
			}

			x = &ast.PostfixOp{
				X: x,
				Opr: &ast.Operator{
					Start: opr.Start,
					End:   opr.End,
					Kind:  postfixOpKind,
				},
			}

		case token.LBracket:
			x = &ast.Index{
				X:    x,
				Args: p.parseBracketList(p.parseExpr),
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

//------------------------------------------------
// Complex expressions
//------------------------------------------------

func (p *Parser) parseMemberAccess(x ast.Node) ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if x == nil {
		panic("can't use nil node as left-hand side expression")
	}

	if dot := p.consume(token.Dot); dot != nil {
		y := p.parseIdentNode()

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

func (p *Parser) parseBuiltIn() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if tok := p.expect(token.At); tok != nil {
		name := p.parseIdentNode()
		x := ast.Node(nil)

		switch p.tok.Kind {
		case token.LParen:
			if list := p.parseParenList(p.parseExpr); list != nil {
				x = list
			}

		case token.LCurly:
			x = p.parseBlock()

		default:
			p.errorExpected(
				p.tok.Start,
				p.tok.End,
				"builtin function call requires argument list or block",
			)
		}

		if x != nil {
			return &ast.BuiltInCall{
				Name: name,
				Args: x,
				Loc:  tok.Start,
			}
		}
	}

	return nil
}

//------------------------------------------------
// Declarations
//------------------------------------------------

func (p *Parser) parseVarDecl() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwVar)
	binding, ok := p.parseBinding().(*ast.Binding)

	if !ok {
		return nil
	}

	value := ast.Node(nil)

	if eq := p.consume(token.Eq); eq != nil {
		if value = p.parseExpr(); value == nil {
			return nil
		}
	} else if binding.Type == nil {
		p.errorExpected(p.tok.Start, p.tok.End, "value for binding with undefined type")
		return nil
	}

	return &ast.VarDecl{
		Binding: binding,
		Value:   value,
		Loc:     tok.Start,
	}
}

func (p *Parser) parseFuncDecl() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwFunc)
	name := p.parseIdentNode()

	if name == nil {
		return nil
	}

	signature := p.parseSignature(nil)

	if signature == nil {
		return nil
	}

	body := (*ast.CurlyList)(nil)

	if p.tok.Kind == token.LCurly {
		if list, ok := p.parseBlock().(*ast.CurlyList); ok {
			body = list
		} else {
			start, end := p.skipTo()
			p.errorExpected(start, end, "body")
			return nil
		}
	} else if p.tok.Kind != token.NewLine && p.tok.Kind != token.EOF {
		start, end := p.skipTo()
		p.errorExpectedToken(start, end, token.Arrow, token.LCurly, token.NewLine)
	}

	// if body == nil {
	// 	body = &ast.CurlyList{List: &ast.List{}}
	// }

	return &ast.FuncDecl{
		Loc:       tok.Start,
		Name:      name,
		Signature: signature,
		Body:      body,
	}
}

//------------------------------------------------
// IDK
//------------------------------------------------

func (p *Parser) parseAttributes() *ast.AttributeList {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tokenIdx := p.save()

	if tok := p.consume(token.At); tok != nil {
		if list := p.parseParenList(p.parseExpr); list != nil {
			return &ast.AttributeList{
				Loc:  tok.Start,
				List: list,
			}
		}
	}

	p.restore(tokenIdx)
	return nil
}

func (p *Parser) parseBracketExpr() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if list := p.parseBracketList(p.parseExpr); list != nil {
		return list
	}

	return nil
}

func (p *Parser) parseBindingAndValue() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	var binding *ast.Binding

	if b := p.parseBinding(); b != nil {
		binding = b.(*ast.Binding)
	} else {
		return nil
	}

	if opr := p.consume(token.Eq); opr != nil {
		p.consume(token.NewLine)
		value := p.parseExpr()

		if value == nil {
			start, end := p.skipTo(endOfStmtKinds...)
			p.errorExpected(start, end, "expression")
			return nil
		}

		return &ast.BindingWithValue{
			Binding: binding,
			Operator: &ast.Operator{
				Start: opr.Start,
				End:   opr.End,
				Kind:  ast.OperatorAssign,
			},
			Value: value,
		}
	}

	if binding.Type == nil {
		p.errorExpected(p.tok.Start, p.tok.End, "type for binding without value")
		return nil
	}

	return binding
}

func (p *Parser) parseBinding() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	name := p.parseIdentNode()

	if name == nil {
		return nil
	}

	typ := p.parseType()

	// if colon := p.consume(token.Colon); colon != nil {
	// 	typ = p.parseType()
	// 	if typ == nil {
	// 		return nil
	// 	}
	// }

	return &ast.Binding{
		Name: name,
		Type: typ,
	}
}

func (p *Parser) parseTypeName() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	switch p.tok.Kind {
	case token.Ident:
		switch expr := p.parseMemberAccess(p.parseIdentNode()); expr.(type) {
		case *ast.Ident:
			return expr

		default:
			start, end := p.skipTo()
			p.errorExpected(start, end, "type name")
			return nil
		}

	default:
		start, end := p.skipTo()
		p.error(start, end, "expected type name")
		return nil
	}
}

func (p *Parser) parseSignature(funcTok *token.Token) *ast.Signature {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	paramList := p.parseParenList(p.parseBindingAndValue)

	if paramList == nil {
		return nil
	}

	returnType := ast.Node(nil)

	if p.tok.Kind != token.Arrow && p.tok.Kind != token.LCurly {
		returnType = p.parseType()
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
		ampLoc := p.expect().Start

		if varTok := p.consume(token.KwVar); varTok != nil {
			return &ast.PrefixOp{
				X: p.parseType(),
				Opr: &ast.Operator{
					Start: ampLoc,
					End:   varTok.End,
					Kind:  ast.OperatorMutAddr,
				},
			}
		}

		return &ast.PrefixOp{
			X: p.parseType(),
			Opr: &ast.Operator{
				Start: ampLoc,
				End:   ampLoc,
				Kind:  ast.OperatorAddr,
			},
		}

	case token.LBracket:
		brackets := p.parseBracketList(p.parseExpr)

		return &ast.ArrayType{
			X:    p.parseType(),
			Args: brackets,
		}

	case token.LParen:
		if tuple := p.parseParenList(p.parseType); tuple != nil {
			return tuple
		}

		return nil

	case token.Ident:
		return p.parseTypeName()

	case token.KwFunc:
		return p.parseSignature(p.consume())

	case token.At:
		return p.parseBuiltIn()

	default:
		return nil
	}
}

/*
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

	declKind := ast.UnknownDecl

	switch tok.Kind {
	case token.KwVar:
		declKind = ast.VarDecl

	case token.KwVal:
		declKind = ast.ValDecl

	case token.KwConst:
		declKind = ast.ConstDecl

	default:
		panic("unreachable")
	}

	return &ast.GenericDecl{
		Field: field,
		Loc:   tok.Start,
		Kind:  declKind,
	}
} */

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
				p.error(start, end, "expected `if` clause or block after `else`")
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
		p.error(start, end, "expected conditional expression for `if` clause")
		return nil
	}

	body := p.parseCurlyList(p.parseStmt)

	if body == nil {
		start, end := p.skipTo()
		p.error(start, end, "expected body for 'if' clause")
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

func (p *Parser) parseBreakOrContinue() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	tok := p.expect(token.KwBreak, token.KwContinue)
	label := (*ast.Ident)(nil)

	if p.tok.Kind == token.Ident {
		label = p.parseIdentNode()
	}

	switch tok.Kind {
	case token.KwBreak:
		return &ast.Break{
			Label: label,
			Loc:   tok.Start,
		}

	case token.KwContinue:
		return &ast.Continue{
			Label: label,
			Loc:   tok.Start,
		}

	default:
		panic("unreachable")
	}
}

func (p *Parser) parseModule() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if tok := p.consume(token.KwModule); tok != nil {
		if name := p.parseIdentNode(); name != nil {
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

	p.error(p.tok.Start, p.tok.End, "first statement in the file should be `module`")
	return nil
}

func (p *Parser) parseAlias() ast.Node {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if tokAlias := p.consume(token.KwAlias); tokAlias != nil {
		id := p.parseIdentNode()

		if id == nil {
			return nil
		}

		p.expect(token.Eq)
		p.consume(token.NewLine)

		expr := p.parseType()

		if expr == nil {
			p.errorExpected(p.tok.Start, p.tok.End, "type")
			return nil
		}

		return &ast.TypeAliasDecl{
			Loc:  tokAlias.Start,
			Name: id,
			Expr: expr,
		}
	}

	return nil
}

//------------------------------------------------
// Lists
//------------------------------------------------

func (p *Parser) parseBracketList(f func() ast.Node) *ast.BracketList {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if exprs, openLoc, closeLoc := p.parseBracketedList(
		f,
		token.LBracket,
		token.RBracket,
		token.Comma,
	); exprs != nil {
		return &ast.BracketList{
			ExprList: &ast.ExprList{Exprs: exprs},
			Open:     openLoc,
			Close:    closeLoc,
		}
	}

	return nil
}

func (p *Parser) parseParenList(f func() ast.Node) *ast.ParenList {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if exprs, openLoc, closeLoc := p.parseBracketedList(
		f,
		token.LParen,
		token.RParen,
		token.Comma,
	); exprs != nil {
		return &ast.ParenList{
			ExprList: &ast.ExprList{Exprs: exprs},
			Open:     openLoc,
			Close:    closeLoc,
		}
	}

	return nil
}

func (p *Parser) parseCurlyList(f func() ast.Node) *ast.CurlyList {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if nodes, openLoc, closeLoc := p.parseBracketedList(
		f,
		token.LCurly,
		token.RCurly,
		token.Semicolon,
		token.NewLine,
	); nodes != nil {
		return &ast.CurlyList{
			List:  &ast.List{Nodes: nodes},
			Open:  openLoc,
			Close: closeLoc,
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

func (p *Parser) parseStmtList() *ast.List {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if nodes := p.listWithDelimiter(
		p.parseStmt,
		token.EOF,
		token.Semicolon,
		token.NewLine,
	); len(nodes) > 0 {
		return &ast.List{Nodes: nodes}
	}

	return nil
}

//------------------------------------------------
// Helper functions
//------------------------------------------------

func (p *Parser) listWithDelimiter(
	f func() ast.Node,
	delimiter token.Kind,
	separators ...token.Kind,
) []ast.Node {
	if len(separators) < 1 {
		panic("expect at least 1 separator")
	}

	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	nodes := []ast.Node{}

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
			return nil
		}

		nodeStart := p.tok.Start

		if node := f(); node != nil {
			if p.consume(separators...) != nil || p.match(delimiter) {
				// All is OK.
				nodes = append(nodes, node)
				continue
			}

			// [parseFunc] set the correct node, but no separator was found.
			// Report it and assign [ast.BadNode] instead.
			p.error(node.Pos(), node.LocEnd(), "unterminated expression")
		}

		// Something went wrong, advance to some delimiter and
		// continue parsing elements until we find the [closing] token.
		p.skipTo(append(separators, delimiter)...)
		p.consume(separators...)
		nodes = append(nodes, &ast.BadNode{Loc: nodeStart})
	}

	return nodes
}

func (p *Parser) parseBracketedList(
	f func() ast.Node,
	opening, closing token.Kind,
	separators ...token.Kind,
) (nodes []ast.Node, openLoc, closeLoc token.Loc) {
	if p.flags&Trace != 0 {
		p.trace()
		defer p.untrace()
	}

	if tok := p.expect(opening); tok != nil {
		openLoc = tok.Start
	} else {
		return nil, token.Loc{}, token.Loc{}
	}

	nodes = p.listWithDelimiter(f, closing, separators...)

	if nodes == nil {
		return nil, token.Loc{}, token.Loc{}
	}

	if tok := p.consume(closing); tok != nil {
		closeLoc = tok.Start
	} else {
		if p.tok.Kind == token.EOF {
			p.error(openLoc, openLoc, "bracket is never closed (end of file reached)")
		} else {
			start, end := p.skipTo()
			p.errorExpectedToken(start, end, append(separators, closing)...)
		}

		return nil, token.Loc{}, token.Loc{}
	}

	return nodes, openLoc, closeLoc
}
