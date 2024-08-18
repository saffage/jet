package parser

import (
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

func precedenceOf(kind token.Kind) int {
	switch kind {
	case token.Asterisk,
		token.Slash,
		token.Percent:
		return 10

	case token.Plus,
		token.Minus:
		return 9

	case token.Shl,
		token.Shr:
		return 8

	case token.Amp,
		token.Pipe,
		token.Caret:
		return 7

	case token.EqOp,
		token.NeOp,
		token.LtOp,
		token.GtOp,
		token.LeOp,
		token.GeOp:
		return 6

	case token.And:
		return 5

	case token.Or:
		return 4

	case token.KwAs:
		return 3

	case token.Dot2:
		return 2

	case token.Eq,
		token.PlusEq,
		token.MinusEq,
		token.AsteriskEq,
		token.SlashEq,
		token.PercentEq,
		token.AmpEq,
		token.PipeEq,
		token.CaretEq,
		token.ShlEq,
		token.ShrEq:
		return 1

	default:
		return 0
	}
}

func (p *parser) skipUntil(kinds ...token.Kind) (start, end token.Pos) {
	if len(kinds) == 0 {
		panic("must be at least 1 token")
	}

	start = p.tok.Start

	for p.tok.Kind != token.EOF && !slices.Contains(kinds, p.tok.Kind) {
		end = p.next().End
	}

	return start, end
}

var (
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
		token.And:        ast.OperatorAnd,
		token.Or:         ast.OperatorOr,
		token.KwAs:       ast.OperatorAs,
		token.Dot2:       ast.OperatorRangeInclusive,
	}

	literals = map[token.Kind]ast.LiteralKind{
		token.Int:    ast.IntLiteral,
		token.Float:  ast.FloatLiteral,
		token.String: ast.StringLiteral,
	}
)
