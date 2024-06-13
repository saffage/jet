package scanner

import (
	"errors"
	"fmt"

	"github.com/saffage/jet/internal/report"
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
	Start   token.Pos
	End     token.Pos
	Message string

	err error
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

func (e Error) Is(err error) bool {
	return e.err == err
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
