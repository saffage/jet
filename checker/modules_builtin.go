package checker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

var ErrInvalidPkgPath = errors.New("invalid path to the package")

// This module contains the declaration of the Jet built-in types.
var ModuleBuiltin *Module = NewModule(NewScope(nil, "module builtin"), "builtin", nil)

// This module contains C type declarations and other tools for
// interacting with the C backend.
var ModuleC *Module = NewModule(NewScope(nil, "module c"), "c", nil)

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

	report.TaggedHintf("checker", "checking package 'core'")

	var libDir string

	if cfg.Options.CoreLibPath != "" {
		libDir = filepath.Clean(cfg.Options.CoreLibPath)
	} else {
		compilerDir := filepath.Dir(cfg.Exe)
		libDir = filepath.Join(compilerDir, "lib")
	}

	if dir, err := os.Stat(libDir); os.IsNotExist(err) || (dir != nil && !dir.IsDir()) {
		return fmt.Errorf("invalid path to the core package: '%s'", libDir)
	}

	corePkgDir := filepath.Join(libDir, "core")

	if _, err := os.Stat(corePkgDir); os.IsNotExist(err) {
		return errors.New("package 'core' was not found")
	}

	corePkgFiles, err := os.ReadDir(corePkgDir)
	if err != nil {
		return errors.Join(errors.New("while reading package 'core'"), err)
	}

	var builtinModuleFilepath, cModuleFilepath string

	for _, entry := range corePkgFiles {
		switch entry.Name() {
		case "builtin.jet":
			builtinModuleFilepath = filepath.Join(corePkgDir, entry.Name())

		case "c.jet":
			cModuleFilepath = filepath.Join(corePkgDir, entry.Name())

		default:
			return fmt.Errorf("unexpected file in package 'builtin': '%s'", entry.Name())
		}
	}

	switch {
	case builtinModuleFilepath == "":
		return errors.New("module 'builtin' was not found")

	case cModuleFilepath == "":
		return errors.New("module 'c' was not found")
	}

	builtinModuleContent, err := os.ReadFile(builtinModuleFilepath)
	if err != nil {
		return err
	}

	cModuleContent, err := os.ReadFile(cModuleFilepath)
	if err != nil {
		return err
	}

	builtinFileID := config.NextFileID()
	cFileID := config.NextFileID()
	cfg.Files[builtinFileID] = config.FileInfo{
		Name: "Types",
		Path: builtinModuleFilepath,
		Buf:  bytes.NewBuffer(builtinModuleContent),
	}
	cfg.Files[cFileID] = config.FileInfo{
		Name: "C",
		Path: cModuleFilepath,
		Buf:  bytes.NewBuffer(cModuleContent),
	}

	ModuleBuiltin, err = CheckFile(cfg, builtinFileID)
	if err != nil {
		report.TaggedErrorf("internal", "while checking package 'builtin'")
		return err
	}

	// ModuleC, err = CheckFile(config.Global, cFileID)
	// if err != nil {
	// 	report.TaggedErrorf("internal", "while checking package 'builtin'")
	// 	report.Errors(err)
	// 	os.Exit(1)
	// }

	for _, sym := range ModuleBuiltin.Scope.symbols {
		_ = Global.Define(sym)
	}

	return nil
}
