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

type parser struct {
	scanner *token.Scanner
	flags   Flags

	token.Token
	tracer
}

func ParseFile(file *text.File, scannerFlags token.ScannerFlags, flags Flags) (*ast.Stmts, error) {
	return File(file, scannerFlags, flags).Parse()
}

func File(file *text.File, scannerFlags token.ScannerFlags, flags Flags) *parser {
	return New(token.NewScanner(file.Content, file.ID, scannerFlags), flags)
}

func New(s *token.Scanner, flags Flags) *parser {
	p := &parser{
		scanner: s,
		flags:   flags,
		tracer:  tracer{enabled: flags&Trace != 0, stack: []traceEntry{}},
	}
	p.next()
	return p
}

func (parse *parser) Parse() (*ast.Stmts, error) {
	decls, err := parse.sequence(parse.declOrExpr, token.Semicolon)

	if err != nil {
		return nil, err
	}

	if parse.flags&AllowTopLevelCode == 0 {
		for _, node := range decls {
			switch node.(type) {
			case *ast.LetDecl, *ast.TypeDef, *ast.TypeAlias:
			case nil:
				panic("unreachable")
			default:
				return nil, errExpectedDecl(node.Range())
			}
		}
	}

	return &ast.Stmts{
		Items:      decls,
		DesiredPos: parse.Span.From,
	}, nil
}
