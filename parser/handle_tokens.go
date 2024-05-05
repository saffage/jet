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
		kindStrs = append(kindStrs, kind.UserString())
	}

	p.errorExpected(p.tok.Start, p.tok.End, strings.Join(kindStrs, " or "))
	return nil
}

func (p *Parser) save() (tokenIndex int) {
	p.restoreData = append(p.restoreData, restoreData{
		index:  p.current,
		errors: p.errors,
	})
	return p.current
}

func (p *Parser) restore(tokenIndex int) {
	if tokenIndex < 0 {
		panic(fmt.Sprintf("invalid restore index '%d'", tokenIndex))
	}

	for i := len(p.restoreData) - 1; i >= 0; i-- {
		data := p.restoreData[i]

		if tokenIndex == data.index {
			p.current = tokenIndex
			p.errors = data.errors
			p.tok = p.tokens[p.current]
			p.restoreData = p.restoreData[i+1:]
			return
		}
	}

	panic(fmt.Sprintf("unsaved restore point '%d', active point is %v", tokenIndex, p.restoreData))
}
