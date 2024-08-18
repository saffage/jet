package parser

import (
	"errors"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

type parseFunc func() (ast.Node, error)

func (parse *parser) decls() (ast.Node, error) {
	decls, err := parse.listDelimiter(parse.decl, token.EOF, token.Illegal)

	if err != nil {
		return nil, err
	}

	return &ast.StmtList{Nodes: decls}, nil
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

	if parse.matchAny(token.Name, token.Type, token.LParen) {
		if ty, err = parse.typeExpr(); err != nil {
			return nil, err
		}
	}

	return &ast.Decl{Name: name, Type: ty}, nil
}

func (parse *parser) labeledExpr(f parseFunc) parseFunc {
	return func() (ast.Node, error) {
		var (
			label *ast.Name
			expr  ast.Node
			err   error
		)

		if parse.matchSequence(token.Name, token.Colon) {
			name, _ := parse.name()
			label, _ = name.(*ast.Name)
			_ = parse.next()
		}

		if expr, err = f(); err != nil {
			return nil, err
		}

		if label != nil {
			return &ast.Label{Label: label, X: expr}, nil
		}

		return expr, nil
	}
}

func (parse *parser) typeVariant() (ast.Node, error) {
	var (
		ty     ast.Ident
		tyExpr ast.Node
		err    error
	)

	if ty, err = parse.ty(); err != nil {
		return nil, err
	}

	if parse.match(token.LParen) {
		if tyExpr, err = parse.parens(parse.labeledExpr(parse.typeExpr)); err != nil {
			return nil, err
		}
	}

	return &ast.Decl{Name: ty, Type: tyExpr}, nil
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

	if expr, err = parse.expr(); err != nil {
		return nil, err
	}

	return &ast.LetDecl{
		LetTok: letTok.Start,
		Decl:   decl.(*ast.Decl),
		Value:  expr,
	}, nil
}

func (parse *parser) typeDecl() (ast.Node, error) {
	var (
		typeTok token.Token
		name    ast.Ident
		args    *ast.ParenList
		expr    ast.Node
		err     error
	)

	if typeTok, err = parse.expect(token.KwType); err != nil {
		return nil, err
	}

	if name, err = parse.ty(); err != nil {
		return nil, err
	}

	if parse.tok.Kind == token.LParen {
		var parenList ast.Node

		if parenList, err = parse.parens(parse.variable); err != nil {
			return nil, err
		}

		args = parenList.(*ast.ParenList)
	}

	switch parse.tok.Kind {
	case token.Eq:
		parse.next()

		if expr, err = parse.typeExpr(); err != nil {
			return nil, err
		}

	case token.LCurly:
		if expr, err = parse.curlies(parse.typeVariant); err != nil {
			return nil, err
		}

	default:
		// External type.
	}

	return &ast.TypeDecl{
		TypeTok: typeTok.Start,
		Type:    name.(*ast.Type),
		Args:    args,
		Expr:    expr,
	}, nil
}

func (parse *parser) underscore() (ast.Ident, error) {
	tok, err := parse.expect(token.Underscore)

	if err != nil {
		return nil, err
	}

	return &ast.Underscore{
		Data:  tok.Data,
		Start: tok.Start,
		End:   tok.End,
	}, nil
}

func (parse *parser) name() (ast.Ident, error) {
	tok, err := parse.expect(token.Name)

	if err != nil {
		return nil, parse.error(ErrExpectedIdent)
	}

	return &ast.Name{Data: tok.Data, Start: tok.Start, End: tok.End}, nil
}

func (parse *parser) ty() (ast.Ident, error) {
	tok, err := parse.expect(token.Type)

	if err != nil {
		return nil, parse.error(ErrExpectedIdent)
	}

	return &ast.Type{Data: tok.Data, Start: tok.Start, End: tok.End}, nil
}

func (parse *parser) literal() (ast.Node, error) {
	tok, ok := parse.takeAny(token.Int, token.Float, token.String)

	if !ok {
		return nil, parse.error(ErrExpectedOperand)
	}

	return &ast.Literal{
		Kind:  literals[tok.Kind],
		Data:  tok.Data,
		Start: tok.Start,
		End:   tok.End,
	}, nil
}

func (parse *parser) ident() (ast.Ident, error) {
	switch parse.tok.Kind {
	case token.Name:
		return parse.name()

	case token.Underscore:
		return parse.underscore()

	default:
		return nil, parse.errorf(
			ErrUnexpectedToken,
			"expeted %s or %s",
			token.Name.UserString(),
			token.Underscore.UserString(),
		)
	}
}

func (parse *parser) block() (ast.Node, error) {
	return parse.curlies(parse.expr)
}

func (parse *parser) curlies(f parseFunc) (*ast.CurlyList, error) {
	if !parse.match(token.LCurly) {
		return nil, parse.error(ErrExpectedBlock)
	}

	list, err := parse.listOpenClose(f, token.LCurly, token.RCurly, 0)

	if err != nil {
		return nil, err
	}

	return &ast.CurlyList{
		StmtList: &ast.StmtList{Nodes: list.nodes},
		Open:     list.open,
		Close:    list.close,
	}, nil
}

func (parse *parser) parens(f parseFunc) (*ast.ParenList, error) {
	list, err := parse.listOpenClose(
		f,
		token.LParen,
		token.RParen,
		token.Comma,
	)

	if err != nil {
		return nil, err
	}

	return &ast.ParenList{
		List:  &ast.List{Nodes: list.nodes},
		Open:  list.open,
		Close: list.close,
	}, nil
}

func (parse *parser) brackets(f parseFunc) (*ast.BracketList, error) {
	list, err := parse.listOpenClose(
		f,
		token.LBracket,
		token.RBracket,
		token.Comma,
	)

	if err != nil {
		return nil, err
	}

	return &ast.BracketList{
		List:  &ast.List{Nodes: list.nodes},
		Open:  list.open,
		Close: list.close,
	}, nil
}

func (parse *parser) args() (ast.Node, error) {
	return parse.parens(parse.labeledExpr(parse.expr))
}

func (parse *parser) typeArgs() (ast.Node, error) {
	return parse.parens(parse.typeExpr)
}

func (parse *parser) typeExpr() (ast.Node, error) {
	var node ast.Node

	switch parse.tok.Kind {
	case token.Type:
		node, _ = parse.ty()

		if parse.match(token.LParen) {
			typeArgs, err := parse.typeArgs()

			if err != nil {
				return nil, err
			}

			node = &ast.Call{X: node, Args: typeArgs.(*ast.ParenList)}
		}

	case token.Name:
		tok := parse.next()
		node = &ast.Name{
			Data:  tok.Data,
			Start: tok.Start,
			End:   tok.End,
		}

	case token.LParen:
		params, err := parse.parens(parse.variable)

		if err != nil {
			return nil, err
		}

		var result ast.Node

		if parse.matchAny(token.Type, token.Name, token.LParen) {
			if result, err = parse.typeExpr(); err != nil {
				return nil, err
			}
		}

		node = &ast.Signature{Params: params, Result: result}

	default:
		return nil, parse.errorf(
			ErrUnexpectedToken,
			"expected %s or %s",
			token.Type.UserString(),
			token.Name.UserString(),
		)
	}

	return node, nil
}

func (parse *parser) expr() (ast.Node, error) {
	switch {
	case parse.match(token.KwWhen):
		return parse.whenExpr()

	default:
		return parse.binaryExpr(nil, 2)
	}
}

func (parse *parser) whenExpr() (ast.Node, error) {
	var (
		whenTok token.Token
		expr    ast.Node
		body    ast.Node
		err     error
	)

	if whenTok, err = parse.expect(token.KwWhen); err != nil {
		return nil, err
	}

	if expr, err = parse.expr(); err != nil {
		return nil, err
	}

	if body, err = parse.curlies(parse.whenCase); err != nil {
		return nil, err
	}

	return &ast.When{
		TokPos: whenTok.Start,
		Expr:   expr,
		Body:   body.(*ast.CurlyList),
	}, nil
}

func (parse *parser) whenCase() (ast.Node, error) {
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
		X:     pattern,
		Y:     expr,
		Start: arrowTok.Start,
		End:   arrowTok.End,
		Kind:  ast.OperatorFatArrow,
	}, nil
}

