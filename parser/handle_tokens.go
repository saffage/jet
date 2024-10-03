package parser

import (
	"slices"

	"github.com/saffage/jet/token"
)

func (parse *parser) next() (prev token.Token) {
	if parse.current >= len(parse.tokens) {
		panic("EOF token was skipped or missing in the token stream")
	}

	if parse.tok.Kind == token.EOF {
		return parse.tok
	}

	prev = parse.tok
	parse.current++

	for parse.tok.Kind == token.Comment {
		parse.current++
	}

	if parse.flags&SkipIllegal != 0 {
		for parse.tok.Kind == token.Illegal {
			parse.current++
		}
	}

	parse.tok = parse.tokens[parse.current]
	return prev
}

func (parse *parser) match(kind token.Kind) bool {
	return parse.tok.Kind == kind
}

func (parse *parser) matchAny(kinds ...token.Kind) bool {
	return slices.Contains(kinds, parse.tok.Kind)
}

func (parse *parser) matchSequence(kinds ...token.Kind) bool {
	if len(kinds)+parse.current-1 >= len(parse.tokens) {
		return false
	}

	for i, kind := range kinds {
		if parse.tokens[parse.current+i].Kind != kind {
			return false
		}
	}

	return true
}

// Consumes a specified token or returns nil without emitting error.
func (parse *parser) consume(kind token.Kind) bool {
	_, ok := parse.take(kind)
	return ok
}

// Consumes a specified token or returns nil without emitting error.
func (parse *parser) consumeAny(kinds ...token.Kind) bool {
	_, ok := parse.takeAny(kinds...)
	return ok
}

// Consumes a specified token or returns false without emitting error.
func (parse *parser) take(kind token.Kind) (token.Token, bool) {
	if parse.match(kind) {
		return parse.next(), true
	}

	return token.Token{}, false
}

// Consumes a specified tokens or returns nil without emitting error.
func (parse *parser) takeAny(kinds ...token.Kind) (token.Token, bool) {
	if len(kinds) == 0 || parse.matchAny(kinds...) {
		return parse.next(), true
	}

	return token.Token{}, false
}

// Consumes a specified token or returns nil and emits error.
func (parse *parser) expect(kind token.Kind) (token.Token, error) {
	if tok, ok := parse.take(kind); ok {
		return tok, nil
	}

	return token.Token{}, parse.errorf(
		ErrUnexpectedToken,
		"expected %s",
		kind.UserString(),
	)
}
