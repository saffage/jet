package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/token"
)

type Error struct {
	err error

	Message string
	Node    ast.Node
	Hints   []*Error // TODO make a distinct type for the notes.
}

func newErrorf(node ast.Node, format string, args ...any) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Node:    node,
	}
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Is(target error) bool {
	return e.err == target
}

func (e *Error) Info() *report.Info {
	var span token.Range

	if e.Node != nil {
		span = e.Node.Pos().WithEnd(e.Node.PosEnd())
	}

	var hints []report.HintInfo

	for _, hint := range e.Hints {
		var span token.Range

		if hint.Node != nil {
			span = hint.Node.Pos().WithEnd(hint.Node.PosEnd())
		}

		hints = append(hints, report.HintInfo{
			Message:         hint.Message,
			SuggestionRange: span,
		})
	}

	return &report.Info{
		Tag:            "checker",
		Title:          e.Message,
		SelectionRange: span,
		Hints:          hints,
	}
}

func (check *Checker) errorf(node ast.Node, format string, args ...any) {
	err := newErrorf(node, format, args...)
	check.addError(err)
}

func (check *Checker) addError(err error) {
	check.errors = append(check.errors, err)
}
