package scanner

import (
	"errors"
	"fmt"

	"github.com/saffage/jet/report"
	"github.com/saffage/jet/token"
)

var (
	ErrorIllegalCharacter        = errors.New("illegal character")
	ErrorIllegalNumericBase      = errors.New("uppercase letters in numeric base prefix is not allowed, use lowercase letter instead")
	ErrorInvalidByte             = errors.New("invalid byte")
	ErrorInvalidEscape           = errors.New("invalid character escape")
	ErrorUnterminatedStringLit   = errors.New("unterminated string literal")
	ErrorFirstDigitIsZero        = errors.New("'0' as the first digit of a number literal is not allowed")
	ErrorExpectedIdentForSuffix  = errors.New("expected identifier for numeric suffix")
	ErrorExpectedDigitAfterPoint = errors.New("expected digit after the point")
	ErrorExpectedDecNumber       = errors.New("expected decimal number")
	ErrorExpectedHexNumber       = errors.New("expected hexadecimal number")
	ErrorExpectedBinNumber       = errors.New("expected binary number")
	ErrorExpectedOctNumber       = errors.New("expected octal number")
)

type Error struct {
	err error

	Message   string
	Selection token.Range
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Is(target error) bool {
	return e.err == target
}

func (e *Error) Info() *report.Info {
	return &report.Info{
		Tag:            "scanner",
		Title:          e.Error(),
		SelectionRange: e.Selection,
	}
}

// Emits an error. Error end is a current scanner position.
func (s *Scanner) error(err error, start token.Pos, message ...any) {
	s.errors = append(s.errors, &Error{
		Message:   fmt.Sprint(message...),
		Selection: start.WithEnd(s.Pos()),
		err:       err,
	})
}
