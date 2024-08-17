package token

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/galsondor/go-ascii"
)

var (
	ErrorFirstIsNotLetter = errors.New("the first character must be an ASCII letter")
	ErrorContainSpace     = errors.New("identifier cannot contain spaces")
	ErrorContainPunct     = errors.New("identifier cannot contain a punctuation character")
	ErrorUnsupportedUTF8  = errors.New("UTF-8 is not supported")
	ErrorIllegalCharacter = errors.New("illegal character in the identifier")
)

type Token struct {
	Data  string
	Start Pos
	End   Pos
	Kind  Kind
}

func (token Token) String() string {
	if len(token.Data) > 0 {
		return fmt.Sprintf("(%s %s at %s)",
			token.Kind,
			strconv.Quote(token.Data),
			token.Start,
		)
	}

	return fmt.Sprintf("(%s at %s)", token.Kind, token.Start)
}

func IsIdentStartChar(char byte) bool {
	return char == '_' || ascii.IsLetter(char)
}

func IsIdentChar(char byte) bool {
	return char == '_' || ascii.IsAlnum(char)
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
