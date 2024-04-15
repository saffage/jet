package jet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/symbol"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/log"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/scanner"
	"github.com/saffage/jet/token"
)

var WriteAstFileHandle *os.File

func reportError(cfg *config.Config, err error) {
	switch err := err.(type) {
	case scanner.Error:
		report.Error(cfg, "scanner error: "+err.Message, err.Start, err.End)

	case parser.Error:
		report.Report(log.KindError, cfg,
			"parser error: "+err.Message,
			err.Start,
			err.End,
		)

		for _, note := range err.Notes {
			report.Note(cfg, "parser note: "+note, token.Loc{}, token.Loc{})
		}

	case symbol.Error:
		start, end := token.Loc{}, token.Loc{}

		if err.Node != nil {
			start, end = err.Node.Pos(), err.Node.PosEnd()
		}

		report.Error(cfg, "checker error: "+err.Message, start, end)

		for _, note := range err.Notes {
			start, end = token.Loc{}, token.Loc{}

			if note.Node != nil {
				start, end = note.Node.Pos(), note.Node.PosEnd()
			}

			report.Note(cfg, "checker note: "+note.Message, start, end)
		}

	default:
		report.Error(cfg, err.Error(), token.Loc{}, token.Loc{})
	}
}

func process(
	cfg *config.Config,
	buffer []byte,
	fileid config.FileID,
	isRepl bool,
) {
	toks, scanErrors := scanner.Scan(buffer, fileid, scanner.SkipWhitespace)

	if len(scanErrors) > 0 {
		for _, err := range scanErrors {
			reportError(cfg, err)
		}

		return
	}

	nodes, parseErrors := &ast.List{}, []error(nil)

	if isRepl {
		expr, errors := parser.ParseExpr(cfg, buffer)
		nodes.Nodes = append(nodes.Nodes, expr)
		parseErrors = errors
	} else {
		nodes, parseErrors = parser.Parse(cfg, toks, parser.DefaultFlags|parser.Trace)
	}

	if len(parseErrors) > 0 {
		for _, err := range parseErrors {
			reportError(cfg, err)
		}

		return
	}

	spew.Dump(nodes)

	if WriteAstFileHandle != nil && nodes != nil {
		if bytes, err := json.MarshalIndent(nodes, "", "  "); err == nil {
			_, err := WriteAstFileHandle.Write(bytes)
			if err != nil {
				panic(err)
			}

			fmt.Printf("AST is writed to %s\n", WriteAstFileHandle.Name())
		} else {
			panic(err)
		}
	}

	defer func() {
		if err := recover(); err != nil {
			switch e := err.(type) {
			case symbol.Error:
				reportError(cfg, e)

			default:
				panic(err)
			}
		}
	}()

	symbol.NewModule(0, &ast.Ident{Name: cfg.Files[config.MainFileID].Name}, nil, nodes)
}