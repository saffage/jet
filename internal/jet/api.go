package jet

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/token"
)

func ProcessArgs(args []string) {
	config.ParseArgs(args)
	report.IsDebug = config.FlagDebug

	if len(config.Args) != 1 {
		report.Errorf("expected filename")
		return
	}

	path := filepath.Clean(config.Args[0])
	stat, err := os.Stat(path)
	if err != nil {
		report.Errorf(err.Error())
		return
	}

	if !stat.Mode().IsRegular() {
		report.Errorf("'%s' is not a file", path)
		return
	}

	fileExt := filepath.Ext(path)

	if fileExt != ".jet" {
		report.Errorf("expected file extension '.jet', not '%s'", fileExt)
		return
	}

	name := filepath.Base(path[:len(path)-len(fileExt)])
	if _, err := token.IsValidIdent(name); err != nil {
		err = errors.Join(fmt.Errorf("invalid module name (file name must be a valid Jet identifier)"), err)
		report.Errorf(err.Error())
		return
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		report.TaggedErrorf("internal", "while reading file '%s': %s", path, err.Error())
		panic(err)
	}

	config.Global.Files[config.MainFileID] = config.FileInfo{
		Name: name,
		Path: path,
		Buf:  bytes.NewBuffer(buf),
	}
	report.Debugf("set file '%s' as main module", path)
	process(config.Global, config.MainFileID)
}
