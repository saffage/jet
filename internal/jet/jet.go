package jet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/cgen"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/scanner"
)

func CheckFile(cfg *config.Config, fileID config.FileID) {
	scannerFlags := scanner.SkipWhitespace | scanner.SkipComments
	parserFlags := parser.DefaultFlags

	if config.FlagTraceParser {
		parserFlags |= parser.Trace
	}

	fi := cfg.Files[fileID]
	if fi.Buf == nil {
		return
	}

	tokens, err := scanner.Scan(fi.Buf.Bytes(), fileID, scannerFlags)
	if err != nil {
		report.Errors(err)
		return
	}

	spew.Dump(tokens)

	stmts, err := parser.Parse(cfg, tokens, parserFlags)
	if err != nil {
		report.Errors(err)
		return
	}
	if stmts == nil {
		// Empty file, nothing to check.
		return
	}

	printRecreatedAST(stmts)
	fmt.Println("memory dump:")
	spew.Dump(stmts)

	if config.FlagParseAst {
		printRecreatedAST(stmts)
		return
	}

	_, err = checker.Check(cfg, fileID, stmts)
	if err != nil {
		report.Errors(err)
		return
	}
}

func printRecreatedAST(nodeList *ast.StmtList) {
	fmt.Println("recreated AST:")

	for i, node := range nodeList.Nodes {
		if _, isEmpty := node.(*ast.Empty); i < len(nodeList.Nodes)-1 || !isEmpty {
			fmt.Println(color.HiGreenString(node.Repr()))
		}
	}
}

func process(cfg *config.Config, fileID config.FileID) bool {
	checker.CheckBuiltInPkgs()

	m, err := checker.CheckFile(cfg, fileID)
	if err != nil {
		report.Errors(err)
		return false
	}

	if config.FlagGenC {
		finfo := cfg.Files[fileID]
		dir := filepath.Dir(finfo.Path)

		err := os.Mkdir(filepath.Join(dir, ".jet"), os.ModePerm)
		if err != nil && !os.IsExist(err) {
			report.Errors(err)
			return false
		}

		dir = filepath.Join(dir, ".jet")

		for _, mImported := range m.Imports {
			genCFile(mImported, dir)
		}

		return genCFile(m, dir)
	}

	return true
}

func genCFile(m *checker.Module, dir string) bool {
	filename := filepath.Join(dir, m.Name()+"__jet.c")
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	report.Hintf("generating module '%s'", m.Name())
	report.TaggedDebugf("gen", "module file is '%s'", filename)

	if err := cgen.Generate(f, m); err != nil {
		report.Errors(err)
		return false
	}

	return true
}
