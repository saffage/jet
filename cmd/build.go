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

func Build(cfg *config.Config) error {
	name := cfg.Files[config.MainFileID].Name
	path := cfg.Files[config.MainFileID].Path

	report.Debugf("set file '%s' as main module", path)

	if err := internalBuild(cfg, config.MainFileID); err != nil {
		return err
	}

	if err := compileToC(cfg, filepath.Dir(path), name); err != nil {
		return err
	}

	if cfg.Flags.Run {
		exePath := "." + string(filepath.Separator) + name

		if runtime.GOOS == "windows" {
			exePath += ".exe"
		}

		report.Hintf("running: '%s'", exePath)

		cmd := exec.Command(exePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			wd, _ := os.Getwd()
			report.Hint(wd)
			report.TaggedError("run", err.Error())
		}
	}

	return nil
}

func beforeBuild(ctx *cli.Context) error {
	config.Global.Flags.Run = ctx.Bool("run")
	config.Global.Flags.DumpCheckerState = ctx.Bool("dump-checker-state")
	config.Global.Flags.ParseAst = ctx.Bool("parse-ast")
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

	config.Global.Files[config.MainFileID] = config.FileInfo{
		Name: name,
		Path: path,
		Buf:  bytes.NewBuffer(data),
	}

	return Build(config.Global)
}

func internalBuild(cfg *config.Config, fileID config.FileID) error {
	if err := checker.CheckBuiltInPkgs(cfg); err != nil {
		return err
	}

	m, err := checker.CheckFile(cfg, fileID)
	if err != nil {
		return err
	}

	finfo := cfg.Files[fileID]
	dir := filepath.Join(filepath.Dir(finfo.Path), cfg.Options.CacheDir)
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

	report.Hintf("generating module '%s'", m.Name())
	report.TaggedDebugf("gen", "module file is '%s'", filename)

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
	report.TaggedHint("cc", cmd.String())

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
