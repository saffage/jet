package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

func (p *Parser) parseIdentNode() *ast.Ident {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if tok := p.consume(token.Ident); tok != nil {
		return &ast.Ident{
			Name:  tok.Data,
			Start: tok.Start,
			End:   tok.End,
		}
	}

	return nil
}

func (p *Parser) parseLiteralNode() *ast.Literal {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if tok := p.consume(token.Int, token.Float, token.String); tok != nil {
		var litKind ast.LiteralKind

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

func (p *Parser) parseDeclName() (mut token.Loc, name *ast.Ident) {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	if tokMut := p.consume(token.KwMut); tokMut != nil {
		mut = tokMut.Start
	}

	if ident := p.parseIdentNode(); ident != nil {
		return mut, ident
	}

	if mut.IsValid() {
		p.error(ErrorExpectedIdentAfterMut)
	} else {
		p.errorExpectedToken(token.Ident)
	}

	return token.Loc{}, nil
}

func (p *Parser) parseTypeName() ast.Node {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	ident := p.parseIdentNode()
	if ident == nil {
		p.error(ErrorExpectedTypeName)
		return nil
	}

	path := p.parseDot(ident)
	if path == nil {
		return nil
	}

	if p.tok.Kind == token.LBracket {
		brackets := p.parseBracketList(p.parseExpr)
		if brackets == nil {
			return nil
		}

		return &ast.Index{
			X:    path,
			Args: brackets,
		}
	}

	return path
}

func (p *Parser) isExprStart(kind token.Kind) bool {
	switch kind {
	case token.Ident,
		token.Dollar,
		token.Int,
		token.Float,
		token.String,
		token.LParen,
		token.LCurly,
		token.KwStruct,
		token.KwIf,
		token.KwWhile,
		token.KwReturn,
		token.KwBreak,
		token.KwContinue:
		return true

	default:
		return false
	}
}
