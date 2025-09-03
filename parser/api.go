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
	*token.Scanner
	token.Token

	tracer tracer
	flags  Flags
}

func FromFile(
	file *text.File,
	scannerFlags token.ScannerFlags,
	flags Flags,
	errorHandler func(error),
) *parser {
	scanner := token.NewScannerFromFile(file, scannerFlags)
	return New(scanner, flags, errorHandler)
}

func New(s *token.Scanner, flags Flags, errorHandler func(error)) *parser {
	if config.TraceParser {
		flags |= Trace
	}
	p := &parser{
		Scanner: s,
		flags:   flags,
		tracer:  tracer{enabled: flags&Trace != 0, stack: []traceEntry{}},
	}
	p.ErrorHandler = errorHandler
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
			case *ast.LetDecl,
				*ast.ValDecl,
				*ast.VarDecl,
				*ast.TypeDef,
				*ast.TypeAlias,
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
	defer func(errorHandler func(error)) {
		parse.ErrorHandler = errorHandler
	}(parse.ErrorHandler)

	errs := []error{}
	parse.ErrorHandler = func(err error) { errs = append(errs, err) }
	stmts := parse.Parse()

	return stmts, report.Join(errs...)
}
