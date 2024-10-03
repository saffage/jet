package types

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

//-----------------------------------------------
// Problem

type Problem struct {
	err     error
	node    ast.Node
	hint    string
	notes   []note
	warning bool
}

func errorf(node ast.Node, format string, args ...any) *Problem {
	return &Problem{
		err:  fmt.Errorf(format, args...),
		node: node,
	}
}

func warningf(node ast.Node, format string, args ...any) *Problem {
	return &Problem{
		err:     fmt.Errorf(format, args...),
		node:    node,
		warning: true,
	}
}

func problem(p *Problem, notes ...note) *Problem {
	if p != nil {
		p.notes = append(p.notes, notes...)
		return p
	}
	return nil
}

func (err *Problem) Error() string {
	if err.node.Range().IsValid() {
		return fmt.Sprintf("%s: %s", err.node.Range(), err.err.Error())
	}
	return err.err.Error()
}

func (err *Problem) Info() *report.Info {
	info := &report.Info{
		Title: err.err.Error(),
		Hint:  err.hint,
	}

	if err.node != nil && err.node.Range().IsValid() {
		info.Range = err.node.Range()
	}

	if err.warning {
		info.Kind = report.KindWarning
	}

	for _, note := range err.notes {
		description := report.Description{Description: note.Message}

		if note.Node != nil {
			if note.Node.Range().IsValid() {
				description.Range = note.Node.Range()
			} else {
				description.Ast = note.Node
			}
		}

		info.Descriptions = append(info.Descriptions, description)
	}

	return info
}

//-----------------------------------------------
// Note

type note struct {
	Node    ast.Node
	Message string
}

func notef(node ast.Node, format string, args ...any) note {
	return note{
		Node:    node,
		Message: fmt.Sprintf(format, args...),
	}
}

func (note *note) Info() report.Info {
	panic("todo")
	// var rng token.Range

	// if note.Node != nil {
	// 	rng = note.Node.Range()
	// }

	// report.TaggedNoteAt("checker", rng, note.Message)
}

//-----------------------------------------------
// Checker

func (check *checker) warningf(node ast.Node, format string, args ...any) {
	err := warningf(node, format, args...)
	check.problem(err)
}

func (check *checker) errorf(node ast.Node, format string, args ...any) {
	err := errorf(node, format, args...)
	check.problem(err)
}

func (check *checker) problem(err error, notes ...note) {
	if err == nil {
		return
	}

	if len(notes) > 0 {
		p, _ := err.(*Problem)
		if p == nil {
			p = &Problem{err: err}
		}
		err = problem(p, notes...)
	}

	if errs, _ := err.(interface{ Unwrap() []error }); errs != nil {
		check.problems = append(check.problems, errs.Unwrap()...)
	} else {
		check.problems = append(check.problems, err)
	}
}
