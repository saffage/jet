package types

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

//-----------------------------------------------
// Internal error

type internalError struct {
	err  error
	node ast.Node
}

func internalErrorf(node ast.Node, format string, args ...any) *internalError {
	return &internalError{
		err:  fmt.Errorf(format, args...),
		node: node,
	}
}

func (err *internalError) Error() string {
	if err.node.Range().IsValid() {
		return fmt.Sprintf("%s: %s", err.node.Range(), err.err.Error())
	}

	return err.err.Error()
}

func (err *internalError) Unwrap() error {
	return err.err
}

func (err *internalError) Info() *report.Info {
	info := &report.Info{Title: err.err.Error()}

	if err.node != nil && err.node.Range().IsValid() {
		info.Range = err.node.Range()
	}

	return info
}

//-----------------------------------------------
// Checker

func (check *checker) internalErrorf(node ast.Node, format string, args ...any) {
	check.error(internalErrorf(node, format, args...))
}

func (check *checker) error(err error) {
	if err != nil {
		if errs, _ := err.(interface{ Unwrap() []error }); errs != nil {
			check.problems = append(check.problems, errs.Unwrap()...)
		} else {
			check.problems = append(check.problems, err)
		}
	}
}
