package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
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
	var span text.Span

	if e.Node != nil {
		span = text.Span{From: e.Node.Pos(), To: e.Node.PosEnd()}
	}

	var hints []report.Suggestion

	for _, hint := range e.Hints {
		var span text.Span

		if hint.Node != nil {
			span = text.Span{From: hint.Node.Pos(), To: hint.Node.PosEnd()}
		}

		hints = append(hints, report.Suggestion{
			Message:   hint.Message,
			Selection: report.Selection{Range: span},
		})
	}

	return &report.Info{
		Tag:         "checker",
		Title:       e.Message,
		Selection:   report.Selection{Range: span},
		Suggestions: hints,
	}
}

func (check *Checker) errorf(node ast.Node, format string, args ...any) {
	err := newErrorf(node, format, args...)
	check.addError(err)
}

func (check *Checker) addError(err error) {
	check.errors = append(check.errors, err)
}
