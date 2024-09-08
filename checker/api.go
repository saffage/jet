package checker

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/scanner"
)

var ErrEmptyFileBuf = errors.New("empty file buffer or invalid file ID")

func Check(cfg *config.Config, fileID config.FileID, stmts *ast.Stmts) (*Module, error) {
	var (
		moduleName = cfg.Files[fileID].Name
		scope      = NewScope(Global, "module "+moduleName)
		module     = NewModule(scope, moduleName, stmts)
		check      = &checker{
			module: module,
			scope:  scope,
			errs:   make([]error, 0),
			cfg:    cfg,
			fileID: fileID,
		}
	)

	report.Hintf("checking module '%s'", moduleName)

	for _, stmt := range stmts.Nodes {
		ast.WalkTopDown(stmt, check)
	}

	module.completed = true

	// TODO: remove it?
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

	return check.module, errors.Join(check.errs...)
}

func CheckFile(cfg *config.Config, fileID config.FileID) (*Module, error) {
	scannerFlags := scanner.SkipComments
	parserFlags := parser.DefaultFlags

	fi := cfg.Files[fileID]
	if fi.Buf == nil {
		return nil, ErrEmptyFileBuf
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

	return Check(cfg, fileID, stmts)
}
