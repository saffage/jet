//go:build ignore

package types

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

var ErrInvalidPkgPath = errors.New("invalid path to the package")

func (check *checker) resolveImport(node *ast.Import) {
	path := check.resolveImportPath(node.Module)
	if path == "" {
		check.errorf(node.Module, "cannot find module named '%s'", node.Module)
		return
	}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		check.errorf(node.Module, "while reading file: %s", err.Error())
	}

	fileID := config.NextFileID()
	check.cfg.Files[fileID] = config.FileInfo{
		Name: node.Module.Data,
		Path: path,
		Buf:  bytes.NewBuffer(fileContent),
	}

	m, err := CheckFile(check.cfg, fileID)
	if err != nil {
		report.Errors(err)
		check.errorf(node.Module, "the module check was finished with errors")
	}

	if defined := check.module.Env.Define(m); defined != nil {
		check.problem(errorAlreadyDefined(node.Module, defined.Ident()))
		return
	}
	check.module.Imports = append(check.module.Imports, m)
	check.newDef(node.Module, m)
}

func (check *checker) resolveImportPath(ident *ast.Lower) string {
	modulePath := ""
	dir := filepath.Dir(check.cfg.Files[check.module.file.ID].Path)
	err := filepath.Walk(dir, makeWalkFunc(dir, ident.Data, &modulePath))
	if err != nil {
		check.errorf(ident, "while walking dir: %s", err.Error())
		return ""
	}
	return modulePath
}

func makeWalkFunc(root, expectedName string, result *string) filepath.WalkFunc {
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

			report.TaggedDebugf("importer", "found file: '%s'", path)
			return filepath.SkipAll
		}

		return nil
	}
}
