package checker

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

func (check *Checker) resolveImport(node *ast.Import) {
	path := check.resolveImportPath(node.Module)
	if path == "" {
		check.errorf(node.Module, "cannot find module named '%s'", node.Module)
		return
	}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		check.errorf(node.Module, "while reading file: %s", err.Error())
	}

	file := config.Global.NewFile()
	file.Name = node.Module.Name
	file.Path = path
	file.Buf = bytes.NewBuffer(fileContent)

	m, err := CheckFile(check.cfg, file.ID)
	if err != nil {
		report.Report(config.Global, err)
		check.errorf(node.Module, "the module check was finished with errors")
	}

	if defined := check.module.Scope.Define(m); defined != nil {
		check.addError(errorAlreadyDefined(node.Module, defined.Ident()))
		return
	}
	check.module.Imports = append(check.module.Imports, m)
	check.newDef(node.Module, m)
}

func (check *Checker) resolveImportPath(ident *ast.Ident) string {
	modulePath := ""
	dir := filepath.Dir(check.cfg.File(check.fileID).Path)
	err := filepath.Walk(dir, makeWalkFunc(dir, ident.Name, &modulePath))
	if err != nil {
		check.errorf(ident, "while walking dir: %s", err.Error())
		return ""
	}
	return modulePath
}

func makeWalkFunc(root string, expectedName string, result *string) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			if path != "." && path != root {
				return fs.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		name := filepath.Base(path[:len(path)-len(ext)])

		if name == expectedName && ext == ".jet" {
			if result != nil {
				*result = path
			}

			report.DebugX("importer", "found file: '%s'", path)
			return filepath.SkipAll
		}

		return nil
	}
}
