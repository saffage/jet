package scanner

import (
	"fmt"

	"github.com/saffage/jet/token"
)

type Error struct {
	Message    string
	Details    string
	Start, End token.Loc
}

func (e Error) Error() string {
	if e.Details == "" {
		return e.Message
	}

	return e.Message + "; " + e.Details
}

// Emits an error. Error end is a current scanner position.
func (s *Scanner) error(message string, start token.Loc, details ...any) {
	s.errors = append(s.errors, Error{
		Message: message,
		Details: fmt.Sprint(details...),
		Start:   start,
		End:     s.Pos(),
	})
}

func (s *Scanner) errorExpected(message string, pos token.Loc, details ...any) {
	message = "expected " + message
	s.error(message, pos, details...)
}

func (s *Scanner) errorUnexpected(message string, pos token.Loc, details ...any) {
	message = "unexpected " + message
	s.error(message, pos, details...)
}
