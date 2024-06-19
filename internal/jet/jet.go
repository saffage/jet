package jet

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/saffage/jet/cgen"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/token"
)

func process(cfg *config.Config, fileID config.FileID) bool {
	checker.CheckBuiltInPkgs()

	m, err := checker.CheckFile(cfg, fileID)
	if err != nil {
		report.Errors(err)
		return false
	}

	if config.FlagGenC {
		finfo := cfg.Files[fileID]
		dir := filepath.Dir(finfo.Path)

		err := os.Mkdir(filepath.Join(dir, ".jet"), os.ModePerm)
		if err != nil && !os.IsExist(err) {
			report.Errors(err)
			return false
		}

		dir = filepath.Join(dir, ".jet")

		for _, mImported := range m.Imports {
			genCFile(mImported, dir)
		}

		return genCFile(m, dir)
	}

	return true
}

func genCFile(m *checker.Module, dir string) bool {
	filename := filepath.Join(dir, m.Name()+".c")
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	report.Hintf("generating module '%s'", m.Name())
	report.TaggedDebugf("gen", "module file is '%s'", filename)

	if err := cgen.Generate(f, m); err != nil {
		report.Errors(err)
		return false
	}

	return true
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
		err = fmt.Errorf("expected file extension '.jet', got '%s'", fileExt)
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
	file := filepath.Join(dir, ".jet", module+".c")
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