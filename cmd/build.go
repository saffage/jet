package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/urfave/cli/v2"
)

func Build(file *text.File) error {
	if buildError := Check(file); buildError != nil {
		return buildError
	}

	// filename, ccError := cc(filepath.Dir(file.Path), file.Name)
	//
	// if ccError != nil {
	// 	return ccError
	// }
	//
	// if config.Run {
	// 	run(filename)
	// }

	return nil
}

func beforeBuild(ctx *cli.Context) error {
	config.CC = ctx.String("cc")
	config.CCFlags = ctx.String("cc-flags")
	config.LDFlags = ctx.String("ld-flags")

	return nil
}

func actionBuild(ctx *cli.Context) error {
	file, err := fileArgument(ctx)

	if err != nil {
		return err
	}

	return Build(file)
}

func run(filename string) {
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	report.HintX("run", "running: '%s'", filename)

	cmd := exec.Command(filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		wd, _ := os.Getwd()
		report.HintX("run", "%s", wd)
		report.ErrorX("run", "%s", err.Error())
	}
}

func cc(dir, name string) (filename string, err error) {
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
		return "", err
	}

	filename = filepath.Join(".", name)

	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	return filename, nil
}
