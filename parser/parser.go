package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/scanner"
	"github.com/saffage/jet/token"
)

type Parser struct {
	config  *config.Config
	errors  []error
	tokens  []token.Token
	current int // index of the current token in `tokens` stream
	flags   Flags

	// Quick access
	tok token.Token

	// Debugging
	indent int

	// State
	restoreIndices []int
	annots         []*ast.Annotation
}

type Flags int

const (
	DefaultFlags Flags = SkipWhitespace | SkipIllegal
	NoFlags      Flags = 0
	Trace        Flags = 1 << iota
	SkipWhitespace
	SkipIllegal
)

func New(cfg *config.Config, tokens []token.Token, flags Flags) *Parser {
	return &Parser{
		config: cfg,
		tokens: tokens,
		flags:  flags,
		tok:    tokens[0],
	}
}

func (p *Parser) Errors() []error {
	return p.errors
}

func Parse(cfg *config.Config, tokens []token.Token, flags Flags) (*ast.List, []error) {
	p := New(cfg, tokens, flags)
	return p.parseStmtList(), p.Errors()
}

func ParseExpr(cfg *config.Config, input []byte) (ast.Node, []error) {
	toks, errors := scanner.Scan(input, 0, scanner.DefaultFlags)

	if len(errors) > 0 {
		return nil, errors
	}

	p := New(cfg, toks, DefaultFlags|Trace)
	return p.parseExpr(), p.Errors()
}
