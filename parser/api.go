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
	scanner      *token.Scanner
	flags        Flags
	tracer       tracer
	errorHandler func(error)

	token.Token
}

func FromFile(
	file *text.File,
	scannerFlags token.ScannerFlags,
	flags Flags,
	handler func(error),
) *parser {
	return New(token.NewScannerFromFile(file, scannerFlags), flags, handler)
}

func New(s *token.Scanner, flags Flags, handler func(error)) *parser {
	if config.TraceParser {
		flags |= Trace
	}
	p := &parser{
		scanner:      s,
		flags:        flags,
		errorHandler: handler,
		tracer:       tracer{enabled: flags&Trace != 0, stack: []traceEntry{}},
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
	defer func(handler func(error)) {
		parse.errorHandler = handler
	}(parse.errorHandler)

	errs := []error{}
	parse.errorHandler = func(err error) { errs = append(errs, err) }
	stmts := parse.Parse()

	return stmts, report.Join(errs...)
}

func (parse *parser) handleError(err error) {
	if err != nil && parse.errorHandler != nil {
		parse.errorHandler(err)
	}
}
