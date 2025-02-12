package parser

import (
	"errors"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

func Parse(tokens []token.Token, flags Flags) (*ast.StmtList, error) {
	return New(tokens, flags).Parse()
}

func ParseExpr(tokens []token.Token, flags Flags) (ast.Node, error) {
	return New(tokens, flags).ParseExpr()
}

type parser struct {
	errors  []error
	tokens  []token.Token
	current int // index of the current token in `tokens` stream
	flags   Flags

	// Quick access
	tok token.Token

	// Debugging
	indent int

	// State
	restoreData []restoreData
}

func New(tokens []token.Token, flags Flags) *parser {
	if len(tokens) < 1 {
		panic("expected at least 1 token (EOF)")
	}

	if tokens[len(tokens)-1].Kind != token.EOF {
		panic("expected EOF token is the end of the stream")
	}

	return &parser{
		tokens: tokens,
		flags:  flags,
		tok:    tokens[0],
	}
}

func (p *parser) Parse() (*ast.StmtList, error) {
	decls := p.parseDeclList()
	return decls, errors.Join(p.errors...)
}

func (p *parser) ParseExpr() (ast.Node, error) {
	expr := p.parseExpr()
	return expr, errors.Join(p.errors...)
}

type Flags int

const (
	Trace Flags = 1 << iota
	SkipIllegal

	NoFlags      = Flags(0)
	DefaultFlags = SkipIllegal
)
