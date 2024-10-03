package parser

import (
	"slices"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser/token"
)

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

func (parse *parser) Parse() (*ast.Stmts, error) {
	return parse.decls()
}

func (parse *parser) ParseExpr() (ast.Node, error) {
	return parse.expr()
}

func (parse *parser) MustParse() *ast.Stmts {
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
