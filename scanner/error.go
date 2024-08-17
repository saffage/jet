package scanner

import (
	"errors"
	"fmt"

	"github.com/saffage/jet/report"
	"github.com/saffage/jet/token"
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
	Start   token.Pos
	End     token.Pos
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

func (e Error) Report() {
	err, ok := e.err.(report.Reporter)
	if ok && err != nil {
		err.Report()
	}
	if !ok || e.Message != "" {
		message := e.err.Error()
		if e.Message != "" {
			message += ": " + e.Message
		}
		report.TaggedErrorAt("scanner", e.Start, e.End, message)
	}
}

// Emits an error. Error end is a current scanner position.
func (s *Scanner) error(err error, start token.Pos, message ...any) {
	s.errors = append(s.errors, Error{
		Message: fmt.Sprint(message...),
		Start:   start,
		End:     s.Pos(),
		err:     err,
	})
}
