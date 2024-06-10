package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
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
	restoreData []restoreData
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
	if len(tokens) < 1 {
		panic("expected at least 1 token (EOF)")
	}

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

func Parse(cfg *config.Config, tokens []token.Token, flags Flags) (*ast.StmtList, []error) {
	p := New(cfg, tokens, flags)
	stmts := p.parseDeclList()
	return stmts, p.Errors()
}

func ParseExpr(cfg *config.Config, tokens []token.Token, flags Flags) (ast.Node, []error) {
	p := New(cfg, tokens, flags)
	node := p.parseExpr()
	return node, p.Errors()
}

type restoreData struct {
	tokenIndex int
	errors     []error
}
