package parser

import (
	"slices"

	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

func (parse *parser) next() (previous token.Token) {
	if parse.Kind != token.EOF {
		for {
			tok := parse.scanner.NextToken()

			if tok.Kind != token.Illegal && tok.Kind != token.Comment {
				previous = parse.Token
				parse.Token = tok
				parse.tracer.tokenIndex++
				break
			}
		}
	}

	return
}

func (parse *parser) match(kind token.Kind) bool {
	return parse.Kind == kind
}

func (parse *parser) matchAny(kinds ...token.Kind) bool {
	return len(kinds) == 0 || slices.Contains(kinds, parse.Kind)
}

func (parse *parser) skip(kind token.Kind) (skipped bool) {
	if parse.match(kind) {
		parse.next()
		skipped = true
	}

	return
}

// Skips any of the specified tokens. Returns true if token was skipped.
func (parse *parser) skipAny(kinds ...token.Kind) (skipped bool) {
	if parse.matchAny(kinds...) {
		parse.next()
		skipped = true
	}

	return
}

// Skips any of the specified tokens. Returns true if token was skipped.
func (parse *parser) skipSeq(kinds ...token.Kind) bool {
	for _, kind := range kinds {
		if parse.Kind != kind {
			return false
		}

		parse.next()
	}

	return true
}

// Skips tokens until any specified token (or EOF).
func (parse *parser) skipUntil(kinds ...token.Kind) (skipped text.Span) {
	if len(kinds) == 0 {
		panic("must be at least 1 token")
	}

	skipped = parse.Span

	for !(parse.match(token.EOF) || parse.matchAny(kinds...)) {
		skipped.To = parse.next().Span.To
	}

	return
}

// Skips newline tokens.
func (parse *parser) skipNewLines() {
	for parse.Kind == token.Newline {
		parse.next()
	}
}

// Consumes a specified token or returns false without emitting error.
func (parse *parser) consume(kind token.Kind) (token.Token, bool) {
	if parse.match(kind) {
		return parse.next(), true
	}

	return token.Token{}, false
}

// Consumes a specified tokens or returns nil without emitting error.
func (parse *parser) consumeAny(kinds ...token.Kind) (token.Token, bool) {
	if len(kinds) == 0 || parse.matchAny(kinds...) {
		return parse.next(), true
	}

	return token.Token{}, false
}

// Skips any of the specified tokens. Returns true if token was skipped.
func (parse *parser) consumeSeq(kinds ...token.Kind) ([]token.Token, bool) {
	tokens := make([]token.Token, 0, len(kinds))
	consumedAll := true

	for _, kind := range kinds {
		tok, ok := parse.consume(kind)

		if !ok {
			consumedAll = false
			break
		}

		tokens = append(tokens, tok)
	}

	return tokens, consumedAll
}

func (parse *parser) expect(kind token.Kind) token.Token {
	tok, ok := parse.consume(kind)

	if !ok {
		panic(errUnexpectedToken(parse.Span, parse.Kind, kind))
	}

	return tok
}

func (parse *parser) expectAny(kinds ...token.Kind) token.Token {
	tok, ok := parse.consumeAny(kinds...)

	if !ok {
		panic(errUnexpectedToken(parse.Span, parse.Kind, kinds...))
	}

	return tok
}
