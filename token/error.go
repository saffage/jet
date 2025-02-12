package token

import (
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

type Error struct {
	err error

	Message   string
	Selection text.Span
}

func (e *Error) Error() string        { return e.Message }
func (e *Error) Is(target error) bool { return e.err == target }

func (e *Error) Info() *report.Info {
	return &report.Info{
		Tag:       "scanner",
		Title:     e.Error(),
		Selection: report.Selection{Range: e.Selection},
	}
}
