package jet

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
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
			exe := "./" + name + exe
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
