package types

import (
	"errors"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/parser/scanner"
	"github.com/saffage/jet/report"
)

var ErrEmptyFileBuf = errors.New("empty file buffer or invalid file ID")

func Check(cfg *config.Config, file ast.File) (*Module, error) {
	InitModuleCore(cfg)

	var (
		moduleName = cfg.Files[file.ID].Name
		scope      = NewNamedEnv(ModuleCore.Env, "module "+moduleName)
		module     = NewModule(scope, moduleName, file)
		check      = &checker{
			module: module,
			env:    scope,
			cfg:    cfg,
		}
	)

	_ = scope.Use(ModuleCore.Env)
	_ = scope.UseTypes(ModuleCore.Env)

	report.Hint("checking module '%s'", moduleName)
	for _, stmt := range module.file.Ast.Nodes {
		ast.WalkTopDown(stmt, check)
	}

	module.completed = true
	return check.module, errors.Join(check.problems...)
}

func CheckFile(cfg *config.Config, fileID config.FileID) (*Module, error) {
	fi := cfg.Files[fileID]

	if fi.Buf == nil {
		return nil, ErrEmptyFileBuf
	}

	tokens, err := scanner.Scan(fi.Buf.Bytes(), fileID, scanner.SkipComments)

	if err != nil {
		return nil, err
	}

	stmts, err := parser.Parse(tokens, parser.DefaultFlags)

	if err != nil {
		return nil, err
	}

	file := ast.File{Ast: stmts, ID: fileID}

	if stmts == nil {
		// Empty file, nothing to check.
		return NewModule(NewNamedEnv(nil, "module "+fi.Name), fi.Name, file), nil
	}

	return Check(cfg, file)
}
