package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

func (parse *parser) decls() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if decls := parse.listDelimiter(parse.decl, token.EOF, 0); decls != nil {
		return &ast.StmtList{Nodes: decls}
	}

	return nil
}

func (parse *parser) decl() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	switch parse.tok.Kind {
	case token.KwLet:
		return parse.letDecl()

	default:
		parse.error(ErrExpectedDecl)
		return nil
	}
}

func (parse *parser) letDecl() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if letTok := parse.expect(token.KwLet); letTok != nil {
		if name := parse.name(); name != nil {
			t := parse.typeMaybe()

			if eq := parse.expect(token.Eq); eq != nil {
				if expr := parse.expr(); expr != nil {
					return &ast.Decl{
						Ident: name.(*ast.Ident),
						Type:  t,
						Value: expr,
					}
				}
			}
		}
	}

	return nil
}

func (parse *parser) underscore() *ast.Underscore {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if tok := parse.expect(token.Underscore); tok != nil {
		return &ast.Underscore{
			Name:  tok.Data,
			Start: tok.Start,
			End:   tok.End,
		}
	}

	return nil
}

func (parse *parser) ident() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if tok := parse.expect(token.Ident); tok != nil {
		return &ast.Ident{
			Name:  tok.Data,
			Start: tok.Start,
			End:   tok.End,
		}
	}

	return nil
}

func (parse *parser) literal() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	var lit ast.Node

	if kind, ok := literals[parse.tok.Kind]; ok {
		lit = &ast.Literal{
			Kind:  kind,
			Value: parse.tok.Data,
			Start: parse.tok.Start,
			End:   parse.tok.End,
		}
		parse.next()
	}

	parse.error(ErrExpectedOperand)
	return lit
}

func (parse *parser) name() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	var node ast.Node

	switch parse.tok.Kind {
	case token.Ident:
		node = &ast.Ident{
			Name:  parse.tok.Data,
			Start: parse.tok.Start,
			End:   parse.tok.End,
		}

	case token.Underscore:
		node = &ast.Underscore{
			Name:  parse.tok.Data,
			Start: parse.tok.Start,
			End:   parse.tok.End,
		}

	default:
		parse.error(ErrExpectedIdent)
		return nil
	}

	parse.next()
	return node
}

func (parse *parser) typeMaybe() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	var t ast.Node

	if tok := parse.consume(token.Type); tok != nil {
		t = &ast.Type{
			Name:  tok.Data,
			Start: tok.Start,
			End:   tok.End,
		}

		if parse.match(token.LParen) {
			if args := parse.typeArgs(); args != nil {
				t = &ast.Call{X: t, Args: args.(*ast.ParenList)}
			}
		}
	}

	return t
}

func (parse *parser) t() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	return nil
}

func (parse *parser) block() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	return parse.blockFunc(parse.expr)
}

func (parse *parser) blockFunc(f func() ast.Node) *ast.CurlyList {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if !parse.match(token.LCurly) {
		parse.error(ErrExpectedBlock)
		return nil
	}

	if nodes, openPos, closePos := parse.listOpenClose(
		f,
		token.LCurly,
		token.RCurly,
		0,
	); nodes != nil {
		return &ast.CurlyList{
			StmtList: &ast.StmtList{Nodes: nodes},
			Open:     openPos,
			Close:    closePos,
		}
	}

	return nil
}

func (parse *parser) args() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if args, openPos, closePos := parse.listOpenClose(
		parse.expr,
		token.LParen,
		token.RParen,
		token.Comma,
	); args != nil {
		return &ast.ParenList{
			List:  &ast.List{Nodes: args},
			Open:  openPos,
			Close: closePos,
		}
	}

	return nil
}

func (parse *parser) typeArg() (node ast.Node) {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	switch parse.tok.Kind {
	case token.Type:
		node = &ast.Type{
			Name:  parse.tok.Data,
			Start: parse.tok.Start,
			End:   parse.tok.End,
		}
		parse.next()

		if parse.match(token.LParen) {
			if typeArgs := parse.typeArgs(); typeArgs != nil {
				node = &ast.Call{X: node, Args: typeArgs.(*ast.ParenList)}
			}
		}

	case token.Ident:
		node = &ast.Ident{
			Name:  parse.tok.Data,
			Start: parse.tok.Start,
			End:   parse.tok.End,
		}
		parse.next()
	}

	return
}

