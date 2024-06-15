package parser

import (
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

func (p *parser) parseIdentNode() *ast.Ident {
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

func (p *parser) parseLiteralNode() *ast.Literal {
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

func (p *parser) skip(to ...token.Kind) (start, end token.Pos) {
	if len(to) == 0 {
		to = endOfExprKinds
	}

	start = p.tok.Start

	for p.tok.Kind != token.EOF && !slices.Contains(to, p.tok.Kind) {
		end = p.tok.End
		p.next()
	}

	if p.flags&Trace != 0 && end.IsValid() {
		// TODO must be removed
		warn := Error{
			Start:      start,
			End:        end,
			Message:    "tokens was skipped for some reason",
			isWarn:     true,
			isInternal: true,
		}
		p.errors = append(p.errors, warn)
	}

	return
}

var (
	endOfStmtKinds = []token.Kind{
		token.Semicolon,
		token.NewLine,
	}

	endOfExprKinds = append(endOfStmtKinds, []token.Kind{
		token.Comma,
		token.RParen,
		token.RCurly,
		token.RBracket,
	}...)

	simpleExprStartKinds = []token.Kind{
		token.Minus,
		token.Bang,
		token.Asterisk,
		token.Amp,
		token.Ident,
		token.Int,
		token.Float,
		token.String,
		token.Dollar,
		token.KwIf,
		token.KwWhile,
		token.KwFor,
		token.KwStruct,
		token.KwEnum,
		token.LCurly,
		token.LBracket,
		token.LParen,
	}

	exprStartKinds = append(simpleExprStartKinds, []token.Kind{
		token.KwDefer,
		token.KwReturn,
		token.KwBreak,
		token.KwContinue,
	}...)

	operators = map[token.Kind]ast.OperatorKind{
		token.Plus:       ast.OperatorAdd,
		token.Minus:      ast.OperatorSub,
		token.Asterisk:   ast.OperatorMul,
		token.Slash:      ast.OperatorDiv,
		token.Percent:    ast.OperatorMod,
		token.Eq:         ast.OperatorAssign,
		token.PlusEq:     ast.OperatorAddAssign,
		token.MinusEq:    ast.OperatorSubAssign,
		token.AsteriskEq: ast.OperatorMultAssign,
		token.SlashEq:    ast.OperatorDivAssign,
		token.PercentEq:  ast.OperatorModAssign,
		token.EqOp:       ast.OperatorEq,
		token.NeOp:       ast.OperatorNe,
		token.LtOp:       ast.OperatorLt,
		token.LeOp:       ast.OperatorLe,
		token.GtOp:       ast.OperatorGt,
		token.GeOp:       ast.OperatorGe,
		token.Amp:        ast.OperatorBitAnd,
		token.Pipe:       ast.OperatorBitOr,
		token.Caret:      ast.OperatorBitXor,
		token.Shl:        ast.OperatorBitShl,
		token.Shr:        ast.OperatorBitShr,
		token.KwAnd:      ast.OperatorAnd,
		token.KwOr:       ast.OperatorOr,
		token.KwAs:       ast.OperatorAs,
		token.Dot2:       ast.OperatorRangeInclusive,
		token.Dot2Less:   ast.OperatorRangeExclusive,
	}
)
