package checker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
)

// This module contains the declaration of the Jet built-in types.
var ModuleTypes *Module = NewModule(NewScope(nil), nil)

// This module contains C type declarations and other tools for
// interacting with the C backend.
var ModuleC *Module = NewModule(NewScope(nil), nil)

func init() {
	var libDir string

	if config.FlagCoreLibPath != "" {
		libDir = filepath.Clean(config.FlagCoreLibPath)
	} else {
		compilerDir := filepath.Dir(config.Exe)
		libDir = filepath.Join(compilerDir, "lib")
	}

	if dir, err := os.Stat(libDir); os.IsNotExist(err) || !dir.IsDir() {
		panic(fmt.Errorf("invalid path to the core library: '%s'", libDir))
	}

	builtinPkgDir := filepath.Join(libDir, "builtin")

	if _, err := os.Stat(builtinPkgDir); os.IsNotExist(err) {
		panic("package 'builtin' was not found")
	}

	builtInFiles, err := os.ReadDir(builtinPkgDir)
	if err != nil {
		panic(errors.Join(errors.New("while reading package 'builtin'"), err))
	}

	var ModuleTypesFilepath, ModuleCFilepath string

	for _, entry := range builtInFiles {
		switch entry.Name() {
		case "Types.jet":
			ModuleTypesFilepath = filepath.Join(builtinPkgDir, entry.Name())

		case "C.jet":
			ModuleCFilepath = filepath.Join(builtinPkgDir, entry.Name())

		default:
			panic(fmt.Sprintf("unexpected file in package 'builtin': '%s'", entry.Name()))
		}
	}

	switch {
	case ModuleTypesFilepath == "":
		panic("module 'Types' was not found")

	case ModuleCFilepath == "":
		panic("module 'C' was not found")
	}

	moduleTypesContent, err := os.ReadFile(ModuleTypesFilepath)
	if err != nil {
		panic(err)
	}

	moduleCContent, err := os.ReadFile(ModuleCFilepath)
	if err != nil {
		panic(err)
	}

	moduleTypesFileID := config.NextFileID()
	moduleCFileID := config.NextFileID()
	config.Global.Files[moduleTypesFileID] = config.FileInfo{
		Name: "Types",
		Path: ModuleTypesFilepath,
		Buf:  bytes.NewBuffer(moduleTypesContent),
	}
	config.Global.Files[moduleCFileID] = config.FileInfo{
		Name: "C",
		Path: ModuleCFilepath,
		Buf:  bytes.NewBuffer(moduleCContent),
	}

	var errs []error

	ModuleTypes, errs = CheckFile(config.Global, moduleTypesFileID)
	checkErrors(errs)

	// ModuleC, errs = CheckFile(config.Global, moduleCFileID)
	// checkErrors(errs)

	for _, sym := range ModuleTypes.Scope.symbols {
		_ = Global.Define(sym)
	}
}

func checkErrors(errs []error) {
	if len(errs) != 0 {
		report.TaggedErrorf("internal", "while checking package 'builtin'")
		report.Report(errs...)
		os.Exit(1)
	}
}
