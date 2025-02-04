package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/saffage/jet/cgen"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/token"
	"github.com/urfave/cli/v2"
)

func Build(cfg *config.Config, file *config.File) error {
	report.Debug("set file '%s' as main module", file.Path)

	if err := internalBuild(cfg, file); err != nil {
		return err
	}

	if err := compileToC(cfg, filepath.Dir(file.Path), file.Name); err != nil {
		return err
	}

	if cfg.Flags.Run {
		exePath := "." + string(filepath.Separator) + file.Name

		if runtime.GOOS == "windows" {
			exePath += ".exe"
		}

		report.Hint("running: '%s'", exePath)

		cmd := exec.Command(exePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			wd, _ := os.Getwd()
			report.Hint(wd)
			report.ErrorX("run", "%s", err.Error())
		}
	}

	return nil
}

func beforeBuild(ctx *cli.Context) error {
	config.Global.Flags.Run = ctx.Bool("run")
	config.Global.Flags.DumpCheckerState = ctx.Bool("dump-checker-state")
	config.Global.Flags.ParseAst = ctx.Bool("parse-ast")
	config.Global.Flags.TraceParser = ctx.Bool("trace-parser")
	config.Global.Options.CC = ctx.String("cc")
	config.Global.Options.CCFlags = ctx.String("cc-flags")
	config.Global.Options.LDFlags = ctx.String("ld-flags")
	return nil
}

func actionBuild(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return errors.New("expected path to a file")
	}

	if ctx.Args().Len() != 1 {
		return errors.New("invalid arguments count (expected 1)")
	}

	path := filepath.Clean(ctx.Args().Get(0))
	name, data, err := readFile(path)
	if err != nil {
		return err
	}

	mainFile := config.Global.NewFile()
	mainFile.Name = name
	mainFile.Path = path
	mainFile.Buf = bytes.NewBuffer(data)

	return Build(config.Global, mainFile)
}

func internalBuild(cfg *config.Config, file *config.File) error {
	if err := checker.CheckBuiltInPkgs(cfg); err != nil {
		return err
	}

	m, err := checker.CheckFile(cfg, file.ID)
	if err != nil {
		return err
	}

	dir := filepath.Join(filepath.Dir(file.Path), cfg.Options.CacheDir)
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	for _, importedModule := range m.Imports {
		if err := genModule(importedModule, dir); err != nil {
			return err
		}
	}

	return genModule(m, dir)
}

func genModule(m *checker.Module, dir string) error {
	filename := filepath.Join(dir, m.Name()+".c")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	report.Hint("generating module '%s'", m.Name())
	report.DebugX("gen", "module file is '%s'", filename)

	if err := cgen.Generate(f, m); err != nil {
		return err
	}

	return nil
}

func readFile(path string) (name string, data []byte, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		return
	}

	if !stat.Mode().IsRegular() {
		err = fmt.Errorf("'%s' is not a file", path)
		return
	}

	fileExt := filepath.Ext(path)

	if fileExt != ".jet" {
		err = fmt.Errorf("expected file extension '.jet', got '%s' instead", fileExt)
		return
	}

	name = filepath.Base(path[:len(path)-len(fileExt)])
	if _, err = token.IsValidIdent(name); err != nil {
		err = errors.Join(fmt.Errorf("filename is not a valid identifier"), err)
		return
	}

	data, err = os.ReadFile(path)
	if err != nil {
		err = errors.Join(fmt.Errorf("while reading file '%s'", path), err)
		return
	}

	return
}

func compileToC(cfg *config.Config, dir, moduleName string) error {
	file := filepath.Join(dir, cfg.Options.CacheDir, moduleName+".c")
	args := []string{"-o", moduleName, file}

	if len(cfg.Options.LDFlags) > 0 {
		args = append(args, strings.Split(cfg.Options.LDFlags, " ")...)
	}

	cmd := exec.Command(cfg.Options.CC, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	report.HintX("cc", "%s", cmd.String())

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
