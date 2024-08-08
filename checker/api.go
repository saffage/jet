package checker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/scanner"
)

var ErrorEmptyFileBuf = errors.New("empty file buffer or invalid file ID")

func Check(cfg *config.Config, fileID config.FileID, stmts *ast.StmtList) (*Module, error) {
	moduleName := cfg.Files[fileID].Name
	report.Hintf("checking module '%s'", moduleName)

	module := NewModule(NewScope(Global, "module "+moduleName), moduleName, stmts)
	check := &Checker{
		module: module,
		scope:  module.Scope,
		errors: make([]error, 0),
		cfg:    cfg,
		fileID: fileID,
	}

	visitor := ast.Visitor(check.visit)

	for _, node := range stmts.Nodes {
		visitor.WalkTopDown(node)
	}

	module.completed = true

	if cfg.Flags.DumpCheckerState {
		err := os.Mkdir(cfg.Options.CacheDir, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		f, err := os.Create(filepath.Join(cfg.Options.CacheDir, "checker-state.txt"))
		if err != nil {
			panic(err)
		}

		defer f.Close()
		report.TaggedHintf("checker", "dumping checker state")
		spew.Fdump(f, check)
	}

	return check.module, errors.Join(check.errors...)
}

func CheckFile(cfg *config.Config, fileID config.FileID) (*Module, error) {
	scannerFlags := scanner.SkipWhitespace | scanner.SkipComments
	parserFlags := parser.DefaultFlags

	if cfg.Flags.TraceParser {
		parserFlags |= parser.Trace
	}

	fi := cfg.Files[fileID]
	if fi.Buf == nil {
		return nil, ErrorEmptyFileBuf
	}

	tokens, err := scanner.Scan(fi.Buf.Bytes(), fileID, scannerFlags)
	if err != nil {
		return nil, err
	}

	stmts, err := parser.Parse(tokens, parserFlags)
	if err != nil {
		return nil, err
	}
	if stmts == nil {
		// Empty file, nothing to check.
		return NewModule(NewScope(nil, "module "+fi.Name), fi.Name, nil), nil
	}

	if cfg.Flags.ParseAst {
		printRecreatedAST(stmts)
		return NewModule(NewScope(nil, "module "+fi.Name), fi.Name, nil), nil
	}

	return Check(cfg, fileID, stmts)
}

func printRecreatedAST(nodeList *ast.StmtList) {
	fmt.Println("recreated AST:")

	for i, node := range nodeList.Nodes {
		if _, isEmpty := node.(*ast.Empty); i < len(nodeList.Nodes)-1 || !isEmpty {
			fmt.Println(color.HiGreenString(node.Repr()))
		}
	}
}
