package token

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/galsondor/go-ascii"
)

var (
	ErrFirstIsNotLetter = errors.New("the first character must be an ASCII letter")
	ErrContainSpace     = errors.New("identifier cannot contain spaces")
	ErrContainPunct     = errors.New("identifier cannot contain a punctuation character")
	ErrUnsupportedUTF8  = errors.New("UTF-8 is not supported")
	ErrIllegalCharacter = errors.New("illegal character in the identifier")
)

type Token struct {
	Data string
	Kind Kind
	Range
}

func (token Token) String() string {
	if len(token.Data) > 0 {
		return fmt.Sprintf("(%s %s at %s)",
			token.Kind,
			strconv.Quote(token.Data),
			token.StartPos(),
		)
	}

	return fmt.Sprintf("(%s at %s)", token.Kind, token.StartPos())
}

func IsIdentStartChar(char byte) bool {
	return char == '_' || ascii.IsLetter(char)
}

func IsIdentChar(char byte) bool {
	return char == '_' || ascii.IsAlnum(char)
}

func IsValidIdent(s string) (int, error) {
	if !ascii.IsLetter(s[0]) {
		return 0, ErrFirstIsNotLetter
	}

	for i := 1; i < len(s); i++ {
		switch {
		case ascii.IsLetter(s[i]), ascii.IsDigit(s[i]), s[i] == '_':
			// OK

		case ascii.IsSpace(s[i]):
			return i, ErrContainSpace

		case ascii.IsPunct(s[i]):
			return i, ErrContainPunct

		case utf8.RuneStart(s[i]):
			return i, ErrUnsupportedUTF8

		default:
			return i, ErrIllegalCharacter
		}
	}

	return 0, nil
}