func (parse *parser) pattern() (ast.Node, error) {
	switch {
	case parse.match(token.Name):
		return parse.name()

	case parse.match(token.Underscore):
		return parse.underscore()

	case parse.match(token.Type):
		ty, _ := parse.ty()

		if parse.match(token.LParen) {
			if list, err := parse.parens(parse.pattern); err != nil {
				return nil, err
			} else {
				return &ast.Call{X: ty, Args: list}, nil
			}
		}

		return ty, nil

	default:
		return parse.expr()
	}
}

func (parse *parser) binaryExpr(x ast.Node, precedence int) (ast.Node, error) {
	var err error

	if x == nil {
		if x, err = parse.prefix(); err != nil {
			return nil, err
		}
	}

	for oprKind, ok := operators[parse.tok.Kind]; ok &&
		precedenceOf(parse.tok.Kind) >= precedence; {

		oprTok := parse.next()
		y, err := parse.binaryExpr(nil, precedenceOf(oprTok.Kind)+1)

		if err != nil {
			return nil, err
		}

		x = &ast.Op{
			X:     x,
			Y:     y,
			Start: oprTok.Start,
			End:   oprTok.End,
			Kind:  oprKind,
		}
	}

	return x, nil
}

func (parse *parser) prefix() (ast.Node, error) {
	return parse.primary(nil)
}

