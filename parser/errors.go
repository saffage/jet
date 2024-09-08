package parser

import (
	"errors"
	"fmt"

	"github.com/saffage/jet/report"
	"github.com/saffage/jet/token"
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
	ErrExpectedDecl           = errors.New("expected declaration")
	ErrExpectedDeclAfterAttrs = errors.New("expected declaration after attribute list")
	ErrExpectedIdent          = errors.New("expected identifier")
)

type Error struct {
	err error

	Message string
	Start   token.Pos
	End     token.Pos

	isWarn     bool
	isInternal bool
}

func newError(err error, start, end token.Pos, args ...any) *Error {
	return &Error{
		err:     err,
		Start:   start,
		End:     end,
		Message: fmt.Sprint(args...),
	}
}

func newErrorf(err error, start, end token.Pos, format string, args ...any) *Error {
	return &Error{
		err:     err,
		Start:   start,
		End:     end,
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

func (e *Error) Report() {
	err, ok := e.err.(report.Reporter)
	if ok && err != nil {
		err.Report()
	}
	if !ok || e.Message != "" {
		tag := "parser"
		if e.isInternal {
			tag = "internal: " + tag
		}
		message := e.err.Error()
		if e.Message != "" {
			message += ": " + e.Message
		}
		if e.isWarn {
			report.TaggedWarningAt(tag, e.Start, e.End, message)
		} else {
			report.TaggedErrorAt(tag, e.Start, e.End, message)
		}
	}
}

func (parse *parser) error(err error, args ...any) error {
	return newError(err, parse.tok.Start, parse.tok.End, args...)
}

func (parse *parser) errorf(err error, format string, args ...any) error {
	return newErrorf(err, parse.tok.Start, parse.tok.End, format, args...)
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
