package parser

import (
	"errors"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/token"
)

type Parser struct {
	cfg     *config.Config
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
		cfg:    cfg,
		tokens: tokens,
		flags:  flags,
		tok:    tokens[0],
	}
}

func Parse(cfg *config.Config, tokens []token.Token, flags Flags) (*ast.StmtList, error) {
	p := New(cfg, tokens, flags)
	stmts := p.parseDeclList()
	return stmts, errors.Join(p.errors...)
}

func MustParse(cfg *config.Config, tokens []token.Token, flags Flags) *ast.StmtList {
	stmts, err := Parse(cfg, tokens, flags)
	if err != nil {
		panic(err)
	}
	return stmts
}

func ParseExpr(cfg *config.Config, tokens []token.Token, flags Flags) (ast.Node, error) {
	p := New(cfg, tokens, flags)
	node := p.parseExpr()
	return node, errors.Join(p.errors...)
}

func MustParseExpr(cfg *config.Config, tokens []token.Token, flags Flags) ast.Node {
	expr, err := ParseExpr(cfg, tokens, flags)
	if err != nil {
		panic(err)
	}
	return expr
}
