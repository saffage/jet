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

	tokens, errs := scanner.Scan(fi.Buf.Bytes(), fileID, scannerFlags)
	if len(errs) > 0 {
		report.Errors(errs...)
		return
	}

	spew.Dump(tokens)

	stmts, errs := parser.Parse(cfg, tokens, parserFlags)
	if len(errs) > 0 {
		report.Errors(errs...)
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

	_, errs = checker.Check(cfg, fileID, stmts)
	if len(errs) > 0 {
		report.Errors(errs...)
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
	// CheckFile(cfg, fileID)
	checker.CheckBuiltInPkgs()

	m, errs := checker.CheckFile(cfg, fileID)
	// tree, errs := parser.Parse(cfg, fileID)
	if len(errs) != 0 {
		report.Errors(errs...)
		return false
	}

	if config.FlagGenC {
		finfo := cfg.Files[fileID]
		dir := filepath.Dir(finfo.Path)

		err := os.Mkdir(filepath.Join(dir, ".jet"), os.ModePerm)
		if err != nil && !os.IsExist(err) {
			panic(err)
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

	if errs := cgen.Generate(f, m); len(errs) != 0 {
		report.Errors(errs...)
		return false
	}

	return true
}
