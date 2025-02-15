package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

type Flags int

const (
	Trace Flags = 1 << iota
	AllowTopLevelCode

	NoFlags      = Flags(0)
	DefaultFlags = NoFlags
)

func ParseFile(file *text.File, scannerFlags token.ScannerFlags, flags Flags) (*ast.Stmts, error) {
	return File(file, scannerFlags, flags).Parse()
}

// func ParseExpr(tokens []token.Token, flags Flags) (ast.Node, error) {
// 	return New(tokens, flags).ParseExpr()
// }

type parser struct {
	scanner *token.Scanner
	errors  []error

	// quick access
	token.Token

	// parser configuration
	flags Flags

	// debug
	trace       bool
	traceIndent int
}

// func NewFromIter(tokens iter.Seq[token.Token]) *parser {
// 	next, stop := iter.Pull(tokens)
//
// 	return &parser{
// 		nextToken: next,
// 		stop:      stop,
// 	}
// }

func File(file *text.File, scannerFlags token.ScannerFlags, flags Flags) *parser {
	return NewFrom(token.NewScanner(file.Content, file.ID, scannerFlags), flags)
}

func NewFrom(s *token.Scanner, flags Flags) *parser {
	p := &parser{
		scanner: s,
		flags:   flags,
		trace:   flags&Trace != 0,
	}
	p.next()
	return p
}

func (parse *parser) Parse() (*ast.Stmts, error) {
	if parse.flags&AllowTopLevelCode != 0 {
		decls, err := parse.sequence(parse.declOrExpr, token.Semicolon)
		return &ast.Stmts{Items: decls}, err
	}

	decls, err := parse.sequence(parse.decl, token.Semicolon)
	return &ast.Stmts{Items: decls}, err
}

// func (p *parser) ParseExpr() (ast.Node, error) {
// 	expr := p.parseExpr()
// 	return expr, errors.Join(p.errors...)
// }
