package parser

import (
	"errors"
	"fmt"

	"github.com/saffage/jet/parser/token"
	"github.com/saffage/jet/report"
)

var (
	ErrUnterminatedList       = errors.New("unterminated list, bracket is never closed")
	ErrUnterminatedExpr       = errors.New("unterminated expression")
	ErrUnexpectedToken        = errors.New("unexpected token")
	ErrUnexpectedOperator     = errors.New("unexpected operator")
	ErrExpectedExpr           = errors.New("expected expression")
	ErrExpectedOperand        = errors.New("expected operand")
	ErrExpectedBlock          = errors.New("expected block")
	ErrExpectedBlockOrIf      = errors.New("expected block or 'if' clause")
	ErrExpectedType           = errors.New("expected type")
	ErrExpectedTypeVar        = errors.New("expected type variable")
	ErrExpectedTypeOrBlock    = errors.New("expected type or block")
	ErrExpectedDecl           = errors.New("expected declaration")
	ErrExpectedDeclAfterAttrs = errors.New("expected declaration after attribute list")
	ErrExpectedIdent          = errors.New("expected identifier")
	ErrExpectedPattern        = errors.New("expected pattern")
)

type Error struct {
	err error

	Message string
	Range   token.Range

	isWarn     bool
	isInternal bool
}

func (e *Error) Error() string {
	if e.Message != "" {
		return "invalid syntax: " + e.err.Error() + ": " + e.Message
	}
	return "invalid syntax: " + e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Info() *report.Info {
	if p, ok := e.err.(report.Problem); ok {
		return p.Info()
	}

	info := &report.Info{
		Tag:   "syntax",
		Title: e.err.Error(),
		Range: e.Range,
	}

	if e.Message != "" {
		info.Title += ": " + e.Message
	}

	if e.isInternal {
		info.Tag += ", internal"
	}

	if e.isWarn {
		info.Kind = report.KindWarning
	}

	return info
}

func (parse *parser) error(err error, args ...any) error {
	return &Error{
		err:     err,
		Range:   parse.tok.Range,
		Message: fmt.Sprint(args...),
	}
}

func (parse *parser) errorf(err error, format string, args ...any) error {
	return &Error{
		err:     err,
		Range:   parse.tok.Range,
		Message: fmt.Sprintf(format, args...),
	}
}