func (parse *parser) typeArgs() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if args, openPos, closePos := parse.listOpenClose(
		parse.typeArg,
		token.LParen,
		token.RParen,
		token.Comma,
	); args != nil {
		return &ast.ParenList{
			List:  &ast.List{Nodes: args},
			Open:  openPos,
			Close: closePos,
		}
	}

	return nil
}

func (parse *parser) expr() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	return parse.binaryExpr(nil, 2)
}

func (parse *parser) binaryExpr(x ast.Node, precedence int) ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if x == nil {
		if x = parse.prefix(); x == nil {
			return nil
		}
	}

	for precedenceOf(&parse.tok) >= precedence {
		tok := parse.consumeAny()
		y := parse.binaryExpr(nil, precedenceOf(tok)+1)

		if y == nil {
			return nil
		}

		opKind, ok := operators[tok.Kind]

		if !ok {
			parse.errorfAt(
				ErrUnexpectedOperator,
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
			Kind:  opKind,
		}
	}

	return x
}

func (parse *parser) prefix() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	return parse.primary(nil)
}

func (parse *parser) primary(x ast.Node) ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if x == nil {
		if x = parse.operand(); x == nil {
			return nil
		}
	}

	for {
		if x == nil {
			return nil
		}

		switch parse.tok.Kind {
		case token.Dot:
			x = parse.dotExpr(x)

		case token.LParen:
			if args := parse.args(); args != nil {
				x = &ast.Call{X: x, Args: args.(*ast.ParenList)}
			} else {
				return nil
			}

		default:
			return x
		}
	}
}

func (parse *parser) operand() ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	switch parse.tok.Kind {
	case token.Ident:
		return parse.ident()

	case token.Int, token.Float, token.String:
		return parse.literal()

	case token.LCurly:
		return parse.block()

	case token.LBracket:
		panic("todo")
	}

	parse.error(ErrExpectedOperand)
	return nil
}

///
///
///

func (parse *parser) dotExpr(x ast.Node) ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	if dot := parse.consume(token.Dot); dot != nil {
		if y := parse.ident(); y != nil {
			return &ast.Dot{
				X:      x,
				Y:      y.(*ast.Ident),
				DotPos: dot.Start,
			}
		}
	}

	return nil
}

///
///
///

func (parse *parser) listDelimiter(
	f func() ast.Node,
	delimiter, separator token.Kind,
) []ast.Node {
	if parse.flags&Trace != 0 {
		defer un(trace(parse))
	}

	nodes := []ast.Node{}

	// list(f) = f {separator f} [separator]
	for {
		// Possible cases:
		//  - empty list `{}`
		//  - trailing separator `{x,}`
		//  - unterminated list `{`
		if parse.tok.Kind == delimiter {
			break
		}

		if parse.tok.Kind == token.EOF {
			return nil
		}

		nodeStart := parse.tok.Start

		if node := f(); node != nil {
			if separator != 0 && parse.drop(separator) ||
				parse.match(delimiter) {
				nodes = append(nodes, node)
				continue
			}
			// The node is correct, but no separator\delimiter
			// was found. Report it and assign [ast.BadNode] instead.
			parse.errorAt(ErrUnterminatedExpr, node.Pos(), node.PosEnd())
		}

		// Something went wrong, advance to some delimiter and
		// continue parsing elements until we find the 'closing' token.
		parse.skip(separator, delimiter)
		parse.consume(separator)
		nodes = append(nodes, &ast.BadNode{DesiredPos: nodeStart})
	}

	return nodes
}

func (p *parser) listOpenClose(
	f func() ast.Node,
	opening, closing, separator token.Kind,
) (nodes []ast.Node, openPos, closePos token.Pos) {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if openTok := p.expect(opening); openTok != nil {
		if nodes := p.listDelimiter(f, closing, separator); nodes != nil {
			if closeTok := p.consume(closing); closeTok != nil {
				return nodes, openTok.Start, closeTok.Start
			}

			if p.tok.Kind == token.EOF {
				p.errorAt(ErrBracketIsNeverClosed, openTok.Start, openTok.End)
			} else {
				start, end := p.skip()
				p.errorExpectedTokenAt(start, end, separator, closing)
			}
		}
	}

	return
}
