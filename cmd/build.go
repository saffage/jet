package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/saffage/jet/cgen"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/urfave/cli/v2"
)

func Build(file *text.File) error {
	report.Debug("set file '%s' as main module", file.Path)

	if err := build(file); err != nil {
		return err
	}

	if err := compileToC(filepath.Dir(file.Path), file.Name); err != nil {
		return err
	}

	if config.Run {
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
	config.Run = ctx.Bool("run")
	config.DumpCheckerState = ctx.Bool("dump-checker-state")
	config.ParseAst = ctx.Bool("parse-ast")
	config.TraceParser = ctx.Bool("trace-parser")

	config.CC = ctx.String("cc")
	config.CCFlags = ctx.String("cc-flags")
	config.LDFlags = ctx.String("ld-flags")

	return nil
}

func actionBuild(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return errors.New("expected path to a file")
	}

	if ctx.Args().Len() != 1 {
		return errors.New("invalid arguments count (expected 1)")
	}

	argument := ctx.Args().Get(0)
	file, err := config.ReadFile(argument)

	if err != nil {
		return err
	}

	return Build(file)
}

func build(file *text.File) error {
	if err := checker.CheckBuiltInPackage(); err != nil {
		return err
	}

	m, err := checker.CheckFile(file)
	if err != nil {
		return err
	}

	fileDir := filepath.Dir(file.Path)
	dir := filepath.Join(fileDir, config.CacheDirName)

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

func compileToC(dir, name string) error {
	file := filepath.Join(dir, config.CacheDirName, name+".c")
	args := []string{"-o", name, file}

	if len(config.LDFlags) > 0 {
		args = append(args, strings.Split(config.LDFlags, " ")...)
	}

	cmd := exec.Command(config.CC, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	report.HintX("cc", "%s", cmd.String())

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
