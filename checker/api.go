package checker

import (
	"errors"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/scanner"
)

var ErrorEmptyFileBuf = errors.New("empty file buffer or invalid file ID")

func Check(cfg *config.Config, fileID config.FileID, stmts *ast.StmtList) (*Module, []error) {
	moduleName := cfg.Files[fileID].Name
	report.Hintf("checking module '%s'", moduleName)

	module := NewModule(NewScope(Global, "module "+moduleName), moduleName, stmts)
	check := &Checker{
		module:         module,
		scope:          module.Scope,
		errors:         make([]error, 0),
		isErrorHandled: true,
		cfg:            cfg,
		fileID:         fileID,
	}

	visitor := ast.Visitor(check.visit)

	for _, node := range stmts.Nodes {
		visitor.WalkTopDown(node)
	}

	module.completed = true

	if config.FlagDumpCheckerState {
		err := os.Mkdir(".jet", os.ModePerm)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		f, err := os.Create(".jet/checker_state.txt")
		if err != nil {
			panic(err)
		}

		defer f.Close()
		report.TaggedHintf("checker", "dumping checker state")
		spew.Fdump(f, check)
	}

	return check.module, check.errors
}

func CheckFile(cfg *config.Config, fileID config.FileID) (*Module, []error) {
	scannerFlags := scanner.SkipWhitespace | scanner.SkipComments
	parserFlags := parser.DefaultFlags

	if config.FlagTraceParser {
		parserFlags |= parser.Trace
	}

	fi := cfg.Files[fileID]
	if fi.Buf == nil {
		return nil, []error{ErrorEmptyFileBuf}
	}

	tokens, errs := scanner.Scan(fi.Buf.Bytes(), fileID, scannerFlags)
	if len(errs) > 0 {
		return nil, errs
	}

	stmts, errs := parser.Parse(cfg, tokens, parserFlags)
	if len(errs) > 0 {
		return nil, errs
	}
	if stmts == nil {
		// Empty file, nothing to check.
		return NewModule(NewScope(nil, "module "+fi.Name), fi.Name, nil), nil
	}

	if config.FlagParseAst {
		printRecreatedAST(stmts)
		return nil, nil
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
