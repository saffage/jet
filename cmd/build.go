package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
	"github.com/urfave/cli/v2"
)

func Build(cfg *config.Config) error {
	name := cfg.Files[config.MainFileID].Name
	path := cfg.Files[config.MainFileID].Path

	report.Debug("set file '%s' as main module", path)

	if err := buildFile(cfg, config.MainFileID); err != nil {
		return err
	}

	if err := compileC(cfg, filepath.Dir(path), name); err != nil {
		return err
	}

	if cfg.Flags.Run {
		exePath := "." + string(filepath.Separator) + name

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
			report.Hint("%s", wd)
			report.ErrorX("run", "%s", err.Error())
		}
	}

	return nil
}

func beforeBuild(cfg *config.Config) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		cfg.Flags.Run = ctx.Bool("run")
		cfg.Options.CC = ctx.String("cc")
		cfg.Options.CCFlags = ctx.String("cc-flags")
		cfg.Options.LDFlags = ctx.String("ld-flags")
		return nil
	}
}

func actionBuild(cfg *config.Config) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		err := readFileToConfig(ctx, cfg, config.MainFileID)
		if err != nil {
			return err
		}
		return Build(cfg)
	}
}

func buildFile(cfg *config.Config, fileID config.FileID) error {
	types.InitModuleCore(cfg)

	m, err := types.CheckFile(cfg, fileID)
	if err != nil {
		return err
	}

	fileInfo := cfg.Files[fileID]
	dir := filepath.Join(filepath.Dir(fileInfo.Path), cfg.Options.CacheDir)
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

func genModule(m *types.Module, dir string) error {
	filename := filepath.Join(dir, m.Name()+".c")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	report.Hint("generating module '%s'", m.Name())
	report.DebugX("gen", "module file is '%s'", filename)
	// return cgen.Generate(f, m)
	return nil
}

func compileC(cfg *config.Config, dir, moduleName string) error {
	file := filepath.Join(dir, cfg.Options.CacheDir, moduleName+".c")
	args := []string{"-o", moduleName, file}

	if len(cfg.Options.LDFlags) > 0 {
		args = append(args, strings.Split(cfg.Options.LDFlags, " ")...)
	}

	cmd := exec.Command(cfg.Options.CC, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	report.HintX("cc", cmd.String())
	return cmd.Run()
}
