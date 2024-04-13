package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/token"
)

type Error struct {
	Message    string
	Details    string
	Notes      []string
	Start, End token.Loc
}

func (e Error) Error() string {
	if e.Details == "" {
		return e.Message
	}
	return e.Message + "; " + e.Details
}

func (p *Parser) addError(err error) {
	p.errors = append(p.errors, err)
}

func NewError(message string, start, end token.Loc, details string, notes ...string) Error {
	return Error{
		Message: message,
		Details: details,
		Notes:   notes,
		Start:   start,
		End:     end,
	}
}

func (p *Parser) error(message string, start, end token.Loc, details ...any) {
	if p.flags&Trace != 0 {
		p.trace(color.RedString("error"))
		defer p.untrace()
	}

	p.addError(NewError(message, start, end, fmt.Sprint(details...)))
}

func (p *Parser) errorExpected(message string, start, end token.Loc, details ...any) {
	message = fmt.Sprintf("expected %s, found %s", message, p.tok.Kind.String())
	p.error(message, start, end, details...)
}

func (p *Parser) errorExpectedToken(start, end token.Loc, tokens ...token.Kind) {
	if len(tokens) < 1 {
		panic("Parser.errorExpectedToken: required at least 1 token")
	}

	tokenStrs := []string{}

	for _, tok := range tokens {
		tokenStrs = append(tokenStrs, tok.String())
	}

	message := fmt.Sprintf("expected %s, found %s", strings.Join(tokenStrs, " or "), p.tok.Kind.String())
	p.error(message, start, end)
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
