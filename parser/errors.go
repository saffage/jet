package parser

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/saffage/jet/internal/report"
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

func (e Error) Is(err error) bool {
	return e.err == err
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

func (p *Parser) lastErrorIs(err error) bool {
	if len(p.errors) > 0 {
		return errors.Is(p.errors[len(p.errors)-1], err)
	}

	return false
}

func (p *Parser) appendError(err error) {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	p.errors = append(p.errors, err)
}

func (p *Parser) error(err error) {
	p.errorAt(err, p.tok.Start, p.tok.End)
}

func (p *Parser) errorf(err error, format string, args ...any) {
	p.errorfAt(err, p.tok.Start, p.tok.End, format, args...)
}

func (p *Parser) errorExpectedToken(tokens ...token.Kind) {
	p.errorExpectedTokenAt(p.tok.Start, p.tok.End, tokens...)
}

func (p *Parser) errorAt(err error, start, end token.Pos) {
	p.appendError(Error{
		err:   err,
		Start: start,
		End:   end,
	})
}

func (p *Parser) errorfAt(err error, start, end token.Pos, format string, args ...any) {
	p.appendError(Error{
		err:     err,
		Start:   start,
		End:     end,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *Parser) errorExpectedTokenAt(start, end token.Pos, tokens ...token.Kind) {
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
		err:     ErrorUnexpectedToken,
		Start:   start,
		End:     end,
		Message: fmt.Sprintf("want %s, got %s instead", buf.String(), p.tok.Kind.UserString()),
	})
}

func (p *Parser) skip(to ...token.Kind) (start, end token.Pos) {
	if len(to) == 0 {
		to = endOfExprKinds
	}

	start = p.tok.Start

	for p.tok.Kind != token.EOF && !slices.Contains(to, p.tok.Kind) {
		end = p.tok.End
		p.next()
	}

	if p.flags&Trace != 0 && end.IsValid() {
		// TODO must be removed
		warn := Error{
			Start:      start,
			End:        end,
			Message:    "tokens was skipped for some reason",
			isWarn:     true,
			isInternal: true,
		}
		p.errors = append(p.errors, warn)
	}

	return
}

var (
	endOfStmtKinds = []token.Kind{
		token.Semicolon,
		token.NewLine,
	}
	endOfExprKinds = append(endOfStmtKinds, []token.Kind{
		token.Comma,
		token.RParen,
		token.RCurly,
		token.RBracket,
	}...)
)
