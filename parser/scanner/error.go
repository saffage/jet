package scanner

import (
	"errors"
	"fmt"

	"github.com/saffage/jet/parser/token"
	"github.com/saffage/jet/report"
)

var (
	ErrIllegalCharacter        = errors.New("illegal character")
	ErrIllegalNumericBase      = errors.New("uppercase letters in numeric base prefix is not allowed, use lowercase letter instead")
	ErrInvalidByte             = errors.New("invalid byte")
	ErrInvalidEscape           = errors.New("invalid character escape")
	ErrUnterminatedStringLit   = errors.New("unterminated string literal")
	ErrFirstDigitIsZero        = errors.New("'0' as the first digit of decimal number literal is not allowed")
	ErrExpectedIdentForSuffix  = errors.New("expected identifier for numeric suffix")
	ErrExpectedDigitAfterPoint = errors.New("expected digit after the point")
	ErrExpectedDecNumber       = errors.New("expected decimal number")
	ErrExpectedHexNumber       = errors.New("expected hexadecimal number")
	ErrExpectedBinNumber       = errors.New("expected binary number")
	ErrExpectedOctNumber       = errors.New("expected octal number")
)

type Error struct {
	err error

	Message string
	Range   token.Range
}

func (e Error) Error() string {
	if e.Message != "" {
		return e.err.Error() + ": " + e.Message
	}
	return e.err.Error()
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) Info() *report.Info {
	if p, ok := e.err.(report.Problem); ok {
		return p.Info()
	}

	info := &report.Info{
		Tag:   "scanner",
		Title: e.err.Error(),
		Range: e.Range,
	}

	if e.Message != "" {
		info.Title += ": " + e.Message
	}

	return info
}

// Emits an error. Error end is a current scanner position.
func (s *Scanner) error(err error, start token.Pos, message ...any) {
	s.errors = append(s.errors, Error{
		Message: fmt.Sprint(message...),
		Range:   token.RangeFrom(start, s.Pos()),
		err:     err,
	})
}
