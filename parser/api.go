package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
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

type parser struct {
	// TODO: dis-embed fields as it's pollutes API.

	*token.Scanner
	token.Token

	tracer tracer
	flags  Flags
}

type Options struct {
	token.ScannerOptions

	ParserFlags Flags
}

func FromFile(file *text.File, opts ...Options) *parser {
	opt := Options{}

	if len(opts) != 0 {
		opt = opts[0]
	} else {
		opt.ScannerFlags = token.DefaultFlags
		opt.ParserFlags = DefaultFlags
	}

	scanner := token.NewScannerFromFile(file, opt.ScannerOptions)
	return New(scanner, opt)
}

func New(s *token.Scanner, opts ...Options) *parser {
	opt := Options{}

	if len(opts) != 0 {
		opt = opts[0]
	} else {
		opt.ScannerFlags = token.DefaultFlags
		opt.ParserFlags = DefaultFlags
	}

	if config.TraceParser {
		opt.ParserFlags |= Trace
	}

	p := &parser{
		Scanner: s,
		flags:   opt.ParserFlags,
		tracer:  tracer{enabled: opt.ParserFlags&Trace != 0},
	}
	p.next()
	return p
}

func (parse *parser) Parse() *ast.Stmts {
	decls := parse.listUntil(
		parse.declOrExpr,
		parse.Span,
		token.EOF,
		token.Semicolon,
		token.Newline,
	)

	if parse.flags&AllowTopLevelCode == 0 {
		for _, node := range decls {
			switch node.(type) {
			case nil:
				panic("unreachable")
			case *ast.ValueDecl,
				*ast.TypeDecl,
				*ast.TypeAliasDecl,
				*ast.BadNode:
				// OK
			default:
				parse.handleError(
					report.Build(ErrExpectedDecl).
						Tag("parse").
						Selection(node.Range(), "").
						Suggestion("top-level code is not allowed"),
				)
			}
		}
	}

	return &ast.Stmts{Items: decls}
}

func (parse *parser) ParseOrError() (*ast.Stmts, error) {
	errs := []error{}
	prev := parse.SetErrorHandler(func(err error) { errs = append(errs, err) })

	defer parse.SetErrorHandler(prev)

	stmts := parse.Parse()

	return stmts, report.Join(errs...)
}
