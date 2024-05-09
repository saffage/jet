package cgen

import (
	"fmt"

	"github.com/saffage/jet/checker"
)

type Error struct {
	Message string
	Sym     checker.Symbol
}

func newErrorf(sym checker.Symbol, format string, args ...any) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Sym:     sym,
	}
}

func (e *Error) Error() string { return e.Message }

func (gen *Generator) errorf(sym checker.Symbol, format string, args ...any) {
	gen.errors = append(gen.errors, newErrorf(sym, format, args...))
}