func (parse *parser) primary(x ast.Node) (ast.Node, error) {
	var err error

	if x == nil {
		if x, err = parse.operand(); err != nil {
			return nil, err
		}
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

			x = &ast.Call{X: x, Args: args.(*ast.ParenList)}

		default:
			return x, nil
		}
	}
}

func (parse *parser) operand() (ast.Node, error) {
	switch parse.tok.Kind {
	case token.Name:
		return parse.name()

	case token.Type:
		return parse.ty()

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

	label, err := parse.ident()

	if err != nil {
		return nil, parse.error(ErrExpectedIdent)
	}

	return &ast.Dot{
		X:      x,
		Y:      label.(*ast.Name),
		DotPos: dotTok.Start,
	}, nil
}

///
///
///

type list struct {
	nodes []ast.Node
	open  token.Pos
	close token.Pos
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
			return nil, newError(err, openTok.Start, openTok.End)
		}

		return nil, err
	}

	if _, ok = parse.take(closing); !ok {
		panic("unreachable")
	}

	return &list{nodes, openTok.Start, openTok.End}, nil
}

func (parse *parser) listDelimiter(
	f parseFunc,
	delim, sep token.Kind,
) ([]ast.Node, error) {
	nodes := []ast.Node{}
	errs := ([]error)(nil)

	// list(f) = f {separator f} [separator]
	for parse.tok.Kind != delim {
		// Possible cases:
		//  - empty list `{}`
		//  - trailing separator `{x,}`
		//  - unterminated list `{`
		if parse.tok.Kind == token.EOF {
			return nil, ErrUnterminatedList
		}

		nodeStart := parse.tok.Start
		node, err := f()

		if err == nil {
			if sep == 0 || parse.consume(sep) || parse.match(delim) {
				nodes = append(nodes, node)
				continue
			}

			// The node is correct, but no separator\delimiter
			// was found. Report it and assign [ast.BadNode] instead.
			err = newError(ErrUnterminatedExpr, node.Pos(), node.PosEnd())
		}

		// Something went wrong, advance to some delimiter and
		// continue parsing elements until we find the 'closing' token.
		parse.skipUntil(sep, delim)
		parse.take(sep)
		errs = append(errs, err)
		nodes = append(nodes, &ast.BadNode{DesiredPos: nodeStart})
	}

	return nodes, errors.Join(errs...)
}
