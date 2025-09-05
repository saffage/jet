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

	WarnRedundantRebinding = errors.New("redundant rebinding")
)

func errExpectedBlock(span text.Span) report.Builder {
	return report.Build(ErrExpectedBlock).
		Tag("parse").
		Selection(span, "")
}

func errExpectedDecl(span text.Span) report.Builder {
	return report.Build(ErrExpectedDecl).
		Tag("parse").
		Selection(span, "")
}

func errExpectedExpr(span text.Span) report.Builder {
	return report.Build(ErrExpectedExpr).
		Tag("parse").
		Selection(span, "")
}

func errExpectedOperand(span text.Span) report.Builder {
	return report.Build(ErrExpectedOperand).
		Tag("parse").
		Selection(span, "this must be an operand")
}

func errExpectedPattern(span text.Span, pattern int) report.Builder {
	return report.Build(ErrExpectedPattern).
		Tag("parse").
		Selection(span, "")
}

func errExpectedType(span text.Span) report.Builder {
	return report.Build(ErrExpectedType).
		Tag("parse").
		Selection(span, "")
}

func errUnexpectedToken[Found, Expected any](
	span text.Span,
	found Found,
	expected ...Expected,
) report.Builder {
	message := ""

	if len(expected) > 0 {
		message = fmt.Sprintf(
			"expected %s here instead of %v",
			prettyJoin(expected),
			found,
		)
	}

	return report.Build(ErrUnexpectedToken).
		Tag("parse").
		Selection(span, message)
}

func errUnimplementedFeature(span text.Span, featureName string) report.Builder {
	return report.Build(ErrUnimplementedFeature).
		Tag("parse").
		Selection(
			span,
			fmt.Sprintf("feature '%s' is not implemented", featureName),
		)
}

func errUnterminatedExpr(span text.Span, delimiters ...token.Kind) report.Builder {
	return report.Build(ErrUnterminatedExpr).
		Tag("parse").
		Selection(
			span,
			fmt.Sprintf("expected %s after this", prettyJoinString(delimiters)),
		)
}

func errUnterminatedList(span text.Span) report.Builder {
	return report.Build(ErrUnterminatedList).
		Tag("parse").
		Selection(span, "")
}

func warnRedundantRebinding(span text.Span, name string) report.Builder {
	return report.Build(WarnRedundantRebinding).
		Tag("parse").
		Level(report.LevelWarning).
		SelectionF(span, "This could be replaced with just `%s`", name)
}

func prettyJoin[T any](items []T) string {
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

func prettyJoinString[T fmt.Stringer](items []T) string {
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
