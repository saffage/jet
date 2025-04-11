package token

import (
	"errors"
	"unicode"
)

var (
	ErrFirstIsNotLetter      = errors.New("the first character must be an ASCII letter")
	ErrContainSpace          = errors.New("identifier cannot contain spaces")
	ErrContainPunct          = errors.New("identifier cannot contain a punctuation character")
	ErrUnsupportedUTF8       = errors.New("UTF-8 is not supported")
	ErrIllegalIdentCharacter = errors.New("illegal character in the identifier")
)

func IsIdentifierStartChar(char rune) bool {
	return char == '_' ||
		('a' <= char && char <= 'z') ||
		('A' <= char && char <= 'Z')
}

func IsIdentifierChar(char rune) bool {
	return IsIdentifierStartChar(char) || ('0' <= char && char <= '9')
}

func CheckIdent(s string) (errIndex int, err error) {
	if !IsIdentifierStartChar(rune(s[0])) {
		return 0, ErrFirstIsNotLetter
	}

	for i, char := range s {
		switch {
		case IsIdentifierChar(rune(s[i])):
			// OK

		case char > unicode.MaxASCII:
			return i, ErrUnsupportedUTF8

		default:
			return i, ErrIllegalIdentCharacter
		}
	}

	return 0, nil
}

func IsValidIdent(s string) bool {
	_, err := CheckIdent(s)
	return err == nil
}
