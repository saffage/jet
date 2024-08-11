package parser

import (
	"errors"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

func Parse(tokens []token.Token, flags Flags) (*ast.StmtList, error) {
	return New(tokens, flags).Parse()
}

func MustParse(tokens []token.Token, flags Flags) *ast.StmtList {
	return New(tokens, flags).MustParse()
}

func ParseExpr(tokens []token.Token, flags Flags) (ast.Node, error) {
	return New(tokens, flags).ParseExpr()
}

func MustParseExpr(tokens []token.Token, flags Flags) ast.Node {
	return New(tokens, flags).MustParseExpr()
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

func (parse *parser) Parse() (*ast.StmtList, error) {
	decls, _ := parse.decls().(*ast.StmtList)
	return decls, errors.Join(parse.errors...)
}

func (parse *parser) ParseExpr() (ast.Node, error) {
	expr := parse.expr()
	return expr, errors.Join(parse.errors...)
}

func (p *parser) MustParse() *ast.StmtList {
	decls, err := p.Parse()
	if err != nil {
		panic(err)
	}
	return decls
}

func (p *parser) MustParseExpr() ast.Node {
	expr, err := p.ParseExpr()
	if err != nil {
		panic(err)
	}
	return expr
}

type Flags int

const (
	Trace Flags = 1 << iota
	SkipIllegal

	NoFlags      = Flags(0)
	DefaultFlags = NoFlags
)
