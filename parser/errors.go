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
)

type Error struct {
	err error

	Message string
	Range   token.Range

	isWarn     bool
	isInternal bool
}

func newError(err error, rng token.Range, args ...any) *Error {
	return &Error{
		err:     err,
		Range:   rng,
		Message: fmt.Sprint(args...),
	}
}

func newErrorf(err error, rng token.Range, format string, args ...any) *Error {
	return &Error{
		err:     err,
		Range:   rng,
		Message: fmt.Sprintf(format, args...),
	}
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
		Tag:   "parser",
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
	return newError(err, parse.tok.Range, args...)
}

func (parse *parser) errorf(err error, format string, args ...any) error {
	return newErrorf(err, parse.tok.Range, format, args...)
}

// func (p *parser) errorExpectedTokenAt(start, end token.Pos, tokens ...token.Kind) {
// 	if len(tokens) < 1 {
// 		panic("required at least 1 token")
// 	}
// 	kinds := lo.Map(tokens, func(kind token.Kind, _ int) string { return kind.UserString() })
// 	p.appendError(&Error{
// 		err:   ErrUnexpectedToken,
// 		Start: start,
// 		End:   end,
// 		Message: fmt.Sprintf(
// 			"want %s, got %s instead",
// 			strings.Join(kinds, " or "),
// 			p.tok.Kind.UserString(),
// 		),
// 	})
// }
