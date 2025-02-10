package checker

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
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

func CheckBuiltInPkgs(cfg *config.Config) error {
	var err error
	once.Do(func() { err = checkBuiltInPkgsAux(cfg) })
	return err
}

var once sync.Once

func checkBuiltInPkgsAux(cfg *config.Config) error {
	if cfg.Flags.NoCoreLib {
		return nil
	}

	report.Hint("checking package 'core'")

	var libDir string

	if cfg.Options.CoreLibPath != "" {
		libDir = filepath.Clean(cfg.Options.CoreLibPath)
	} else {
		compilerDir := filepath.Dir(cfg.Exe)
		libDir = filepath.Join(compilerDir, "lib")
	}

	if dir, err := os.Stat(libDir); errors.Is(err, fs.ErrNotExist) ||
		(dir != nil && !dir.IsDir()) {

		return report.Join(ErrWhileCheckingPackageCore, ErrPackageWasNotFound)
	}

	corePkgDir := filepath.Join(libDir, "core")

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
				info.Hints = append(info.Hints, report.HintInfo{
					Message: "package 'core' have fixed file set that needs to be checked before anything else",
				})

				errs[0] = info
				// errs = append([]error{info}, errs...)
			}

			info.Hints = append(info.Hints, report.HintInfo{
				Message: fmt.Sprintf("module '%s' was not found in package 'core'", name),
			})
		} else {
			content, err := os.ReadFile(path)

			if err != nil {
				errs = append(errs, report.Wrapf(err, "failed to read file: %s", path))
				continue
			}

			file := config.Global.NewFile()
			file.Name = name
			file.Path = path
			file.Buf = bytes.NewBuffer(content)

			module, err := CheckFile(cfg, file.ID)

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

	if len(errs) > 1 || errs[0] != nil {
		return report.Join(append([]error{ErrWhileCheckingPackageCore}, errs...)...)
	}

	return nil
}
