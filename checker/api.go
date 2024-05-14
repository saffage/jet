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

func Check(cfg *config.Config, fileID config.FileID, node *ast.ModuleDecl) (*Module, []error) {
	report.Hintf("checking module '%s'", cfg.Files[fileID].Name)

	module := NewModule(NewScope(Global), node)
	check := &Checker{
		module:         module,
		scope:          module.Scope,
		errors:         make([]error, 0),
		isErrorHandled: true,
		cfg:            cfg,
		fileID:         fileID,
	}

	{
		nodes := []ast.Node(nil)

		switch body := node.Body.(type) {
		case *ast.List:
			nodes = body.Nodes

		case *ast.CurlyList:
			nodes = body.List.Nodes

		default:
			panic("ill-formed AST")
		}

		for _, node := range nodes {
			ast.WalkTopDown(check.visit, node)
		}

		module.completed = true
	}

	{
		f, err := os.Create("_test/checker_state.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		spew.Fdump(f, check)
	}

	return check.module, check.errors
}

func CheckFile(cfg *config.Config, fileID config.FileID) (*Module, []error) {
	const ScannerFlags = scanner.SkipWhitespace | scanner.SkipComments
	const ParserFlags = parser.DefaultFlags

	fi := cfg.Files[fileID]
	if fi.Buf == nil {
		return nil, []error{ErrorEmptyFileBuf}
	}

	tokens, errs := scanner.Scan(fi.Buf.Bytes(), fileID, ScannerFlags)
	if len(errs) > 0 {
		return nil, errs
	}

	nodeList, errs := parser.Parse(cfg, tokens, ParserFlags)
	if len(errs) > 0 {
		return nil, errs
	}
	if nodeList == nil {
		// Empty file, nothing to check.
		return NewModule(NewScope(nil), nil), nil
	}

	// printRecreatedAST(nodeList)

	return Check(cfg, fileID, &ast.ModuleDecl{
		Name: &ast.Ident{Name: fi.Name},
		Body: nodeList,
	})
}

func printRecreatedAST(nodeList *ast.List) {
	fmt.Println("recreated AST:")

	for i, node := range nodeList.Nodes {
		if _, isEmpty := node.(*ast.Empty); i < len(nodeList.Nodes)-1 || !isEmpty {
			fmt.Println(color.HiGreenString(node.String()))
		}
	}
}
