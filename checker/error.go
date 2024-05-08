package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
)

type Error struct {
	Message string
	Node    ast.Node
	Notes   []*Error
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

func (check *Checker) errorf(node ast.Node, format string, args ...any) {
	err := NewErrorf(node, format, args...)
	check.addError(err)
}

func (check *Checker) addError(err error) {
	check.errors = append(check.errors, err)
	check.isErrorHandled = false
}
