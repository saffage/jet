package parser

import (
	"slices"

	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

func (parse *parser) next() (previous token.Token) {
	if parse.Kind != token.EOF {
		var tok token.Token

		for {
			tok = parse.scanner.NextToken()

			if tok.Kind != token.Illegal {
				break
			}
		}

		previous.Kind, parse.Kind = parse.Kind, tok.Kind
		previous.Data, parse.Data = parse.Data, tok.Data
		previous.Span, parse.Span = parse.Span, tok.Span
	}

	// if p.kind == token.Comment {
	// 		if strings.HasPrefix(p.tok.Data, "##") {
	// 			if p.commentGroup == nil {
	// 				p.commentGroup = &ast.CommentGroup{}
	// 			}
	//
	// 			p.commentGroup.Comments = append(p.commentGroup.Comments, &ast.Comment{
	// 				Data:  p.tok.Data[2:],
	// 				Start: p.tok.Start,
	// 				End:   p.tok.End,
	// 			})
	// 		}
	//
	// 	p.next()
	// }

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

func (parse *parser) expect(kind token.Kind) (token.Token, error) {
	if tok, ok := parse.consume(kind); ok {
		return tok, nil
	}

	return token.Token{}, errUnexpectedToken(parse.Span, kind)
}
