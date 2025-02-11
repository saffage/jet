package checker

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

var (
	ErrWhileCheckingPackageCore = errors.New("while checking package 'builtin'")
	ErrPackageWasNotFound       = errors.New("package was not found")
	ErrInvalidPackagePath       = errors.New("invalid package path")
)

var (
	builtinFilename = "builtin"
	cFilename       = "cc"
)

var coreFilenames = [...]string{
	builtinFilename,
	cFilename,
}

var coreModules = map[string]*Module{
	builtinFilename: nil,
}

// This module contains the declaration of the Jet built-in types.
// var ModuleBuiltin *Module = NewModule(NewScope(nil, "module builtin"), builtinFilename, nil)

// This module contains C type declarations and other tools for
// interacting with the C backend.
// var ModuleC *Module = NewModule(NewScope(nil, "module c"), "c", nil)

var once sync.Once

func CheckBuiltInPackage() error {
	var err error
	once.Do(func() { err = checkBuiltInPackage() })
	return err
}

func checkBuiltInPackage() error {
	if config.NoBuiltinPackage {
		return nil
	}

	report.Hint("checking package 'core'")

	compilerPath := filepath.Dir(config.CompilerFilepath)
	builtinPackagePath := filepath.Join(compilerPath, config.BuiltinPackagePath)

	if dir, err := os.Stat(builtinPackagePath); errors.Is(err, fs.ErrNotExist) ||
		(dir != nil && !dir.IsDir()) {

		return report.Join(ErrWhileCheckingPackageCore, ErrPackageWasNotFound)
	}

	corePkgDir := filepath.Join(builtinPackagePath, "core")

	if dir, err := os.Stat(corePkgDir); errors.Is(err, fs.ErrNotExist) ||
		(dir != nil && !dir.IsDir()) {

		return report.Join(ErrWhileCheckingPackageCore, ErrPackageWasNotFound)
	}

	corePkgFiles, err := os.ReadDir(corePkgDir)

	if err != nil {
		return errors.Join(errors.New("while reading package 'core'"), err)
	}

	files := map[string]string{}

	for _, module := range coreFilenames {
		files[module] = ""
	}

	for _, entry := range corePkgFiles {
		if entry.Type().IsRegular() {
			name := entry.Name()
			ext := filepath.Ext(name)
			base := name[:len(name)-len(ext)]

			if ext == ".jet" {
				if slices.Contains(coreFilenames[:], base) {
					files[base] = filepath.Join(corePkgDir, name)
				}
			}
		}
	}

	// FIXME test it
	var info *report.Info
	var errs = []error{nil}

	for name, path := range files {
		if path == "" {
			if info == nil {
				info = &report.Info{Title: "missing 'core' package files"}
				errs[0] = info
			}

			info.Hints = append(info.Hints, report.HintInfo{
				Message: fmt.Sprintf("module '%s' was not found in package 'core'", name),
			})
		} else {
			file, err := config.ReadFile(path)

			if err != nil {
				errs = append(errs, report.Wrapf(err, "failed to read file: %s", path))
				continue
			}

			module, err := CheckFile(file)

			if err != nil {
				errs = append(errs, report.Wrapf(err, "failed to check file: %s", path))
				continue
			}

			coreModules[name] = module

			if name == builtinFilename {
				for _, sym := range module.Scope.symbols {
					_ = Global.Define(sym)
				}
			}
		}
	}

	if info != nil && len(info.Hints) > 0 {
		info.Hints = append(info.Hints, report.HintInfo{
			Message:    "package 'core' have fixed file set that needs to be checked before anything else",
			Suggestion: strings.Join(coreFilenames[:], "\n"),
		})
	}

	if len(errs) > 1 || errs[0] != nil {
		return report.Join(append([]error{ErrWhileCheckingPackageCore}, errs...)...)
	}

	return nil
}
