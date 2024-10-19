package parser

import (
	"errors"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser/token"
)

var (
	ErrEmptyStream       = errors.New("token stream is empty")
	ErrMissingEOFToken   = errors.New("missing EOF token at the end")
	ErrDuplicateEOFToken = errors.New("duplicate EOF token")
)

type Flags int

const (
	SkipIllegal Flags = 1 << iota

	NoFlags      = Flags(0)
	DefaultFlags = NoFlags
)

func Parse(tokens []token.Token, flags Flags) (ast.Stmts, error) {
	p, err := New(tokens, flags)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}

func MustParse(tokens []token.Token, flags Flags) ast.Stmts {
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
