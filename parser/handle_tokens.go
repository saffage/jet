package parser

import (
	"fmt"
	"slices"

	"github.com/saffage/jet/token"
)

type restoreData struct {
	tokenIndex int
	errors     []error
}

func (p *parser) next() {
	if p.current >= len(p.tokens) {
		panic("EOF token was skipped or missing in the token stream")
	}

	if p.tok.Kind != token.EOF {
		p.current++
	}

	p.tok = p.tokens[p.current]

	// if p.commentGroup != nil && p.tok.End.Line > p.commentGroup.LocEnd().Line+1 {
	// 	p.commentGroup = nil
	// }

	if p.tok.Kind == token.Comment {
		// 	if strings.HasPrefix(p.tok.Data, "##") {
		// 		if p.commentGroup == nil {
		// 			p.commentGroup = &ast.CommentGroup{}
		// 		}

		// 		p.commentGroup.Comments = append(p.commentGroup.Comments, &ast.Comment{
		// 			Data:  p.tok.Data[2:],
		// 			Start: p.tok.Start,
		// 			End:   p.tok.End,
		// 		})
		// 	}

		p.next()
	}

	if p.flags&SkipIllegal != 0 && p.tok.Kind == token.Illegal {
		p.next()
	}
}

func (p *parser) drop(token token.Kind) bool {
	if p.match(token) {
		p.next()
		return true
	}
	return false
}

func (p *parser) match(token token.Kind) bool {
	return p.tok.Kind == token
}

func (p *parser) matchAny(tokens ...token.Kind) bool {
	return slices.Contains(tokens, p.tok.Kind)
}

func (p *parser) matchSequence(tokens ...token.Kind) bool {
	if len(tokens)+p.current-1 >= len(p.tokens) {
		return false
	}
	for i, kind := range tokens {
		if p.tokens[p.current+i].Kind != kind {
			return false
		}
	}
	return true
}

// Cunsumes a specified token or returns nil without emitting error.
func (p *parser) consume(kind token.Kind) *token.Token {
	if p.match(kind) {
		tok := p.tok
		p.next()
		return &tok
	}

	return nil
}

// Cunsumes a specified tokens or returns nil without emitting error.
func (p *parser) consumeAny(kinds ...token.Kind) *token.Token {
	if p.matchAny(kinds...) {
		tok := p.tok
		p.next()
		return &tok
	}

	return nil
}

// Cunsumes a specified token or returns nil and emits error.
func (p *parser) expect(kind token.Kind) *token.Token {
	if tok := p.consume(kind); tok != nil {
		return tok
	}

	p.errorExpectedToken(kind)
	return nil
}

// Cunsumes a specified token or returns nil and emits error.
func (p *parser) expectAny(kinds ...token.Kind) *token.Token {
	if tok := p.consumeAny(kinds...); tok != nil {
		return tok
	}

	p.errorExpectedToken(kinds...)
	return nil
}

// TODO rename it.
func (p *parser) save() (index int) {
	index = len(p.restoreData)
	p.restoreData = append(p.restoreData, restoreData{
		tokenIndex: p.current,
		errors:     p.errors,
	})
	return
}

func (p *parser) restore(index int) {
	if index >= len(p.restoreData) {
		panic(fmt.Sprintf("invalid restore index: %d (length is %d)", index, len(p.restoreData)))
	}

	data := p.restoreData[index]
	p.current = data.tokenIndex
	p.errors = data.errors
	p.tok = p.tokens[p.current]
	p.restoreData = p.restoreData[:index]
}
