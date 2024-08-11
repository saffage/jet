package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/saffage/jet/report"
	"github.com/saffage/jet/token"
)

var (
	ErrBracketIsNeverClosed   = errors.New("bracket is never closed")
	ErrUnterminatedExpr       = errors.New("unterminated expression")
	ErrUnexpectedToken        = errors.New("unexpected token")
	ErrUnexpectedOperator     = errors.New("unexpected operator")
	ErrExpectedExpr           = errors.New("expected expression")
	ErrExpectedOperand        = errors.New("expected operand")
	ErrExpectedBlock          = errors.New("expected block")
	ErrExpectedBlockOrIf      = errors.New("expected block of 'if' clause")
	ErrExpectedType           = errors.New("expected type")
	ErrExpectedTypeName       = errors.New("expected type name")
	ErrExpectedTypeOrValue    = errors.New("expected type or value")
	ErrExpectedDecl           = errors.New("expected declaration")
	ErrExpectedDeclAfterAttrs = errors.New("expected declaration after attribute list")
	ErrExpectedIdent          = errors.New("expected identifier")
	ErrExpectedIdentAfterMut  = errors.New("expected identifier after 'mut'")
)

type Error struct {
	Start   token.Pos
	End     token.Pos
	Message string

	isWarn     bool
	isInternal bool

	err error
}

func (e Error) Error() string {
	if e.Message != "" {
		return e.err.Error() + ": " + e.Message
	}
	return e.err.Error()
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) Report() {
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
	p.errorAt(err, p.tok.Start, p.tok.End)
}

func (p *parser) errorf(err error, format string, args ...any) {
	p.errorfAt(err, p.tok.Start, p.tok.End, format, args...)
}

func (p *parser) errorExpectedToken(tokens ...token.Kind) {
	p.errorExpectedTokenAt(p.tok.Start, p.tok.End, tokens...)
}

func (p *parser) errorAt(err error, start, end token.Pos) {
	p.appendError(Error{
		err:   err,
		Start: start,
		End:   end,
	})
}

func (p *parser) errorfAt(err error, start, end token.Pos, format string, args ...any) {
	p.appendError(Error{
		err:     err,
		Start:   start,
		End:     end,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *parser) errorExpectedTokenAt(start, end token.Pos, tokens ...token.Kind) {
	if len(tokens) < 1 {
		panic("required at least 1 token")
	}
	buf := strings.Builder{}
	for i, tok := range tokens {
		if i != 0 {
			buf.WriteString(" or ")
		}
		buf.WriteString(tok.UserString())
	}
	p.appendError(Error{
		err:     ErrUnexpectedToken,
		Start:   start,
		End:     end,
		Message: fmt.Sprintf("want %s, got %s instead", buf.String(), p.tok.Kind.UserString()),
	})
}
