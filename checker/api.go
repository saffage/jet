package checker

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/scanner"
	"github.com/saffage/jet/text"
)

var ErrorEmptyFileBuf = errors.New("empty file buffer or invalid file ID")

func Check(file *text.File, stmts *ast.StmtList) (*Module, error) {
	moduleName := file.Name
	report.Hint("checking module '%s'", moduleName)

	module := NewModule(NewScope(Global, "module "+moduleName), moduleName, stmts)
	check := &Checker{
		module: module,
		scope:  module.Scope,
		errors: make([]error, 0),
		file:   file,
	}

	visitor := ast.Visitor(check.visit)

	for _, node := range stmts.Nodes {
		visitor.WalkTopDown(node)
	}

	module.completed = true

	// TODO move to separate cli command
	// if cfg.Flags.DumpCheckerState {
	// 	err := os.Mkdir(cfg.Options.CacheDir, os.ModePerm)
	// 	if err != nil && !os.IsExist(err) {
	// 		panic(err)
	// 	}
	//
	// 	f, err := os.Create(filepath.Join(cfg.Options.CacheDir, "checker-state.txt"))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	defer f.Close()
	// 	report.HintX("checker", "dumping checker state")
	// 	spew.Fdump(f, check)
	// }

	return check.module, report.Join(check.errors...)
}

func CheckFile(file *text.File) (*Module, error) {
	scannerFlags := scanner.SkipWhitespace | scanner.SkipComments
	parserFlags := parser.DefaultFlags

	if config.TraceParser {
		parserFlags |= parser.Trace
	}

	tokens, err := scanner.Scan(file.Content, file.ID, scannerFlags)
	if err != nil {
		return nil, err
	}

	stmts, err := parser.Parse(tokens, parserFlags)
	if err != nil {
		return nil, err
	}
	if stmts == nil {
		// Empty file, nothing to check.
		return NewModule(NewScope(nil, "module "+file.Name), file.Name, nil), nil
	}

	// TODO move to separate cli command
	// if config.ParseAst {
	// 	printRecreatedAST(stmts)
	// 	return NewModule(NewScope(nil, "module "+file.Name), file.Name, nil), nil
	// }

	return Check(file, stmts)
}

func printRecreatedAST(nodeList *ast.StmtList) {
	fmt.Println("recreated AST:")

	for i, node := range nodeList.Nodes {
		if _, isEmpty := node.(*ast.Empty); i < len(nodeList.Nodes)-1 || !isEmpty {
			fmt.Println(color.HiGreenString(node.Repr()))
		}
	}
}
