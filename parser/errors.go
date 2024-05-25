package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/token"
)

type Error struct {
	Start, End token.Loc
	Message    string
	Notes      []string
}

func NewError(start, end token.Loc, message string) Error {
	return Error{
		Start:   start,
		End:     end,
		Message: message,
	}
}

func NewErrorf(start, end token.Loc, format string, args ...any) Error {
	return NewError(start, end, fmt.Sprintf(format, args...))
}

func (e Error) Error() string { return e.Message }

func (e Error) Report() {
	report.TaggedErrorAt("parser", e.Start, e.End, e.Message)

	for _, note := range e.Notes {
		report.TaggedNote("parser", note)
	}
}

func (p *Parser) addError(err error) {
	p.errors = append(p.errors, err)
}

func (p *Parser) error(start, end token.Loc, message string) {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	p.addError(NewError(start, end, message))
}

func (p *Parser) errorf(start, end token.Loc, format string, args ...any) {
	if p.flags&Trace != 0 {
		defer un(trace(p))
	}

	p.addError(NewErrorf(start, end, format, args...))
}

func (p *Parser) errorExpected(start, end token.Loc, message string) {
	message = fmt.Sprintf("expected %s, found %s", message, p.tok.Kind.UserString())
	p.error(start, end, message)
}

func (p *Parser) errorExpectedToken(start, end token.Loc, tokens ...token.Kind) {
	if len(tokens) < 1 {
		panic("required at least 1 token")
	}

	tokenStrs := []string{}

	for _, tok := range tokens {
		tokenStrs = append(tokenStrs, tok.UserString())
	}

	message := fmt.Sprintf("expected %s, found %s", strings.Join(tokenStrs, " or "), p.tok.Kind.UserString())
	p.error(start, end, message)
}

func (p *Parser) skipTo(to ...token.Kind) (start, end token.Loc) {
	if len(to) == 0 {
		to = endOfExprKinds
	}

	if p.flags&Trace != 0 {
		defer func(before string) {
			after := p.tok.String()
			fmt.Printf("parser: skipped tokens from %s to %s\n", before, after)
		}(p.tok.String())
	}

	to = append(to, token.EOF)
	start = p.tok.Start

	for !slices.Contains(to, p.tok.Kind) {
		p.next()
	}

	end = p.tok.End

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
