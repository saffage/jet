package token

import (
	"errors"
	"unicode/utf8"

	"github.com/galsondor/go-ascii"
	"github.com/saffage/jet/text"
)

var (
	ErrorFirstIsNotLetter = errors.New("the first character must be an ASCII letter")
	ErrorContainSpace     = errors.New("identifier cannot contain spaces")
	ErrorContainPunct     = errors.New("identifier cannot contain a punctuation character")
	ErrorUnsupportedUTF8  = errors.New("UTF-8 is not supported")
	ErrorIllegalCharacter = errors.New("illegal character in the identifier")
)

type Token struct {
	Kind Kind
	Data string
	Span text.Span
}

func (t Token) Precedence() int {
	switch t.Kind {
	case Asterisk, Slash, Percent:
		return 10

	case Plus, Minus:
		return 9

	case Shl, Shr:
		return 8

	case Amp, Pipe, Caret:
		return 7

	case EqOp, NeOp, LtOp, GtOp, LeOp, GeOp:
		return 6

	case KwAnd:
		return 5

	case KwOr:
		return 4

	case KwAs:
		return 3

	case Dot2, Dot2Less:
		return 2

		// TODO maybe add 'Arrow' & 'FatArrow' operators

	case Eq, PlusEq, MinusEq, AsteriskEq, SlashEq, PercentEq, AmpEq, PipeEq, CaretEq, ShlEq, ShrEq:
		return 1

	default:
		return 0
	}
}

func IsIdentifierStartChar(char byte) bool {
	return char == '_' || ascii.IsLetter(char)
}

func IsIdentifierChar(char byte) bool {
	return IsIdentifierStartChar(char) || ascii.IsDigit(char)
}

func IsValidIdent(s string) (int, error) {
	if !ascii.IsLetter(s[0]) {
		return 0, ErrorFirstIsNotLetter
	}

	for i := 1; i < len(s); i++ {
		switch {
		case ascii.IsLetter(s[i]), ascii.IsDigit(s[i]), s[i] == '_':
			// OK

		case ascii.IsSpace(s[i]):
			return i, ErrorContainSpace

		case ascii.IsPunct(s[i]):
			return i, ErrorContainPunct

		case utf8.RuneStart(s[i]):
			return i, ErrorUnsupportedUTF8

		default:
			return i, ErrorIllegalCharacter
		}
	}

	return 0, nil
}
