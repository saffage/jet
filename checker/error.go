package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/token"
)

type Error struct {
	Message string
	Node    ast.Node
	Notes   []*Error // TODO make a distinct type for the notes.
}

func NewError(node ast.Node, message string) *Error {
	return &Error{
		Message: message,
		Node:    node,
	}
}

func NewErrorf(node ast.Node, format string, args ...any) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Node:    node,
	}
}

func (err *Error) Error() string { return err.Message }

func (err *Error) Report() {
	var start, end token.Loc
	if err.Node != nil {
		start, end = err.Node.Pos(), err.Node.LocEnd()
	}
	report.TaggedErrorAt("checker", start, end, err.Message)

	for _, note := range err.Notes {
		var start, end token.Loc
		if note.Node != nil {
			start, end = note.Node.Pos(), note.Node.LocEnd()
		}
		report.TaggedNoteAt("checker", start, end, note.Message)
	}
}

func (check *Checker) errorf(node ast.Node, format string, args ...any) {
	err := NewErrorf(node, format, args...)
	check.addError(err)
}

func (check *Checker) addError(err error) {
	check.errors = append(check.errors, err)
	check.isErrorHandled = false
}
