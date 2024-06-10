package jet

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/token"
)

func ProcessArgs(args []string) {
	args, action := config.ParseArgs(args)
	report.IsDebug = config.FlagDebug

	switch action {
	case config.ActionShowHelp:
		// Nothing to do

	case config.ActionDefault:
		if len(args) == 0 {
			report.Errorf("expected filename")
			return
		} else if len(args) > 1 {
			report.Errorf("too many arguments (expected 1)")
			return
		}

		processFile(filepath.Clean(args[0]))

	case config.ActionCompileToC:
		path := filepath.Clean(args[0])
		name, data, err := readFile(path)
		if err != nil {
			report.Errors(err)
		}

		config.Global.Files[config.MainFileID] = config.FileInfo{
			Name: name,
			Path: path,
			Buf:  bytes.NewBuffer(data),
		}

		report.Debugf("set file '%s' as main module", path)
		if !process(config.Global, config.MainFileID) {
			return
		}

		if config.CFlagBuild || config.CFlagRun {
			if err := compileToC(config.CFlagCC, name, filepath.Dir(path)); err != nil {
				report.TaggedError("cc", err.Error())
				return
			}
		}

		if config.CFlagRun {
			exe := "./" + name + ".exe"
			report.TaggedHintf("run", "running '%s'", exe)
			cmd := exec.Command(exe)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				report.TaggedError("run", err.Error())
			}
		}

	default:
		panic("not implemented")
	}
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

	if fileExt != ".jet" && fileExt != ".jetx" {
		err = fmt.Errorf("expected file extension '.jet' or '.jetx', got '%s'", fileExt)
		return
	}

	name = filepath.Base(path[:len(path)-len(fileExt)])
	if _, err = token.IsValidIdent(name); err != nil {
		err = errors.Join(fmt.Errorf("invalid module name (file name must be a valid Jet identifier)"), err)
		return
	}

	data, err = os.ReadFile(path)
	if err != nil {
		err = errors.Join(fmt.Errorf("while reading file '%s'", path), err)
		return
	}

	return
}

func processFile(path string) {
	name, data, err := readFile(path)
	if err != nil {
		report.Errors(err)
	}

	config.Global.Files[config.MainFileID] = config.FileInfo{
		Name: name,
		Path: path,
		Buf:  bytes.NewBuffer(data),
	}

	report.Debugf("set file '%s' as main module", path)
	process(config.Global, config.MainFileID)
}

func compileToC(cc, module, dir string) error {
	file := filepath.Join(dir, ".jet", module+"__jet.c")

	args := []string{"-o", module, file}
	if len(config.CLDflags) > 0 {
		args = append(args, strings.Split(config.CLDflags, " ")...)
	}
	cmd := exec.Command(cc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	report.TaggedHint("cc", cmd.String())

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
