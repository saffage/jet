package jet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/cgen"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/scanner"
)

func process(
	cfg *config.Config,
	buffer []byte,
	fileID config.FileID,
) {
	toks, scanErrors := scanner.Scan(buffer, fileID, scanner.SkipWhitespace|scanner.SkipComments)
	if len(scanErrors) > 0 {
		report.Report(scanErrors...)
		return
	}

	parserFlags := parser.DefaultFlags

	if TraceParser {
		parserFlags |= parser.Trace
	}

	nodeList, parseErrors := parser.Parse(cfg, toks, parserFlags)
	if len(parseErrors) > 0 {
		report.Report(parseErrors...)
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
			switch err := err.(type) {
			case *checker.Error:
				report.Report(err)

			case []*checker.Error:
				for _, err := range err {
					report.Report(err)
				}

			default:
				panic(err)
			}
		}
	}()

	mod := &ast.ModuleDecl{
		Name: &ast.Ident{Name: cfg.Files[config.MainFileID].Name},
		Body: nodeList,
	}
	typeinfo, errs := checker.Check(cfg, fileID, mod)
	_ = typeinfo

	if len(errs) != 0 {
		report.Report(errs...)
		return
	}

	if GenC {
		finfo := cfg.Files[fileID]
		dir := filepath.Dir(finfo.Path)
		f, err := os.Create(filepath.Join(dir, "out.c"))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		errs = cgen.Generate(f, typeinfo)

		if len(errs) != 0 {
			report.Report(errs...)
			return
		}
	}
}
