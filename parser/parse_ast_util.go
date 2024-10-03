package parser

import (
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser/token"
)

func (p *parser) skipUntil(kinds ...token.Kind) token.Range {
	if len(kinds) == 0 {
		panic("must be at least 1 token")
	}

	rng := p.tok.Range

	for p.tok.Kind != token.EOF && !slices.Contains(kinds, p.tok.Kind) {
		rng.End = p.next().End
	}

	return rng
}

var precedences = map[token.Kind]int{
	token.Asterisk:   10,
	token.Slash:      10,
	token.Percent:    10,
	token.Plus:       9,
	token.Minus:      9,
	token.Shl:        8,
	token.Shr:        8,
	token.Amp:        7,
	token.Pipe:       7,
	token.Caret:      7,
	token.EqOp:       6,
	token.NeOp:       6,
	token.LtOp:       6,
	token.GtOp:       6,
	token.LeOp:       6,
	token.GeOp:       6,
	token.And:        5,
	token.Or:         4,
	token.KwAs:       3,
	token.Dot2:       2,
	token.Eq:         1,
	token.PlusEq:     1,
	token.MinusEq:    1,
	token.AsteriskEq: 1,
	token.SlashEq:    1,
	token.PercentEq:  1,
	token.AmpEq:      1,
	token.PipeEq:     1,
	token.CaretEq:    1,
	token.ShlEq:      1,
	token.ShrEq:      1,
}

var operators = map[token.Kind]ast.OperatorKind{
	token.Plus:       ast.OperatorAdd,
	token.Minus:      ast.OperatorSub,
	token.Asterisk:   ast.OperatorMul,
	token.Slash:      ast.OperatorDiv,
	token.Percent:    ast.OperatorMod,
	token.Eq:         ast.OperatorAssign,
	token.PlusEq:     ast.OperatorAddAssign,
	token.MinusEq:    ast.OperatorSubAssign,
	token.AsteriskEq: ast.OperatorMulAssign,
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

var literals = map[token.Kind]ast.LiteralKind{
	token.Int:    ast.IntLiteral,
	token.Float:  ast.FloatLiteral,
	token.String: ast.StringLiteral,
}
