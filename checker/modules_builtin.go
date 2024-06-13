package checker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
)

// This module contains the declaration of the Jet built-in types.
var ModuleBuiltin *Module = NewModule(NewScope(nil, "module builtin"), "builtin", nil)

// This module contains C type declarations and other tools for
// interacting with the C backend.
var ModuleC *Module = NewModule(NewScope(nil, "module c"), "c", nil)

var CheckBuiltInPkgs = sync.OnceFunc(func() {
	if config.FlagNoCoreLib {
		return
	}

	report.Hintf("checking package 'core'")

	var libDir string

	if config.FlagCoreLibPath != "" {
		libDir = filepath.Clean(config.FlagCoreLibPath)
	} else {
		compilerDir := filepath.Dir(config.Exe)
		libDir = filepath.Join(compilerDir, "lib")
	}

	if dir, err := os.Stat(libDir); os.IsNotExist(err) || (dir != nil && !dir.IsDir()) {
		panic(fmt.Errorf("invalid path to the core package: '%s'", libDir))
	}

	corePkgDir := filepath.Join(libDir, "core")

	if _, err := os.Stat(corePkgDir); os.IsNotExist(err) {
		panic("package 'core' was not found")
	}

	corePkgFiles, err := os.ReadDir(corePkgDir)
	if err != nil {
		panic(errors.Join(errors.New("while reading package 'core'"), err))
	}

	var builtinModuleFilepath, cModuleFilepath string

	for _, entry := range corePkgFiles {
		switch entry.Name() {
		case "builtin.jet":
			builtinModuleFilepath = filepath.Join(corePkgDir, entry.Name())

		case "c.jet":
			cModuleFilepath = filepath.Join(corePkgDir, entry.Name())

		default:
			panic(fmt.Sprintf("unexpected file in package 'builtin': '%s'", entry.Name()))
		}
	}

	switch {
	case builtinModuleFilepath == "":
		panic("module 'builtin' was not found")

	case cModuleFilepath == "":
		panic("module 'c' was not found")
	}

	builtinModuleContent, err := os.ReadFile(builtinModuleFilepath)
	if err != nil {
		panic(err)
	}

	cModuleContent, err := os.ReadFile(cModuleFilepath)
	if err != nil {
		panic(err)
	}

	builtinFileID := config.NextFileID()
	cFileID := config.NextFileID()
	config.Global.Files[builtinFileID] = config.FileInfo{
		Name: "Types",
		Path: builtinModuleFilepath,
		Buf:  bytes.NewBuffer(builtinModuleContent),
	}
	config.Global.Files[cFileID] = config.FileInfo{
		Name: "C",
		Path: cModuleFilepath,
		Buf:  bytes.NewBuffer(cModuleContent),
	}

	ModuleBuiltin, err = CheckFile(config.Global, builtinFileID)
	if err != nil {
		report.TaggedErrorf("internal", "while checking package 'builtin'")
		report.Errors(err)
		os.Exit(1)
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
})
