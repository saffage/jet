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
	ErrorInvalidBinaryOperator  = errors.New("invalid binary operator")
	ErrorBracketIsNeverClosed   = errors.New("bracket is never closed")
	ErrorUnterminatedExpr       = errors.New("unterminated expression")
	ErrorUnexpectedToken        = errors.New("unexpected token")
	ErrorExpectedExpr           = errors.New("expected expression")
	ErrorExpectedOperand        = errors.New("expected operand")
	ErrorExpectedBlock          = errors.New("expected block")
	ErrorExpectedBlockOrIf      = errors.New("expected block of 'if' clause")
	ErrorExpectedType           = errors.New("expected type")
	ErrorExpectedTypeName       = errors.New("expected type name")
	ErrorExpectedTypeOrValue    = errors.New("expected type or value")
	ErrorExpectedDecl           = errors.New("expected declaration")
	ErrorExpectedDeclAfterAttrs = errors.New("expected declaration after attribute list")
	ErrorExpectedIdent          = errors.New("expected identifier")
	ErrorExpectedIdentAfterMut  = errors.New("expected identifier after 'mut'")
)

type Error struct {
	err error

	Message        string
	SelectionRange text.Span
	Warning        bool
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Is(target error) bool {
	return e.err == target
}

func (e *Error) Info() *report.Info {
	level := report.LevelError

	if e.Warning {
		level = report.LevelWarning
	}

	return &report.Info{
		Tag:       "parse",
		Title:     e.Error(),
		Selection: report.Selection{Range: e.SelectionRange},
		Level:     level,
	}
}

func (p *parser) lastErrorIs(err error) bool {
	if len(p.errors) > 0 {
		return errors.Is(p.errors[len(p.errors)-1], err)
	}

	return false
}

func (p *parser) appendError(err error) {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	p.errors = append(p.errors, err)
}

func (p *parser) error(err error) {
	p.errorAt(err, p.tok.Span.From, p.tok.Span.To)
}

func (p *parser) errorf(err error, format string, args ...any) {
	p.errorfAt(err, p.tok.Span.From, p.tok.Span.To, format, args...)
}

func (p *parser) errorExpectedToken(tokens ...token.Kind) {
	p.errorExpectedTokenAt(p.tok.Span.From, p.tok.Span.To, tokens...)
}

func (p *parser) errorAt(err error, start, end text.Pos) {
	p.appendError(&Error{
		err:            err,
		SelectionRange: text.Span{From: start, To: end},
	})
}

func (p *parser) errorfAt(err error, start, end text.Pos, format string, args ...any) {
	p.appendError(&Error{
		err:            err,
		SelectionRange: text.Span{From: start, To: end},
		Message:        fmt.Sprintf(format, args...),
	})
}

func (p *parser) errorExpectedTokenAt(start, end text.Pos, tokens ...token.Kind) {
	if len(tokens) < 1 {
		panic("required at least 1 token")
	}
	buf := strings.Builder{}
	for i, tok := range tokens {
		if i != 0 {
			buf.WriteString(" or ")
		}
		buf.WriteString(tok.String())
	}
	p.appendError(&Error{
		err:            ErrorUnexpectedToken,
		SelectionRange: text.Span{From: start, To: end},
		Message:        fmt.Sprintf("want %s, got %s instead", buf.String(), p.tok.Kind.String()),
	})
}
