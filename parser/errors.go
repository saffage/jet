package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

var (
	ErrExpectedBlock   = errors.New("expected block")
	ErrExpectedDecl    = errors.New("expected declaration")
	ErrExpectedExpr    = errors.New("expected expression")
	ErrExpectedIdent   = errors.New("expected identifier")
	ErrExpectedOperand = errors.New("expected operand")
	ErrExpectedPattern = errors.New("expected pattern")
	ErrExpectedType    = errors.New("expected type")
	// ErrExpectedTypeOrBlockError = errors.New("expected type or block")
	// ErrExpectedTypeVar          = errors.New("expected type variable")
	ErrUnexpectedToken      = errors.New("unexpected token")
	ErrUnimplementedFeature = errors.New("unimplemented feature")
	ErrUnterminatedExpr     = errors.New("unterminated expression")
	ErrUnterminatedList     = errors.New("unterminated list, bracket is never closed")
	// ErrInvalidBinaryOperator    = errors.New("invalid binary operator")
)

func errExpectedBlock(span text.Span) error {
	return report.Build(ErrExpectedBlock).
		Tag("parse").
		Selection(span, "")
}

func errExpectedDecl(span text.Span) error {
	return report.Build(ErrExpectedDecl).
		Tag("parse").
		Selection(span, "")
}

func errExpectedExpr(span text.Span) error {
	return report.Build(ErrExpectedExpr).
		Tag("parse").
		Selection(span, "")
}

func errExpectedOperand(span text.Span) error {
	return report.Build(ErrExpectedOperand).
		Tag("parse").
		Selection(span, "this must be an operand")
}

func errExpectedPattern(span text.Span, pattern int) error {
	return report.Build(ErrExpectedPattern).
		Tag("parse").
		Selection(span, "")
}

func errExpectedType(span text.Span) error {
	return report.Build(ErrExpectedType).
		Tag("parse").
		Selection(span, "")
}

func errUnexpectedToken(span text.Span, expected ...any) error {
	message := ""
	tokens := prettyJoinAny(expected)

	if len(expected) > 0 {
		message = "expected " + tokens + " here"
	}

	return report.Build(ErrUnexpectedToken).
		Tag("parse").
		Selection(span, message)
}

func errUnimplementedFeature(span text.Span, featureName string) error {
	return report.Build(ErrUnimplementedFeature).
		Tag("parse").
		Selection(
			span,
			fmt.Sprintf("feature '%s' is not implemented", featureName),
		)
}

func errUnterminatedExpr(span text.Span, delimiters ...token.Kind) error {
	return report.Build(ErrUnterminatedExpr).
		Tag("parse").
		Selection(
			span,
			fmt.Sprintf("expected %s after this", prettyJoin(delimiters)),
		)
}

func errUnterminatedList(span text.Span) error {
	return report.Build(ErrUnterminatedList).
		Tag("parse").
		Selection(span, "")
}

// func (p *parser) lastErrorIs(err error) bool {
// 	if len(p.errors) > 0 {
// 		return errors.Is(p.errors[len(p.errors)-1], err)
// 	}

// 	return false
// }

// func (p *parser) appendError(err error) {
// 	if p.flags&Trace != 0 {
// 		defer un(trace(p))
// 	}

// 	p.errors = append(p.errors, err)
// }

// func (p *parser) error(err error) {
// 	p.errorAt(err, p.span)
// }

// func (p *parser) errorf(err error, format string, args ...any) {
// 	p.errorfAt(err, p.span, fmt.Sprintf(format, args...))
// }

// func (p *parser) errorAt(err error, span text.Span, message ...any) {
// 	p.appendError(&Error{
// 		Message:   fmt.Sprint(message...),
// 		err:       err,
// 		Selection: span,
// 	})
// }

// func (p *parser) errorfAt(err error, span text.Span, format string, args ...any) {
// 	p.appendError(&Error{
// 		err:       err,
// 		Selection: span,
// 		Message:   fmt.Sprintf(format, args...),
// 	})
// }

func prettyJoinAny(items []any) string {
	buf := strings.Builder{}

	switch {
	case len(items) > 1:
		n := len(items) - 1

		for i, item := range items[:n] {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprint(item))
		}

		buf.WriteString(" or ")
		buf.WriteString(fmt.Sprint(items[n]))

	case len(items) == 1:
		buf.WriteString(fmt.Sprint(items[0]))
	}

	return buf.String()
}

func prettyJoin[T fmt.Stringer](items []T) string {
	buf := strings.Builder{}

	switch {
	case len(items) > 1:
		n := len(items) - 1

		for i, item := range items[:n] {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(item.String())
		}

		buf.WriteString(" or ")
		buf.WriteString(items[n].String())

	case len(items) == 1:
		buf.WriteString(items[0].String())
	}

	return buf.String()
}
