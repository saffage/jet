package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/saffage/jet/token"
)

func (p *Parser) next() {
	if p.current >= len(p.tokens) {
		panic("EOF token was skipped or missing in the token stream")
	}

	if p.tok.Kind != token.EOF {
		p.current++
	}

	p.tok = p.tokens[p.current]

	if p.flags&SkipWhitespace != 0 && p.tok.Kind == token.Whitespace {
		p.next()
	}

	if p.flags&SkipIllegal != 0 && p.tok.Kind == token.Illegal {
		p.next()
	}
}

func (p *Parser) match(kinds ...token.Kind) bool {
	return len(kinds) > 0 && slices.Contains(kinds, p.tok.Kind)
}

// Cunsumes a specified token or returns nil without emitting error.
func (p *Parser) consume(kinds ...token.Kind) *token.Token {
	if len(kinds) == 0 || p.match(kinds...) {
		tok := p.tok
		p.next()
		return &tok
	}

	return nil
}

// Cunsumes a specified token or returns nil and emits error.
func (p *Parser) expect(kinds ...token.Kind) *token.Token {
	if tok := p.consume(kinds...); tok != nil {
		return tok
	}

	kindStrs := []string{}

	for _, kind := range kinds {
		kindStrs = append(kindStrs, kind.String())
	}

	p.errorExpected(strings.Join(kindStrs, " or "), p.tok.Start, p.tok.End)
	return nil
}

func (p *Parser) save() (tokenIndex int) {
	p.restoreIndices = append([]int{p.current}, p.restoreIndices...)
	return p.current
}

func (p *Parser) restore(tokenIndex int) {
	if tokenIndex < 0 {
		panic(fmt.Sprintf("invalid restore index '%d'", tokenIndex))
	}

	for i, idx := range p.restoreIndices {
		if tokenIndex == idx {
			p.current = tokenIndex
			p.tok = p.tokens[p.current]
			p.restoreIndices = p.restoreIndices[i+1:]
			return
		}
	}

	panic(fmt.Sprintf("unsaved restore point '%d', active point is %v", tokenIndex, p.restoreIndices))
}
