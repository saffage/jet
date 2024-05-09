package jet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/scanner"
	"github.com/saffage/jet/token"
)

func reportError(err error) {
	switch err := err.(type) {
	case scanner.Error:
		report.TaggedErrorAt("scanner", err.Message, err.Start, err.End)

	case parser.Error:
		report.TaggedErrorAt(
			"parser",
			err.Message,
			err.Start,
			err.End,
		)

		for _, note := range err.Notes {
			report.TaggedNoteAt("parser", note, token.Loc{}, token.Loc{})
		}

	case *checker.Error:
		start, end := token.Loc{}, token.Loc{}

		if err.Node != nil {
			start, end = err.Node.Pos(), err.Node.LocEnd()
		}

		report.TaggedErrorAt("checker", err.Message, start, end)

		for _, note := range err.Notes {
			start, end = token.Loc{}, token.Loc{}

			if note.Node != nil {
				start, end = note.Node.Pos(), note.Node.LocEnd()
			}

			report.TaggedNoteAt("checker", note.Message, start, end)
		}

	default:
		report.TaggedErrorAt("", err.Error(), token.Loc{}, token.Loc{})
	}
}

func process(
	cfg *config.Config,
	buffer []byte,
	fileID config.FileID,
	isRepl bool,
) {
	toks, scanErrors := scanner.Scan(buffer, fileID, scanner.SkipWhitespace|scanner.SkipComments)

	if len(scanErrors) > 0 {
		for _, err := range scanErrors {
			reportError(err)
		}

		return
	}

	parserFlags := parser.DefaultFlags

	if TraceParser {
		parserFlags |= parser.Trace
	}

	nodeList, parseErrors := parser.Parse(cfg, toks, parserFlags)

	if len(parseErrors) > 0 {
		for _, err := range parseErrors {
			reportError(err)
		}

		return
	}

	if nodeList == nil {
		return
	}

	fmt.Println("recreated AST:")
	for i, node := range nodeList.Nodes {
		if _, isEmpty := node.(*ast.Empty); i < len(nodeList.Nodes)-1 || !isEmpty {
			fmt.Println(color.HiGreenString(node.String()))
		}
	}

	if WriteAstFileHandle != nil && nodeList != nil {
		if bytes, err := json.MarshalIndent(nodeList.Nodes, "", "  "); err == nil {
			_, err := WriteAstFileHandle.Write(bytes)
			if err != nil {
				panic(err)
			}

			fmt.Printf("AST is writed to %s\n", WriteAstFileHandle.Name())
		} else {
			panic(err)
		}
	}

	if ParseAst {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			switch e := err.(type) {
			case checker.Error:
				reportError(&e)

			case []checker.Error:
				for i := range e {
					reportError(&e[i])
				}

			default:
				panic(err)
			}
		}
	}()

	if isRepl {
		decl := &ast.FuncDecl{
			Name: &ast.Ident{Name: "repl"},
			Body: &ast.CurlyList{List: nodeList},
		}
		checker.NewFunc(nil, nil, nil, decl)
	} else {
		mod := &ast.ModuleDecl{
			Name: &ast.Ident{Name: cfg.Files[config.MainFileID].Name},
			Body: nodeList,
		}
		typeinfo, errs := checker.Check(mod)
		_ = typeinfo

		if len(errs) != 0 {
			for _, err := range errs {
				reportError(err)
			}
			return
		}
	}
}
