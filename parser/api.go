package parser

import (
	"errors"
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/token"
)

var (
	ErrEmptyStream       = errors.New("token stream is empty")
	ErrMissingEOFToken   = errors.New("missing EOF token at the end")
	ErrDuplicateEOFToken = errors.New("duplicate EOF token")
)

func Parse(tokens []token.Token, flags Flags) (*ast.StmtList, error) {
	p, err := New(tokens, flags)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}

func MustParse(tokens []token.Token, flags Flags) *ast.StmtList {
	p, err := New(tokens, flags)
	if err != nil {
		panic(err)
	}
	return p.MustParse()
}

func ParseExpr(tokens []token.Token, flags Flags) (ast.Node, error) {
	p, err := New(tokens, flags)
	if err != nil {
		return nil, err
	}
	return p.ParseExpr()
}

func MustParseExpr(tokens []token.Token, flags Flags) ast.Node {
	p, err := New(tokens, flags)
	if err != nil {
		panic(err)
	}
	return p.MustParseExpr()
}

type parser struct {
	tokens  []token.Token
	tok     token.Token // For quick access.
	flags   Flags
	current int
}

func New(tokens []token.Token, flags Flags) (*parser, error) {
	if len(tokens) == 0 {
		return nil, ErrEmptyStream
	}
	if !slices.ContainsFunc(
		tokens,
		func(tok token.Token) bool {
			return tok.Kind == token.EOF
		},
	) {
		return nil, ErrMissingEOFToken
	}
	return &parser{tokens: tokens, flags: flags, tok: tokens[0]}, nil
}

func (parse *parser) Parse() (*ast.StmtList, error) {
	decls, err := parse.decls()
	stmts, _ := decls.(*ast.StmtList)
	return stmts, err
}

func (parse *parser) ParseExpr() (ast.Node, error) {
	return parse.expr()
}

func (parse *parser) MustParse() *ast.StmtList {
	decls, err := parse.Parse()
	if err != nil {
		panic(err)
	}
	return decls
}

func (parse *parser) MustParseExpr() ast.Node {
	expr, err := parse.ParseExpr()
	if err != nil {
		panic(err)
	}
	return expr
}

type Flags int

const (
	SkipIllegal Flags = 1 << iota

	NoFlags      = Flags(0)
	DefaultFlags = NoFlags
)
